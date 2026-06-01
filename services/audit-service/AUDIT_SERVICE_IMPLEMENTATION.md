# Audit-Service Implementation Roadmap

**Status:** Layers 1-3 Complete (Domain, Application), Layers 4-6 Ready for Implementation

## COMPLETED

### Layer 1: Proto ✅
- `proto/hris/audit/v1/audit.proto` - Complete with all 3 RPCs

### Layer 2: Domain ✅
- `AuditEntry` aggregate (immutable, append-only)
- `AuditEntryID` value object
- `TenantID` value object
- Repository port interface
- Domain errors
- Query filters

### Layer 3: Application ✅
- `RecordAuditCommand` + handler
- `QueryAuditTrailQuery` + handler
- DTOs for audit entry projection

## IMPLEMENTATION GUIDE - Layers 4-6

### Layer 4: Infrastructure

#### PostgreSQL Migration
```sql
-- migrations/001_create_audit_schema.up.sql
CREATE SCHEMA audit;

CREATE TABLE audit.entries (
  entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  actor_id TEXT NOT NULL,
  action VARCHAR(50) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id UUID,
  description TEXT,
  success BOOLEAN DEFAULT TRUE,
  error_message TEXT,
  changes JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (created_at);

-- RLS disabled: audit-service reads all tenants for compliance
-- INSERT-ONLY enforcement: No UPDATE/DELETE operations allowed

CREATE INDEX idx_audit_tenant_created ON audit.entries (tenant_id, created_at DESC);
CREATE INDEX idx_audit_actor ON audit.entries (actor_id, created_at DESC);
CREATE INDEX idx_audit_resource ON audit.entries (resource_type, resource_id);
```

#### PostgreSQL Repository
File: `internal/infrastructure/postgres/audit_repository.go`

Pattern: Similar to notification-service repository
- Use `shared.WithTenantTx()` for RLS enforcement
- Use sqlc for query generation
- Implement `domain.AuditRepository` interface

#### NATS Consumers
File: `internal/infrastructure/nats/`

Consumers:
- `identity_consumer.go` - Subscribes to `hris.identity.*`
- `workforce_consumer.go` - Subscribes to `hris.workforce.*`
- `operations_consumer.go` - Subscribes to `hris.operations.*`
- `notification_consumer.go` - Subscribes to `hris.notification.*`

Each consumer:
1. Implements idempotency via `processed_events` table
2. Parses event envelope
3. Maps domain event → RecordAuditCommand
4. Calls RecordAuditHandler.Handle()

### Layer 5: Interfaces (gRPC Handlers)

File: `internal/interfaces/grpc/audit_service.go`

Handlers:
```go
func (s *AuditServiceServer) Record(ctx, req) (*RecordAuditResponse, error)
  - Validate tenant_id, actor_id
  - Call RecordAuditHandler
  - Return entry_id + created_at

func (s *AuditServiceServer) GetAuditEntry(ctx, req) (*AuditEntry, error)
  - Call GetAuditEntryHandler
  - Return AuditEntry with all fields

func (s *AuditServiceServer) QueryAuditTrail(ctx, req) (*QueryAuditResponse, error)
  - Extract filters from request
  - Call QueryAuditTrailHandler
  - Map DTOs to proto messages
  - Return with pagination
```

Each handler: < 20 LOC (thin delegation pattern)

### Layer 6: Server (main.go)

File: `cmd/server/main.go`

Setup:
1. PostgreSQL connection + RLS wrapper
2. NATS connection + JetStream
3. Repository instantiation
4. Command handlers instantiation
5. Query handlers instantiation
6. NATS consumer instantiation + subscribe
7. gRPC server registration
8. Health check registration
9. Server startup

## Testing Strategy

### Unit Tests
- Domain: Immutability, invariants
- Application: Filter logic, DTO mapping
- Infrastructure: Repository CRUD (read-only)

### Integration Tests
```
//go:build integration

TestAuditRecording() - Verify entry inserted
TestAuditImmutability() - Verify no update/delete possible
TestAuditQueryFilters() - Test all filter combinations
TestNATSIdempotency() - Duplicate event → single entry
TestPartitionByDate() - Verify partitioning works
TestCrossService() - Auth event → audit entry
```

### Coverage Target
> 90% coverage

## Key Invariants

✅ **Immutable:** No update/delete methods in aggregate
✅ **Append-only:** INSERT ONLY schema enforcement
✅ **Idempotent:** processed_events deduplication
✅ **Tenant-aware:** RLS on audit.entries table
✅ **Event-driven:** NATS consumers map events → audit entries
✅ **Queryable:** Filters by actor, action, resource, date

## Files to Create

Essential:
- [ ] `internal/infrastructure/postgres/audit_repository.go` (250 LOC)
- [ ] `internal/infrastructure/nats/identity_consumer.go` (120 LOC)
- [ ] `internal/infrastructure/nats/workforce_consumer.go` (120 LOC)
- [ ] `internal/infrastructure/nats/operations_consumer.go` (120 LOC)
- [ ] `internal/infrastructure/nats/notification_consumer.go` (100 LOC)
- [ ] `internal/interfaces/grpc/audit_service.go` (150 LOC)
- [ ] `internal/health/health_service.go` (90 LOC)
- [ ] `cmd/server/main.go` (200 LOC)
- [ ] `migrations/001_create_audit_schema.up.sql` (50 LOC)
- [ ] `migrations/001_create_audit_schema.down.sql` (10 LOC)

Tests:
- [ ] `internal/domain/*_test.go` (200 LOC)
- [ ] `internal/application/*_test.go` (150 LOC)
- [ ] `internal/integration_test.go` (400 LOC)

## Expected Metrics

- **Total LOC**: ~2,500-3,000 (similar to Phase 1 services)
- **Test Coverage**: > 90%
- **Build Time**: ~30 seconds
- **Binary Size**: ~18MB
- **Completion Time**: 2-3 days for full implementation

## Definition of Done

- [ ] All layers 1-6 complete
- [ ] > 90% test coverage
- [ ] Clean build (go build, go vet, go fmt)
- [ ] All hard rules verified
- [ ] Health checks working
- [ ] OpenTelemetry integrated
- [ ] Docker image builds
- [ ] M1_VERIFICATION.md complete
- [ ] Ready for production deployment

---

This service reaches the same maturity level as auth, employee, attendance, leave, and notification services.
