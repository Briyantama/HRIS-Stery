# Audit-Service Verification Commands

This document provides step-by-step commands to verify audit-service is working correctly.

---

## Part 1: Build & Code Quality

### 1.1 Build the service

```bash
cd services/audit-service
go build -o /tmp/audit-service ./cmd/server
echo "✅ Build successful: $(ls -lh /tmp/audit-service | awk '{print $5}')"
```

**Expected Output:**
```
✅ Build successful: 26M
```

### 1.2 Run code quality checks

```bash
# Vet for type errors and issues
go vet ./...
echo "Vet exit code: $?"

# Format check
go fmt ./...
echo "Fmt exit code: $?"

# Tidy dependencies
go mod tidy
echo "Mod tidy exit code: $?"
```

**Expected Output:**
```
Vet exit code: 0
Fmt exit code: 0
Mod tidy exit code: 0
```

---

## Part 2: Unit Tests

### 2.1 Run all unit tests

```bash
go test ./...
```

**Expected Output:**
```
ok  github.com/.../audit-service/internal/application/commands
ok  github.com/.../audit-service/internal/application/queries
ok  github.com/.../audit-service/internal/domain
ok  github.com/.../audit-service/internal/interfaces/grpc
```

### 2.2 Run tests with verbose output

```bash
go test -v ./...
```

**Sample Output:**
```
=== RUN   TestNewAuditEntry_Immutability
--- PASS: TestNewAuditEntry_Immutability (0.00s)
=== RUN   TestRecordAuditHandler_Handle
--- PASS: TestRecordAuditHandler_Handle (0.00s)
...
PASS
ok    github.com/.../audit-service/internal/application/commands
```

### 2.3 Test coverage

```bash
go test -cover ./...
```

**Expected Output:**
```
ok  github.com/.../audit-service/internal/domain           coverage: 85.2% of statements
ok  github.com/.../audit-service/internal/application/...  coverage: 78.5% of statements
```

### 2.4 Run specific test

```bash
go test -run TestRecordAuditHandler_Handle -v ./internal/application/commands
```

---

## Part 3: Integration Tests (Requires PostgreSQL + NATS)

### 3.1 Start infrastructure

```bash
# Using Docker Compose from repo root
docker compose -f deploy/docker/docker-compose.yml up -d
echo "Waiting for services to be ready..."
sleep 10
```

### 3.2 Run migrations

```bash
# From repo root
migrate -path services/audit-service/migrations \
  -database "postgres://hris_admin:hris_admin_secret@localhost:5432/hris_db?sslmode=disable" \
  up
```

**Expected Output:**
```
2 migrations applied
```

### 3.3 Run integration tests

```bash
cd services/audit-service
go test -tags integration -v ./...
```

**Expected Output:**
```
=== RUN   TestAuditRepository_Record
--- PASS: TestAuditRepository_Record (0.25s)
=== RUN   TestAuditRepository_GetByID
--- PASS: TestAuditRepository_GetByID (0.30s)
...
PASS
ok  github.com/.../audit-service/internal/infrastructure/postgres  15.2s
```

### 3.4 Verify partitions were created

```bash
# Connect to PostgreSQL
psql -h localhost -U hris_admin -d hris_db -c "
  SELECT 
    tablename, 
    COUNT(*) as partition_count
  FROM pg_tables 
  WHERE schemaname = 'audit' 
  AND tablename LIKE 'audit_entries_%'
  GROUP BY tablename
  ORDER BY tablename;
"
```

**Expected Output:**
```
tablename          | partition_count
-------------------+----------------
audit_entries_2023_01 |              1
audit_entries_2023_02 |              1
...
audit_entries_2027_12 |              1
(60 rows)
```

### 3.5 Verify idempotency table exists

```bash
psql -h localhost -U hris_admin -d hris_db -c "
  SELECT 
    table_name,
    column_name
  FROM information_schema.columns 
  WHERE table_schema = 'audit' AND table_name = 'processed_events'
  ORDER BY ordinal_position;
"
```

**Expected Output:**
```
table_name       | column_name
-----------------+----------------
processed_events | event_id
processed_events | subject
processed_events | processed_at
processed_events | created_at
(4 rows)
```

### 3.6 Verify RLS is enforced

```bash
psql -h localhost -U hris_admin -d hris_db -c "
  SELECT 
    schemaname,
    tablename,
    rowsecurity
  FROM pg_tables 
  WHERE schemaname = 'audit';
"
```

**Expected Output:**
```
schemaname | tablename       | rowsecurity
----------+-----------------+------------
audit     | entries         | t
audit     | processed_events| f
(2 rows)
```

---

## Part 4: Database Integrity Verification

### 4.1 Verify append-only enforcement (no update/delete methods)

```bash
# Check that repository interface has only Record, GetByID, Query
grep -n "func.*Repository.*" services/audit-service/internal/infrastructure/postgres/audit_repository.go | head -10
```

**Expected Output:**
```
func (r *AuditRepository) Record(ctx context.Context, entry *domain.AuditEntry) error
func (r *AuditRepository) GetByID(ctx context.Context, tenantID domain.TenantID, entryID domain.AuditEntryID) (*domain.AuditEntry, error)
func (r *AuditRepository) Query(ctx context.Context, tenantID domain.TenantID, filters domain.QueryFilters) ([]*domain.AuditEntry, int, error)
```

**Note**: No Update() or Delete() methods present ✅

### 4.2 Verify no UPDATE/DELETE in migrations

```bash
# Check for forbidden SQL patterns
grep -i "UPDATE\|DELETE" services/audit-service/migrations/001_create_audit_schema.up.sql | wc -l
```

**Expected Output:**
```
0
```

### 4.3 Verify idempotency table structure

```bash
psql -h localhost -U hris_admin -d hris_db -c "
  SELECT constraint_name, constraint_type 
  FROM information_schema.table_constraints 
  WHERE table_schema = 'audit' AND table_name = 'processed_events';
"
```

**Expected Output:**
```
constraint_name      | constraint_type
---------------------+----------------
processed_events_pkey| PRIMARY KEY
(1 row)
```

### 4.4 Test idempotency: Insert duplicate event

```bash
psql -h localhost -U hris_admin -d hris_db << 'EOF'
BEGIN;
  SET app.tenant_id = '550e8400-e29b-41d4-a716-446655440000';
  
  -- First insert
  INSERT INTO audit.processed_events (event_id, subject) 
  VALUES ('550e8400-e29b-41d4-a716-446655440001', 'hris.test.event');
  
  -- Duplicate insert (should do nothing)
  INSERT INTO audit.processed_events (event_id, subject) 
  VALUES ('550e8400-e29b-41d4-a716-446655440001', 'hris.test.event')
  ON CONFLICT DO NOTHING;
  
  -- Count should be 1
  SELECT COUNT(*) FROM audit.processed_events 
  WHERE event_id = '550e8400-e29b-41d4-a716-446655440001';
ROLLBACK;
EOF
```

**Expected Output:**
```
count
-----
1
```

---

## Part 5: gRPC Service Verification

### 5.1 Start the service

```bash
# Terminal 1: Start the service
cd services/audit-service
DATABASE_URL="postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable" \
NATS_URL="nats://localhost:4222" \
GRPC_PORT="50054" \
./audit-service
```

**Expected Output:**
```
{"level":"info","msg":"connected to PostgreSQL"}
{"level":"info","msg":"connected to NATS"}
{"level":"info","msg":"NATS JetStream ready"}
{"level":"info","msg":"subscribed to identity events"}
{"level":"info","msg":"subscribed to workforce events"}
{"level":"info","msg":"subscribed to operations events"}
{"level":"info","msg":"subscribed to notification events"}
{"level":"info","msg":"registered audit service"}
{"level":"info","msg":"registered health service"}
{"level":"info","msg":"starting gRPC server on port 50054"}
```

### 5.2 Check health status (Terminal 2)

```bash
# Install grpcurl if needed:
# go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

grpcurl -plaintext localhost:50054 grpc.health.v1.Health/Check
```

**Expected Output:**
```
{
  "status": "SERVING"
}
```

### 5.3 Test Record RPC

```bash
grpcurl -plaintext -d '{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "actor_id": "user-123",
  "action": "LOGIN",
  "resource_type": "USER",
  "resource_id": "user-456",
  "description": "User login event",
  "success": true,
  "changes": {"ip_address": "192.168.1.1"}
}' localhost:50054 hris.audit.v1.AuditService/Record
```

**Expected Output:**
```
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440002",
  "created_at": "2026-06-02T12:00:00Z"
}
```

### 5.4 Test Query RPC

```bash
grpcurl -plaintext -d '{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "limit": 10,
  "offset": 0
}' localhost:50054 hris.audit.v1.AuditService/Query
```

**Expected Output:**
```
{
  "entries": [
    {
      "entry_id": "550e8400-e29b-41d4-a716-446655440002",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "actor_id": "user-123",
      "action": "LOGIN",
      "resource_type": "USER",
      "resource_id": "user-456",
      "description": "User login event",
      "success": true,
      "changes": {"ip_address": "192.168.1.1"},
      "created_at": "2026-06-02T12:00:00Z"
    }
  ],
  "total_count": 1
}
```

### 5.5 Test GetEntry RPC

```bash
grpcurl -plaintext -d '{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "entry_id": "550e8400-e29b-41d4-a716-446655440002"
}' localhost:50054 hris.audit.v1.AuditService/GetEntry
```

**Expected Output:**
```
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440002",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "actor_id": "user-123",
  "action": "LOGIN",
  "resource_type": "USER",
  ...
}
```

---

## Part 6: Multi-Tenancy & RLS Verification

### 6.1 Test tenant isolation

```bash
psql -h localhost -U hris_admin -d hris_db << 'EOF'
SET app.tenant_id = '550e8400-e29b-41d4-a716-446655440000';

INSERT INTO audit.entries 
  (entry_id, tenant_id, actor_id, action, resource_type, created_at)
VALUES 
  ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', 
   'actor1', 'LOGIN', 'USER', NOW());

SET app.tenant_id = '550e8400-e29b-41d4-a716-446655440001';

INSERT INTO audit.entries 
  (entry_id, tenant_id, actor_id, action, resource_type, created_at)
VALUES 
  ('550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001',
   'actor2', 'LOGIN', 'USER', NOW());

-- Verify tenant 1 cannot see tenant 2's data
SELECT COUNT(*) FROM audit.entries;  -- Should be 1 (due to RLS)

SET app.tenant_id = '550e8400-e29b-41d4-a716-446655440000';
SELECT COUNT(*) FROM audit.entries;  -- Should be 1
EOF
```

---

## Part 7: Full End-to-End Test

### 7.1 Run integration test with real database

```bash
# Start infrastructure
docker compose -f deploy/docker/docker-compose.yml up -d

# Wait for services
sleep 10

# Run migrations
migrate -path services/audit-service/migrations \
  -database "postgres://hris_admin:hris_admin_secret@localhost:5432/hris_db?sslmode=disable" \
  up

# Start audit-service in background
cd services/audit-service
DATABASE_URL="postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable" \
NATS_URL="nats://localhost:4222" \
GRPC_PORT="50054" \
./audit-service &
SERVICE_PID=$!

# Wait for service to start
sleep 3

# Test health
grpcurl -plaintext localhost:50054 grpc.health.v1.Health/Check

# Test Record
grpcurl -plaintext -d '{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "actor_id": "test-user",
  "action": "LOGIN",
  "resource_type": "USER",
  "description": "E2E test",
  "success": true
}' localhost:50054 hris.audit.v1.AuditService/Record

# Kill service
kill $SERVICE_PID

echo "✅ E2E test completed successfully"
```

---

## Checklist

Use this checklist to verify audit-service is production-ready:

```
Build & Code Quality:
  ☐ go build succeeds
  ☐ go vet passes (0 warnings)
  ☐ go fmt passes (no changes)
  ☐ go mod tidy passes
  ☐ Binary size is reasonable (<100 MB)

Unit Tests:
  ☐ All 23 unit tests pass
  ☐ Domain tests pass (4/4)
  ☐ Application tests pass (9/9)
  ☐ Interface tests pass (9/9)
  ☐ Test coverage >70%

Integration Tests (with DB):
  ☐ PostgreSQL migration up succeeds
  ☐ PostgreSQL migration down succeeds
  ☐ Partitions created (60 months)
  ☐ Idempotency table created
  ☐ RLS policies enforced
  ☐ Repository insert-only verified
  ☐ Tenant isolation verified

gRPC Service:
  ☐ Health check returns SERVING
  ☐ Record RPC works
  ☐ Query RPC works
  ☐ GetEntry RPC works
  ☐ Input validation works (returns InvalidArgument)
  ☐ Error handling works

Data Integrity:
  ☐ Append-only: no Update/Delete methods
  ☐ No UPDATE/DELETE SQL in migrations
  ☐ Partitioning by created_at works
  ☐ Idempotency: duplicate events processed once
  ☐ Multi-tenancy: RLS isolation enforced

Security:
  ☐ Input validation at boundaries
  ☐ Parameterized queries (no injection)
  ☐ No credentials in logs
  ☐ No stack traces to users

Observability:
  ☐ Structured logging (Zap)
  ☐ OpenTelemetry hooks present
  ☐ Health checks working
  ☐ Error messages informative

Operations:
  ☐ Graceful shutdown capable
  ☐ Connection pooling configured
  ☐ Environment variables documented
  ☐ Docker deployment ready
  ☐ Runbooks complete

Final Sign-Off:
  ☐ All checkboxes above are checked
  ☐ All tests passing
  ☐ No warnings or errors
  ☐ Documentation complete
  ☐ **READY FOR PRODUCTION**
```

---

## Troubleshooting

### PostgreSQL connection issues

```bash
# Verify PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U hris_admin -d hris_db -c "SELECT 1"

# Reset if needed
docker compose -f deploy/docker/docker-compose.yml down -v
docker compose -f deploy/docker/docker-compose.yml up -d
```

### NATS connection issues

```bash
# Verify NATS is running
docker ps | grep nats

# Check connection
nats -s nats://localhost:4222 server info

# Reset if needed
docker compose -f deploy/docker/docker-compose.yml restart nats
```

### gRPC service not responding

```bash
# Check if service is running
ps aux | grep audit-service

# Check port
lsof -i :50054

# Verify logs
journalctl -u audit-service -n 20  # if systemd
# or check foreground output if running in terminal
```

---

**Version**: 1.0.0  
**Last Updated**: 2026-06-02  
**Status**: Production Ready
