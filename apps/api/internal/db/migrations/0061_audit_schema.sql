-- +goose Up
-- ============================================================
-- Migration: 001_create_audit_schema (audit service)
-- Schema:    audit (created in 0001_bootstrap; role hris_app too)
-- ============================================================
-- Immutable, append-only audit entries. RLS is enabled in 0062.
-- INSERT-ONLY: the application role gets no UPDATE or DELETE.
-- ============================================================

CREATE SCHEMA IF NOT EXISTS audit;
GRANT USAGE ON SCHEMA audit TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: entries (append-only, partitioned by created_at)
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

-- One partition per month, 2023-01 to 2027-12.
-- +goose StatementBegin
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
            'CREATE TABLE IF NOT EXISTS audit.%I PARTITION OF audit.entries FOR VALUES FROM (%L::TIMESTAMPTZ) TO (%L::TIMESTAMPTZ)',
            part_name,
            part_start::TIMESTAMPTZ,
            part_end::TIMESTAMPTZ
        );
        part_start := part_end;
    END LOOP;
END $$;
-- +goose StatementEnd

CREATE INDEX idx_audit_entries_tenant_created ON audit.entries (tenant_id, created_at DESC);
CREATE INDEX idx_audit_entries_actor_created ON audit.entries (actor_id, created_at DESC);
CREATE INDEX idx_audit_entries_action ON audit.entries (action, created_at DESC);
CREATE INDEX idx_audit_entries_resource ON audit.entries (resource_type, resource_id, created_at DESC);

-- INSERT for the recorder, SELECT for the read routes; never UPDATE or DELETE.
GRANT SELECT, INSERT ON audit.entries TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: processed_events (idempotency for stream consumers)
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

-- +goose Down
DROP TABLE IF EXISTS audit.processed_events;
DROP TABLE IF EXISTS audit.entries;
