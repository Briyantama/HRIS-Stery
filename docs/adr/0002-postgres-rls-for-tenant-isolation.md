# ADR-0002: PostgreSQL Row-Level Security for Tenant Isolation

**Status:** Accepted  
**Date:** 2026-05-30  
**Deciders:** Architecture Team  
**Supersedes:** —

---

## Context

This is a multi-tenant SaaS. Every table that contains tenant data has a `tenant_id UUID NOT NULL` column. The question is: what mechanism enforces that a query for Tenant A cannot return data belonging to Tenant B?

### The Risk

Application-level enforcement (e.g., always adding `WHERE tenant_id = $1`) is fragile. A single missing clause — in a new endpoint, a background job, a report query, or a hasty hotfix — produces a cross-tenant data leak. In an HR system handling salary, PII, and performance data, this is a regulatory incident, not just a bug.

### Option A: Application-level enforcement only

Every query includes `WHERE tenant_id = ?`. Enforced by code review and testing.

Problems:
- Single point of human failure. One missed clause = breach.
- Test coverage cannot guarantee every code path.
- Background jobs and migrations are especially prone to missing this filter.
- Provides no defense against SQL injection that bypasses application logic.

### Option B: Separate database per tenant

Each tenant gets its own PostgreSQL database or connection string.

Problems:
- Operational overhead grows linearly with tenant count.
- Connection pooling becomes complex (PgBouncer per tenant or a very large pool).
- Schema migrations must be applied to N databases simultaneously.
- Not justified at this scale (target: hundreds to low-thousands of tenants).

### Option C: PostgreSQL Row-Level Security (selected)

PostgreSQL RLS allows defining security policies at the database level that are enforced for every query, regardless of what the application sends.

When RLS is enabled on a table, PostgreSQL evaluates the policy before returning or modifying any row. The policy references a session variable (`current_setting('app.tenant_id')`) set by the application at connection time.

---

## Decision

**Enable PostgreSQL RLS on every table that contains `tenant_id`.**

### Implementation Pattern

**Step 1: Create an application role with no BYPASSRLS**

```sql
CREATE ROLE hris_app LOGIN PASSWORD '...' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
-- hris_app does NOT have BYPASSRLS privilege
```

**Step 2: Enable RLS and create policy**

```sql
ALTER TABLE employee.employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee.employees FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON employee.employees
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);
```

**Step 3: Set session variable on every connection**

Each Go service, immediately after acquiring a connection from the pool, executes:

```sql
SELECT set_config('app.tenant_id', $1, true);
-- The third argument `true` scopes the setting to the current transaction only.
```

This is implemented in `services/_shared/postgres/rls.go` as a `pgx` connection lifecycle hook.

**Step 4: Migrations and admin tasks run with BYPASSRLS**

A separate `hris_admin` role (used only for migrations and schema changes) has `BYPASSRLS`. Never used by application code.

---

## Audit Table Exception

The `audit.audit_events` table uses a different RLS policy:

```sql
ALTER TABLE audit.audit_events ENABLE ROW LEVEL SECURITY;

-- Only INSERT is allowed through application role
CREATE POLICY audit_insert ON audit.audit_events FOR INSERT
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- SELECT is allowed for same tenant
CREATE POLICY audit_select ON audit.audit_events FOR SELECT
  USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- DELETE is not allowed through any policy (no DELETE policy = deny all DELETE)
```

This makes the audit log append-only at the database level.

---

## Consequences

### Positive
- Tenant isolation is enforced by the database engine, not application code. It cannot be bypassed by application bugs.
- A missing `WHERE tenant_id = ?` in application code does not cause a data breach — RLS filters it.
- SQL injection that reaches the database still cannot cross tenant boundaries.
- Compliance-friendly: isolation is provable at the infrastructure level.

### Negative
- Requires setting `app.tenant_id` on every connection before any query. The shared `postgres` package handles this.
- `EXPLAIN` output includes RLS filter steps — understand this when analyzing query plans.
- Background jobs and migrations that legitimately need cross-tenant access must use the `hris_admin` role explicitly. This must be documented in each job.
- PgBouncer transaction pooling mode is required (session pooling would cause session variable leakage). Configure PgBouncer with `pool_mode = transaction`.

### Neutral
- Performance impact of RLS is negligible when `tenant_id` is indexed (which it must be on every table).
- All services share the same physical Postgres instance but different schemas. RLS policies are schema-scoped.

---

## Migration Template

Every new table migration must follow this template:

```sql
-- {NNN}_create_{table_name}.up.sql
CREATE TABLE {schema}.{table_name} (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    -- ... domain columns ...
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX {table_name}_tenant_id_idx ON {schema}.{table_name} (tenant_id);

ALTER TABLE {schema}.{table_name} ENABLE ROW LEVEL SECURITY;
ALTER TABLE {schema}.{table_name} FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON {schema}.{table_name}
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON {schema}.{table_name} TO hris_app;
```

This template is enforced by CI linting of migration files.

---

## Compliance

- Every `CREATE TABLE` migration must enable RLS if the table has `tenant_id`.
- Every new Go service must use `services/_shared/postgres/rls.go` for connection setup.
- The `hris_admin` role credentials must never appear in application service environment variables.
- Penetration testing must include a cross-tenant access test before each production release.
