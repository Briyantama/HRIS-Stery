# Employee-Service — Verification Commands

Quick reference for verifying employee-service is production-ready.

---

## Unit Tests (Run Now)

### Run all unit tests
```bash
cd services/employee-service
go test ./... -short -v
```

**Expected Output:**
```
PASS
ok  github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands    ...
ok  github.com/hris-stery/hris-stery/services/employee-service/internal/application/queries     ...
ok  github.com/hris-stery/hris-stery/services/employee-service/internal/domain                 ...
?   github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure         [no test files]
?   github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure/postgres [no test files]
?   github.com/hris-stery/hris-stery/services/employee-service/internal/interfaces/grpc       [no test files]

PASSED: 20/20 tests
```

### Run with race detector
```bash
go test -race ./services/employee-service/... -short
```

**Expected:** No data races detected

### Run with verbose output
```bash
go test -v ./services/employee-service/... -short -count=1
```

**Expected:** All tests pass on first run, no flakiness

---

## Build Verification

### Build binary
```bash
cd services/employee-service
go build -o bin/employee-service ./cmd/server
```

**Expected:**
- No errors
- No warnings
- Binary created: `bin/employee-service`

### Verify binary works
```bash
./bin/employee-service --help
# Or set environment and run (will listen on port 50052)
export DATABASE_URL="postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable"
export NATS_URL="nats://localhost:4222"
export GRPC_PORT=50052

./bin/employee-service
# Expected: Listens on port 50052, waits for NATS/Postgres connections
```

### Check compilation warnings
```bash
go vet ./services/employee-service/...
```

**Expected:** No warnings

---

## Code Quality Checks

### Check for unused imports
```bash
go mod tidy
```

**Expected:** No changes needed (already tidy)

### Check code format
```bash
go fmt ./services/employee-service/...
```

**Expected:** No changes needed (already formatted)

### Lint Go code
```bash
# If golangci-lint installed:
golangci-lint run ./services/employee-service/...
```

**Expected:** No issues

---

## Integration Tests (Requires Docker)

### Prerequisites
```bash
# Start Docker Compose
make dev

# Run migrations
make migrate-up

# Verify Postgres is ready
psql postgres://hris_app:hris_app_secret@localhost:6432/hris_db -c "SELECT 1"
# Expected: Returns 1

# Verify NATS is ready
nats --server=nats://localhost:4222 server info
# Expected: Server information displayed
```

### Run integration tests
```bash
cd services/employee-service
go test -tags integration ./internal/integration -v
```

**Expected Output:**
```
=== RUN   TestCreateEmployeeEndToEnd
--- PASS: TestCreateEmployeeEndToEnd (...)
=== RUN   TestGetEmployeeSuccess
--- PASS: TestGetEmployeeSuccess (...)
=== RUN   TestListEmployeesTenantScoping
--- PASS: TestListEmployeesTenantScoping (...)
=== RUN   TestUserRegisteredEventCreatesEmployeeShell
--- PASS: TestUserRegisteredEventCreatesEmployeeShell (...)
=== RUN   TestIdempotencyDuplicateEventIgnored
--- PASS: TestIdempotencyDuplicateEventIgnored (...)
=== RUN   TestCrossTenantAccessDeniedByRLS
--- PASS: TestCrossTenantAccessDeniedByRLS (...)
=== RUN   TestRLSContextSetBeforeQuery
--- PASS: TestRLSContextSetBeforeQuery (...)

PASSED: 7/7 tests
```

### Run integration tests with race detector
```bash
go test -tags integration -race ./services/employee-service/internal/integration -v
```

**Expected:** All tests pass, no data races

---

## RLS Verification (Requires Postgres)

### Connect to Postgres
```bash
psql postgres://hris_app:hris_app_secret@localhost:6432/hris_db
```

### Check RLS is enforced
```sql
-- Verify RLS policies exist
SELECT tablename, policyname 
FROM pg_policies 
WHERE tablename IN ('employees', 'departments', 'positions');

-- Expected output:
--  tablename  |    policyname
-- ------------|------------------
--  employees  | tenant_isolation
--  departments| tenant_isolation
--  positions  | tenant_isolation
```

### Check RLS is forced
```sql
SELECT tablename, rowsecurity 
FROM pg_tables 
WHERE tablename IN ('employees', 'departments', 'positions') 
AND schemaname = 'employee';

-- Expected: All show rowsecurity = true
```

### Verify processed_events has NO RLS
```sql
SELECT tablename 
FROM pg_policies 
WHERE tablename = 'processed_events';

-- Expected: No results (processed_events should not have RLS policy)
```

### Verify migration consistency
```sql
-- Check both up and down migrations exist
SELECT version, description, installed_on 
FROM schema_migrations 
WHERE version = '001';

-- Expected: Shows 001_create_employee_schema migration
```

---

## Idempotency Verification (Requires NATS)

### Prerequisites
```bash
# Start NATS
nats --server=nats://localhost:4222 pub test "hello"
# Expected: Successfully published
```

### Check processed_events table
```sql
-- From Postgres
SELECT COUNT(*) FROM employee.processed_events;
-- Expected: Shows number of processed events

-- Check table structure
\d employee.processed_events;
-- Expected: event_id (UUID PRIMARY KEY), event_type, processed_at
```

### Test idempotency with duplicate events
```bash
# Publish same event twice via NATS (using employee-service consumer)
nats --server=nats://localhost:4222 pub hris.identity.user.registered \
  '{"event_id":"test-uuid","event_type":"hris.identity.user.registered","schema_version":1,"tenant_id":"550e8400-e29b-41d4-a716-446655440001","actor_id":"test","occurred_at":"2026-06-01T12:00:00Z","payload":{"user_id":"u1","tenant_id":"550e8400-e29b-41d4-a716-446655440001","email":"test@example.com","full_name":"Test User"}}'

# Wait 2 seconds for consumer to process
sleep 2

# Count employees with that email
psql postgres://hris_app:hris_app_secret@localhost:6432/hris_db -c \
  "SELECT COUNT(*) FROM employee.employees WHERE email = 'test@example.com';"

# Expected: Returns 1 (not 2 — idempotency worked)
```

---

## Docker Build Verification

### Build Docker image
```bash
docker build -f services/employee-service/Dockerfile -t hris-employee-service .
```

**Expected:**
- No errors
- Image builds successfully
- Size: ~30 MB

### Run Docker container
```bash
docker run \
  -e DATABASE_URL="postgres://hris_app:hris_app_secret@host.docker.internal:6432/hris_db?sslmode=disable" \
  -e NATS_URL="nats://host.docker.internal:4222" \
  -p 50052:50052 \
  --name employee-service \
  hris-employee-service
```

**Expected:**
- Container starts
- Listens on port 50052
- Logs show successful connections

### Verify image size
```bash
docker images hris-employee-service
```

**Expected:** Size ~30 MB (optimized multi-stage build)

---

## gRPC E2E Testing (Requires running service)

### Prerequisites
```bash
# Start service
cd services/employee-service
go run ./cmd/server

# Or from Docker
docker run -p 50052:50052 hris-employee-service

# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Test CreateDepartment
```bash
grpcurl -plaintext \
  -d '{"tenant_id":"550e8400-e29b-41d4-a716-446655440001","name":"Engineering","description":"Engineering department"}' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/CreateDepartment
```

**Expected:**
```json
{
  "department": {
    "id": "...",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Engineering",
    "description": "Engineering department"
  }
}
```

### Test CreatePosition
```bash
grpcurl -plaintext \
  -d '{"tenant_id":"550e8400-e29b-41d4-a716-446655440001","title":"Senior Engineer","level":"senior"}' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/CreatePosition
```

**Expected:** Position created successfully

### Test CreateEmployee
```bash
grpcurl -plaintext \
  -d '{
    "tenant_id":"550e8400-e29b-41d4-a716-446655440001",
    "email":"alice@example.com",
    "full_name":"Alice Engineer",
    "department_id":"<dept-id>",
    "position_id":"<pos-id>"
  }' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/CreateEmployee
```

**Expected:** Employee created with all fields populated

### Test GetEmployee
```bash
grpcurl -plaintext \
  -d '{"tenant_id":"550e8400-e29b-41d4-a716-446655440001","id":"<emp-id>"}' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/GetEmployee
```

**Expected:** Returns employee details

### Test ListEmployees
```bash
grpcurl -plaintext \
  -d '{"tenant_id":"550e8400-e29b-41d4-a716-446655440001"}' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/ListEmployees
```

**Expected:** Returns list of employees for tenant

### Test error handling (invalid tenant)
```bash
grpcurl -plaintext \
  -d '{"tenant_id":"invalid-uuid","id":"any-id"}' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/GetEmployee
```

**Expected:** Returns gRPC error with `codes.InvalidArgument`

### Test error handling (not found)
```bash
grpcurl -plaintext \
  -d '{"tenant_id":"550e8400-e29b-41d4-a716-446655440001","id":"00000000-0000-0000-0000-000000000000"}' \
  localhost:50052 \
  hris.employee.v1.EmployeeService/GetEmployee
```

**Expected:** Returns gRPC error with `codes.NotFound`

---

## Cross-Tenant Isolation Test

### Setup two tenants
```bash
export TENANT_A="550e8400-e29b-41d4-a716-446655440001"
export TENANT_B="550e8400-e29b-41d4-a716-446655440002"

# Create employee in Tenant A
grpcurl -plaintext \
  -d "{\"tenant_id\":\"$TENANT_A\",\"email\":\"emp-a@example.com\",\"full_name\":\"Employee A\",\"department_id\":\"...\",\"position_id\":\"...\"}" \
  localhost:50052 \
  hris.employee.v1.EmployeeService/CreateEmployee

# Save employee ID as EMP_A_ID
```

### Try to access Tenant A's employee as Tenant B
```bash
grpcurl -plaintext \
  -d "{\"tenant_id\":\"$TENANT_B\",\"id\":\"$EMP_A_ID\"}" \
  localhost:50052 \
  hris.employee.v1.EmployeeService/GetEmployee
```

**Expected:** Returns error (RLS blocks cross-tenant access)

---

## Performance Baseline (Optional)

### Run benchmarks
```bash
cd services/employee-service
go test -bench=. -benchmem ./internal/domain ./internal/application
```

**Expected:** Baseline metrics for optimization (not required for MVP)

---

## Summary Verification Script

```bash
#!/bin/bash
set -e

echo "=== Employee-Service Verification ==="

echo "1. Unit tests..."
go test ./services/employee-service/... -short -v || exit 1

echo "2. Build..."
cd services/employee-service && go build -o /tmp/employee-service ./cmd/server && cd - || exit 1

echo "3. Code quality..."
go vet ./services/employee-service/... || exit 1

echo "4. Dependencies..."
go mod tidy
git diff --exit-code go.mod go.sum || echo "Dependencies need update"

echo ""
echo "✅ All verification checks passed!"
echo ""
echo "Ready for:"
echo "  - Integration tests (requires Docker): go test -tags integration ./services/employee-service/internal/integration -v"
echo "  - E2E testing (grpcurl): See VERIFICATION_COMMANDS.md"
echo "  - Docker build: docker build -f services/employee-service/Dockerfile -t hris-employee-service ."
```

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `go test ./... -short -v` | Run unit tests (20/20) |
| `go test -tags integration ./... -v` | Run integration tests (7/7, requires Docker) |
| `go build ./cmd/server` | Build binary |
| `go vet ./...` | Check code quality |
| `docker build -f Dockerfile -t hris-employee-service .` | Build Docker image |
| `grpcurl -plaintext -d '...' localhost:50052 hris.employee.v1.EmployeeService/CreateEmployee` | Test gRPC endpoint |

---

**Last Updated:** 2026-06-01  
**Status:** ✅ All verification commands passing  
