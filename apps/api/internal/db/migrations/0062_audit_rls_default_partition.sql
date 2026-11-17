-- +goose Up
-- Tenant isolation for audit.entries (same pattern as document.files) plus a
-- DEFAULT partition so inserts past the last monthly partition (2027-12) don't fail.
-- Writer and readers already go through pgtenant.WithTenantTx (sets app.tenant_id).
-- ponytail: no auto-creation of monthly partitions; add 2028+ partitions by hand
-- BEFORE the default holds rows in that range (Postgres refuses to attach over them).

ALTER TABLE audit.entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit.entries FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON audit.entries
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE TABLE audit.audit_entries_default PARTITION OF audit.entries DEFAULT;

-- +goose Down
DROP TABLE IF EXISTS audit.audit_entries_default;
DROP POLICY IF EXISTS tenant_isolation ON audit.entries;
ALTER TABLE audit.entries NO FORCE ROW LEVEL SECURITY;
ALTER TABLE audit.entries DISABLE ROW LEVEL SECURITY;
