-- ============================================================
-- Migration: 001_create_leave_schema
-- Service:   leave-service
-- Schema:    leave
-- ============================================================

CREATE SCHEMA IF NOT EXISTS leave;
GRANT USAGE ON SCHEMA leave TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: leave_types
-- Configured per tenant by HR Admin.
-- ────────────────────────────────────────────────────────────
CREATE TABLE leave.leave_types (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    code                TEXT NOT NULL,  -- e.g. "annual", "sick", "maternity"
    name                TEXT NOT NULL,  -- localized display name
    max_days_per_year   INT NOT NULL DEFAULT 12,
    requires_document   BOOLEAN NOT NULL DEFAULT false,
    is_paid             BOOLEAN NOT NULL DEFAULT true,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, code)
);

CREATE INDEX leave_types_tenant_id_idx ON leave.leave_types (tenant_id);

ALTER TABLE leave.leave_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave.leave_types FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON leave.leave_types
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON leave.leave_types TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: leave_requests
-- ────────────────────────────────────────────────────────────
CREATE TYPE leave.leave_status AS ENUM (
    'pending', 'approved', 'rejected', 'cancelled'
);

CREATE TABLE leave.leave_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    employee_id         UUID NOT NULL,    -- references employee.employees.id (cross-service, no FK)
    leave_type_id       UUID NOT NULL REFERENCES leave.leave_types(id),
    status              leave.leave_status NOT NULL DEFAULT 'pending',

    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    days_count          INT NOT NULL,     -- computed by service

    reason              TEXT NOT NULL,
    rejection_reason    TEXT,
    approved_by_id      UUID,             -- employee_id of approver

    document_key        TEXT,             -- MinIO object key

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at         TIMESTAMPTZ,

    CONSTRAINT leave_date_order CHECK (end_date >= start_date)
);

CREATE INDEX leave_requests_tenant_id_idx ON leave.leave_requests (tenant_id);
CREATE INDEX leave_requests_employee_id_idx ON leave.leave_requests (tenant_id, employee_id);
CREATE INDEX leave_requests_status_idx ON leave.leave_requests (tenant_id, status);
CREATE INDEX leave_requests_approver_idx ON leave.leave_requests (tenant_id, approved_by_id);
CREATE INDEX leave_requests_date_range_idx ON leave.leave_requests (tenant_id, start_date, end_date);

ALTER TABLE leave.leave_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave.leave_requests FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON leave.leave_requests
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON leave.leave_requests TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: leave_balances
-- Tracks annual entitlement and usage per employee per year.
-- Updated by leave_balance_transactions.
-- ────────────────────────────────────────────────────────────
CREATE TABLE leave.leave_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    employee_id     UUID NOT NULL,
    leave_type_id   UUID NOT NULL REFERENCES leave.leave_types(id),
    year            INT NOT NULL,
    entitled_days   INT NOT NULL DEFAULT 0,
    used_days       INT NOT NULL DEFAULT 0,
    pending_days    INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, employee_id, leave_type_id, year)
);

CREATE INDEX leave_balances_tenant_id_idx ON leave.leave_balances (tenant_id);
CREATE INDEX leave_balances_employee_id_idx ON leave.leave_balances (tenant_id, employee_id, year);

ALTER TABLE leave.leave_balances ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave.leave_balances FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON leave.leave_balances
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON leave.leave_balances TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: processed_events  (idempotency table — ADR-0003)
-- ────────────────────────────────────────────────────────────
CREATE TABLE leave.processed_events (
    event_id     UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- No RLS — internal service table, not tenant-scoped.
GRANT INSERT, SELECT ON leave.processed_events TO hris_app;
