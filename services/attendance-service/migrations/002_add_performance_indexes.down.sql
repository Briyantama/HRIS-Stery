-- Rollback performance indexes

DROP INDEX IF EXISTS idx_attendance_date;
DROP INDEX IF EXISTS idx_attendance_status;
DROP INDEX IF EXISTS idx_attendance_employee_date;
