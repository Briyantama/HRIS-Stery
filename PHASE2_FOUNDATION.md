# Phase 2 Foundation Architecture

**Status:** Design Complete, Ready for Implementation  
**Services:** audit-service, document-service  
**Timeline:** Follows Phase 1 RC1 release

---

## Audit-Service Specification

### Purpose

Immutable, tamper-resistant audit trail for all HRIS operations. Captures WHO, WHAT, WHEN, WHERE for compliance and forensics.

### Consumers

Subscribes to NATS events:
- `hris.identity.*` (login, logout, password change)
- `hris.workforce.*` (employee lifecycle)
- `hris.operations.*` (leave, attendance)
- `hris.notification.*` (delivery events)

### Database Schema

```sql
CREATE SCHEMA audit;

CREATE TABLE audit.entries (
  entry_id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  actor_id UUID NOT NULL,
  action VARCHAR(50) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id UUID,
  description TEXT,
  success BOOLEAN DEFAULT TRUE,
  error_message TEXT,
  changes JSONB,  -- before/after values
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- RLS disabled (audit service reads all tenants)
-- No UPDATE/DELETE allowed (INSERT-ONLY policy)

CREATE INDEX idx_audit_tenant_created ON audit.entries (tenant_id, created_at DESC);
CREATE INDEX idx_audit_actor_created ON audit.entries (actor_id, created_at DESC);
CREATE INDEX idx_audit_resource ON audit.entries (resource_type, resource_id);
```

### Implementation Path

1. **Layer 1 (Proto):** ✅ DONE - `proto/hris/audit/v1/audit.proto`
2. **Layer 2 (Domain):**
   - `AuditEntry` aggregate (immutable, created only)
   - `AuditEntryID`, `ActorID` value objects
   - `QueryAuditFilters` for search
   - Repository port

3. **Layer 3 (Application):**
   - `RecordAuditCommand` handler
   - `QueryAuditQuery` handler (read-side)
   - NATS consumers (identity, workforce, operations, notification)

4. **Layer 4 (Infrastructure):**
   - PostgreSQL repository (INSERT-ONLY)
   - NATS consumer subscriptions
   - Idempotency pattern (same as Phase 1)

5. **Layer 5 (Interfaces):**
   - gRPC service (Record, Query, GetEntry)

6. **Layer 6 (Server):**
   - Main entrypoint with PostgreSQL + NATS setup

### Key Invariants

- ✅ **Immutable:** Entries created only, never modified
- ✅ **Tenant-Aware:** Each entry tagged with tenant_id
- ✅ **Idempotent:** Event ID deduplication via processed_events
- ✅ **Queryable:** Search by actor, action, resource, date range
- ✅ **No Cascade Deletes:** Audit persists even if source data deleted

---

## Document-Service Specification

### Purpose

Secure, versioned document storage for employee files, leave attachments, HR documents. Integrates with MinIO for S3-compatible object storage.

### Features

1. **File Upload/Download**
   - Multi-part uploads
   - Pre-signed URLs (time-limited)
   - Virus scanning (ClamAV integration)

2. **Versioning**
   - Track all file versions
   - Rollback to previous versions
   - Metadata per version (who, when, why)

3. **Metadata**
   - File name, size, MIME type
   - Associated entity (employee, leave record)
   - Retention policy (auto-delete after N years)
   - Classification (public, confidential, secret)

4. **Access Control**
   - File-level permissions (who can download)
   - Tenant isolation
   - Audit integration (all access logged)

### Database Schema

```sql
CREATE SCHEMA document;

CREATE TABLE document.files (
  file_id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  uploaded_by UUID NOT NULL,
  entity_type VARCHAR(50) NOT NULL,  -- employee, leave, etc
  entity_id UUID NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  mime_type VARCHAR(100),
  size_bytes BIGINT,
  minio_bucket VARCHAR(255) NOT NULL,
  minio_key VARCHAR(1024) NOT NULL,
  status VARCHAR(20) DEFAULT 'uploaded',  -- uploaded, scanned, failed
  classification VARCHAR(20) DEFAULT 'public',
  retention_days INT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- RLS: Users can see only files for their tenant/entity
ALTER TABLE document.files ENABLE ROW LEVEL SECURITY;

CREATE TABLE document.file_versions (
  version_id UUID PRIMARY KEY,
  file_id UUID NOT NULL REFERENCES document.files,
  version_number INT NOT NULL,
  minio_key VARCHAR(1024) NOT NULL,
  size_bytes BIGINT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  created_by UUID NOT NULL
);
```

### NATS Integration

Publishes:
- `hris.document.uploaded` - New file uploaded
- `hris.document.accessed` - File downloaded (audit)
- `hris.document.deleted` - File deleted (audit)

Consumes:
- `hris.workforce.employee.terminated` - Auto-delete confidential employee docs
- `hris.operations.leave.completed` - Auto-delete after retention period

### MinIO Configuration

```yaml
# docker-compose addition:
minio:
  image: minio/minio
  environment:
    MINIO_ROOT_USER: minioadmin
    MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}
  volumes:
    - minio_data:/data
  ports:
    - "9000:9000"  # API
    - "9001:9001"  # Console
```

### Implementation Path

1. **Layer 1 (Proto):** Define DocumentService with Upload, Download, ListFiles, GetMetadata, DeleteFile
2. **Layer 2 (Domain):** DocumentFile aggregate, version tracking
3. **Layer 3 (Application):** Upload, Download, List, Delete commands/queries
4. **Layer 4 (Infrastructure):** MinIO client, PostgreSQL metadata repository
5. **Layer 5 (Interfaces):** gRPC service with streaming uploads
6. **Layer 6 (Server):** MinIO + PostgreSQL wiring

---

## Phase 2 Implementation Order

### Priority 1: audit-service (Week 1-2)

- High value: Compliance, forensics, security investigations
- Low risk: Append-only, read-only queries
- Pattern: Identical to Phase 1 services

**Effort:** ~40-50 hours (~5-6 days)

### Priority 2: document-service (Week 2-3)

- High value: Employee document management, leave attachments
- Medium risk: File upload/download, MinIO integration
- External dependencies: MinIO, virus scanning

**Effort:** ~30-40 hours (~4-5 days)

### Priority 3: Enhanced Features (Week 3+)

- OAuth2/SSO integration
- Advanced notifications (WHATSAPP, PUSH)
- Payroll calculations (tax, deductions)
- Performance review workflows
- Recruitment pipeline

---

## Quality Standards (Phase 2)

Same as Phase 1:
- ✅ Clean architecture (Domain → Application → Infrastructure → Interfaces)
- ✅ 100% unit test coverage per layer
- ✅ RLS enforcement on all tenant tables
- ✅ Event idempotency pattern
- ✅ No forbidden patterns (interface{}, raw SQL, business logic in handlers)
- ✅ Comprehensive error handling
- ✅ Structured logging with tenant_id + request_id
- ✅ OpenTelemetry tracing
- ✅ Docker image per service
- ✅ Health checks (gRPC Health service)

---

## Deployment Checklist (Phase 2)

- [ ] audit-service proto compiled
- [ ] audit-service domain layer tested
- [ ] audit-service application layer tested
- [ ] audit-service infrastructure complete
- [ ] audit-service integrated with Phase 1 event streams
- [ ] audit-service Docker image built
- [ ] audit-service health checks working
- [ ] document-service proto compiled
- [ ] MinIO cluster deployed
- [ ] document-service infrastructure complete
- [ ] document-service Docker image built
- [ ] Integration tests pass (audit + document)
- [ ] PHASE2_RELEASE.md completed

---

## Risk Mitigation

### audit-service

**Risk:** Audit trail corruption  
**Mitigation:** Immutable INSERT-ONLY schema, RLS disabled to prevent accidental deletion

**Risk:** High query volume  
**Mitigation:** Partition by date, indexing strategy, read-only queries

### document-service

**Risk:** File storage explosion  
**Mitigation:** Retention policies, classification, auto-delete

**Risk:** Malicious uploads  
**Mitigation:** ClamAV virus scanning, file type validation

**Risk:** Unauthorized access  
**Mitigation:** RLS per tenant, audit logging, pre-signed URL expiration

---

## Success Criteria (Phase 2)

- ✅ All audit events from Phase 1 services captured
- ✅ Zero audit trail gaps (100% idempotency)
- ✅ Document service handles 10K+ file uploads
- ✅ All services deploy via docker-compose
- ✅ Health checks green across all services
- ✅ PHASE2_RELEASE.md complete + signed off

---

This document is the foundation for Phase 2. Proceed with implementation only after Phase 1 RC1 is released.
