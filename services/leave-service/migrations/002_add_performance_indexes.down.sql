-- Rollback performance indexes

DROP INDEX IF EXISTS idx_leave_requests_approved_by;
DROP INDEX IF EXISTS idx_leave_requests_start_date;
DROP INDEX IF EXISTS idx_leave_requests_status;
DROP INDEX IF EXISTS idx_leave_requests_employee_status;
DROP INDEX IF EXISTS idx_leave_requests_employee_id;
