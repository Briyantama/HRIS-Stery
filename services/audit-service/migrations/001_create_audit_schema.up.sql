-- ============================================================
-- Migration: 001_create_audit_schema
-- Service:   audit-service
-- Schema:    audit
-- ============================================================
-- Creates the audit schema with immutable, append-only audit entries.
-- Note: RLS disabled on audit.entries (audit-service reads all tenants).
-- INSERT-ONLY enforcement: No UPDATE/DELETE operations allowed.
-- Run with hris_admin role (BYPASSRLS). Application role is hris_app.
-- ============================================================

-- Roles (idempotent — run in cluster bootstrap, not per-migration)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'hris_app') THEN
    CREATE ROLE hris_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'hris_admin') THEN
    CREATE ROLE hris_admin LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT BYPASSRLS;
  END IF;
END $$;

-- Schema
CREATE SCHEMA IF NOT EXISTS audit;
GRANT USAGE ON SCHEMA audit TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: entries (append-only, partitioned by created_at)
-- RLS disabled: audit-service reads all tenants for compliance.
-- INSERT-ONLY: No UPDATE/DELETE possible.
-- ────────────────────────────────────────────────────────────
CREATE TABLE audit.entries (
    entry_id         UUID NOT NULL DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL,
    actor_id         TEXT NOT NULL,
    action           VARCHAR(50) NOT NULL,
    resource_type    VARCHAR(50) NOT NULL,
    resource_id      UUID,
    description      TEXT,
    success          BOOLEAN NOT NULL DEFAULT TRUE,
    error_message    TEXT,
    changes          JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (entry_id, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions for each month (12 months forward + 3 back)
-- Partition range: 2023-01 to 2027-12
DO $$
DECLARE
    start_date DATE := '2023-01-01'::DATE;
    end_date DATE := '2027-12-31'::DATE;
    part_start DATE;
    part_end DATE;
    part_name TEXT;
BEGIN
    part_start := start_date;
    WHILE part_start < end_date LOOP
        part_end := part_start + INTERVAL '1 month';
        part_name := 'audit_entries_' || TO_CHAR(part_start, 'YYYY_MM');
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit.entries FOR VALUES FROM (%L::TIMESTAMPTZ) TO (%L::TIMESTAMPTZ)',
            part_name,
            part_start::TIMESTAMPTZ,
            part_end::TIMESTAMPTZ
        );
        part_start := part_end;
    END LOOP;
END $$;

-- Indexes on partitioned table
CREATE INDEX idx_audit_entries_tenant_created ON audit.entries (tenant_id, created_at DESC);
CREATE INDEX idx_audit_entries_actor_created ON audit.entries (actor_id, created_at DESC);
CREATE INDEX idx_audit_entries_action ON audit.entries (action, created_at DESC);
CREATE INDEX idx_audit_entries_resource ON audit.entries (resource_type, resource_id, created_at DESC);

-- INSERT-ONLY enforcement: hris_app can only INSERT on audit.entries
-- No SELECT, UPDATE, DELETE for hris_app (audit-service uses hris_admin role or special reader)
GRANT INSERT ON audit.entries TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: processed_events (idempotency for NATS consumers)
-- ────────────────────────────────────────────────────────────
CREATE TABLE audit.processed_events (
    event_id         UUID PRIMARY KEY,
    subject          VARCHAR(255) NOT NULL,
    processed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_processed_events_subject ON audit.processed_events (subject);
CREATE INDEX idx_processed_events_created ON audit.processed_events (created_at DESC);

GRANT SELECT, INSERT ON audit.processed_events TO hris_app;
