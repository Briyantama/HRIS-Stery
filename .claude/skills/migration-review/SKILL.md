# Migration Review Skill

**Trigger:** Before writing any database migration file

**Purpose:** Ensure migrations are reversible and follow project conventions

## Instructions

1. Check if migration files exist in the service's `migrations/` directory
2. For EACH migration being created:
   - [ ] `{NNN}_description.up.sql` exists
   - [ ] `{NNN}_description.down.sql` exists (fully reverses the up migration)
   - [ ] Sequence number `{NNN}` is monotonically increasing
   - [ ] Down migration drops in reverse dependency order
3. Verify RLS policies:
   - [ ] All tenant-scoped tables have `ALTER TABLE ... FORCE ROW LEVEL SECURITY;`
   - [ ] All tenant-scoped tables have `CREATE POLICY tenant_isolation ON ...`
   - [ ] Policies check `tenant_id = current_setting('app.tenant_id')::uuid`
4. Run verification:
   - [ ] `up → down → up` cycle succeeds (CI checks this)
   - [ ] No hardcoded secrets or credentials
   - [ ] Proper grant statements for hris_app role

## Migration Checklist

- [ ] Both up.sql and down.sql files exist
- [ ] Sequence numbers correct (e.g., 001, 002, 003)
- [ ] Down migration fully reverses up migration
- [ ] RLS policies on all tenant-scoped tables
- [ ] No hardcoded data or secrets
- [ ] Proper role grants (hris_app, hris_admin)
- [ ] All table changes documented in comments

## RLS Template

```sql
-- In up migration
ALTER TABLE {schema}.{table} ENABLE ROW LEVEL SECURITY;
ALTER TABLE {schema}.{table} FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON {schema}.{table}
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- In down migration (reverse order)
DROP POLICY tenant_isolation ON {schema}.{table};
ALTER TABLE {schema}.{table} DISABLE ROW LEVEL SECURITY;
DROP TABLE {schema}.{table};
```

## Examples

**Trigger:** "Add new column to employees table"
→ Both up.sql (ADD COLUMN ...) and down.sql (DROP COLUMN ...) required

**Trigger:** "Create tenant-scoped table"
→ Up must include RLS policy; down must reverse it

**Trigger:** "Add system table (not tenant-scoped)"
→ No RLS needed, but still requires reversible down migration
