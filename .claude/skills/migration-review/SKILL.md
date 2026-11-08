# Migration Review Skill

**Trigger:** Before writing any database migration file

**Purpose:** Ensure migrations are reversible and follow project conventions

## Instructions

1. Migrations live in `apps/api/internal/db/migrations/` as goose files `{NNNN}_description.sql`
2. For EACH migration being created:
   - [ ] File has `-- +goose Up` and `-- +goose Down` (Down fully reverses Up)
   - [ ] Number `{NNNN}` is unique and inside the service's range (auth 001x, employee 002x, attendance 003x, leave 004x, notification 005x, audit 006x, document 007x)
   - [ ] Functions and DO blocks are wrapped in `-- +goose StatementBegin` / `-- +goose StatementEnd`
   - [ ] Down drops in reverse dependency order and never drops a schema or role owned by `0001_bootstrap.sql`
3. Verify RLS policies:
   - [ ] All tenant-scoped tables have `ALTER TABLE ... FORCE ROW LEVEL SECURITY;`
   - [ ] All tenant-scoped tables have `CREATE POLICY tenant_isolation ON ...`
   - [ ] Policies check `tenant_id = current_setting('app.tenant_id')::uuid`
4. Run verification:
   - [ ] `go run ./cmd/migrate up`, `down`, `up` succeeds against a throwaway database
   - [ ] No hardcoded secrets or credentials
   - [ ] Proper grant statements for hris_app role

## Migration Checklist

- [ ] `-- +goose Up` and `-- +goose Down` sections present
- [ ] Number unique and in the service range
- [ ] Down migration fully reverses up migration
- [ ] RLS policies on all tenant-scoped tables
- [ ] No hardcoded data or secrets
- [ ] Proper role grants (hris_app, hris_admin)
- [ ] All table changes documented in comments

## RLS Template

```sql
-- +goose Up
ALTER TABLE {schema}.{table} ENABLE ROW LEVEL SECURITY;
ALTER TABLE {schema}.{table} FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON {schema}.{table}
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- +goose Down (reverse order)
DROP POLICY tenant_isolation ON {schema}.{table};
ALTER TABLE {schema}.{table} DISABLE ROW LEVEL SECURITY;
DROP TABLE {schema}.{table};
```

## Examples

**Trigger:** "Add new column to employees table"
→ Up (ADD COLUMN ...) and Down (DROP COLUMN ...) sections required

**Trigger:** "Create tenant-scoped table"
→ Up must include RLS policy; down must reverse it

**Trigger:** "Add system table (not tenant-scoped)"
→ No RLS needed, but still requires reversible down migration
