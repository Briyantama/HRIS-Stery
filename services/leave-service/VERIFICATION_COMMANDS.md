# Leave-Service Verification Commands

Quick reference for verifying the implementation.

## Build & Compilation

### Build Server Binary
```bash
cd services/leave-service
go build -v ./cmd/server/
ls -lh server  # Should show ~17MB binary
```

### Check Dependencies
```bash
cd services/leave-service
go mod tidy
go mod verify
```

### Verify Go Format
```bash
cd services/leave-service
go vet ./...
go fmt ./...
```

## Testing

### Run All Tests
```bash
cd services/leave-service
go test -v ./internal/...
```

### Run Domain Tests
```bash
cd services/leave-service
go test -v ./internal/domain/...
```

### Run Application Tests
```bash
cd services/leave-service
go test -v ./internal/application/...
```

### Run with Coverage
```bash
cd services/leave-service
go test -v -cover ./internal/domain/... ./internal/application/...
```

## Architecture Verification

### Check File Structure
```bash
cd services/leave-service
tree -L 2 -I "vendor" --charset ascii
# Expected:
# ├── cmd/
# │   └── server/
# │       └── main.go
# ├── internal/
# │   ├── domain/
# │   ├── application/
# │   ├── infrastructure/
# │   └── interfaces/
# ├── migrations/
# └── go.mod
```

### Verify Proto Alignment
```bash
# Check proto file exists
file proto/hris/leave/v1/leave.proto

# Check generated code exists
file gen/go/hris/leave/v1/leave.pb.go
file gen/go/hris/leave/v1/leave_grpc.pb.go

# Count RPC methods in generated code (should be 8)
grep "func.*Server).*(" gen/go/hris/leave/v1/leave_grpc.pb.go | wc -l
```

## Code Quality Checks

### Count Domain Tests
```bash
cd services/leave-service
go test -v ./internal/domain/... -run Test 2>&1 | grep "^--- PASS" | wc -l
# Expected: 10
```

### Count Application Tests
```bash
cd services/leave-service
go test -v ./internal/application/... -run Test 2>&1 | grep "^--- PASS" | wc -l
# Expected: 3
```

### Check for Forbidden Patterns

#### No interface{} in domain
```bash
grep -r "interface{}" services/leave-service/internal/domain/
# Should return empty (no matches)
```

#### No raw SQL in application
```bash
grep -r "Query\|Exec" services/leave-service/internal/application/
grep -r "SELECT\|INSERT\|UPDATE" services/leave-service/internal/application/
# Should return empty (no raw SQL)
```

#### No business logic in handlers
```bash
grep -A 5 "func.*ApplyLeave" services/leave-service/internal/interfaces/grpc/leave_service.go
# Should show: status.Errorf(codes.Unimplemented, ...)
```

#### RLS enforcement check
```bash
grep -r "WithTenantTx" services/leave-service/internal/infrastructure/postgres/
# Should show in every repository method
```

#### No Phase 2 logic
```bash
grep -r "payroll\|tax\|recruitment\|performance" services/leave-service/internal/
# Should return empty (no matches)
```

## Migration Verification

### Check Migrations Exist
```bash
ls -lh services/leave-service/migrations/
# Should show:
# - 001_create_leave_schema.up.sql
# - 001_create_leave_schema.down.sql
```

### Verify Migration Structure
```bash
grep "CREATE TABLE" services/leave-service/migrations/001_create_leave_schema.up.sql
# Should show: leave_types, leave_requests, leave_balances, processed_events
```

### Check RLS Policies
```bash
grep "ROW LEVEL SECURITY\|CREATE POLICY" services/leave-service/migrations/001_create_leave_schema.up.sql
# Should show RLS enabled and policies for leave_types, leave_requests, leave_balances
```

### Verify Down Migration
```bash
grep "DROP SCHEMA" services/leave-service/migrations/001_create_leave_schema.down.sql
# Should show: DROP SCHEMA IF EXISTS leave CASCADE
```

## Event Verification

### Check Event Envelope Structure
```bash
grep -A 10 "EventEnvelope" services/leave-service/internal/infrastructure/nats/event_publisher.go
# Should show all required fields:
# - event_id
# - event_type
# - schema_version
# - tenant_id
# - actor_id
# - occurred_at
# - payload
```

### Check Idempotency Pattern
```bash
grep -A 5 "processed_events" services/leave-service/internal/infrastructure/nats/employee_created_consumer.go
grep "ON CONFLICT" services/leave-service/migrations/001_create_leave_schema.up.sql
# Should show idempotency implementation
```

## Integration Verification (When Wired)

### Check PostgreSQL Connection
```bash
# When handler wiring is complete:
curl -X GET http://localhost:8080/v1/leave-types
# Should return leave types or 401 if auth required
```

### Check NATS Subscription
```bash
# Connect to NATS
nats sub "hris.operations.leave.>" --raw
# Should show events when requests are created/updated
```

### Test Tenant Isolation
```bash
# Two requests with different tenant_ids should not see each other
# This requires integration tests with real databases
```

## Performance Checks

### Query Performance (indexes)
```bash
grep "CREATE INDEX" services/leave-service/migrations/001_create_leave_schema.up.sql
# Should show indexes on tenant_id for all tables
```

### Binary Size
```bash
ls -lh services/leave-service/server
# Should be reasonable (15-20MB)
```

### Dependencies Size
```bash
cd services/leave-service
go mod graph | wc -l
# Should be reasonable number of dependencies
```

## Deployment Readiness

### Docker Build (when Dockerfile exists)
```bash
docker build -f services/leave-service/Dockerfile \
  -t hris-stery/leave-service:1.0.0 \
  .
docker images | grep leave-service
```

### Environment Variables Check
```bash
grep "os.Getenv" services/leave-service/cmd/server/main.go
# Should show: DATABASE_URL, NATS_URL, GRPC_PORT
```

### Port Configuration
```bash
grep "grpc_port\|GRPC_PORT" services/leave-service/cmd/server/main.go
# Default should be 50054
```

## Final Checklist

- [ ] `go build -v ./cmd/server/` returns success
- [ ] `go test -v ./internal/...` shows 13/13 PASS
- [ ] `grep -r "interface{}" internal/domain/` returns empty
- [ ] `grep -r "payroll\|tax" internal/` returns empty
- [ ] Migrations have both up and down files
- [ ] RLS policies enabled on all tenant-scoped tables
- [ ] processed_events table has NO RLS
- [ ] EventEnvelope has all required fields
- [ ] gRPC handlers return Unimplemented (awaiting wiring)
- [ ] No Phase 2 logic detected

---

**Status:** Foundation Complete  
**Next:** Handler wiring and integration testing
