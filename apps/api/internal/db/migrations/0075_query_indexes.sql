-- +goose Up
-- Indexes for the two list queries that sort a whole tenant's rows before applying LIMIT (EXPLAIN ANALYZE
-- on 250k employees / 2.4M attendance rows over 24 skewed tenants, queried as hris_app so RLS applies).

-- employee list (ORDER BY created_at DESC LIMIT n) was Index Scan on the tenant_id index + Sort of every
-- tenant row (4231 buffers for the largest tenant); with this index it reads the first page only (43 buffers).
CREATE INDEX IF NOT EXISTS idx_employees_tenant_created ON employee.employees (tenant_id, created_at DESC);

-- attendance list / daily summary (ORDER BY date DESC, created_at DESC, id) walked idx_attendance_date and
-- discarded other tenants' rows (5910 buffers, 4319 rows removed for the largest tenant). The column order
-- matches the ORDER BY exactly, so LIMIT/OFFSET stop early and date-range filters become index conditions.
CREATE INDEX IF NOT EXISTS idx_attendance_tenant_date_created ON attendance.attendance_records (tenant_id, date DESC, created_at DESC, id);

-- +goose Down
DROP INDEX IF EXISTS attendance.idx_attendance_tenant_date_created;
DROP INDEX IF EXISTS employee.idx_employees_tenant_created;
