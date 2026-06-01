-- Drop RLS policies
DROP POLICY IF EXISTS attendance_tenant_isolation ON attendance.attendance_records;

-- Drop indexes
DROP INDEX IF EXISTS idx_attendance_active_checkins;
DROP INDEX IF EXISTS idx_attendance_date_range;
DROP INDEX IF EXISTS idx_attendance_tenant;
DROP INDEX IF EXISTS idx_attendance_employee_date;

-- Drop table
DROP TABLE IF EXISTS attendance.attendance_records;

-- Drop schema
DROP SCHEMA IF EXISTS attendance;
