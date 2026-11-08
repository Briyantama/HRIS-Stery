-- +goose Up
-- Add performance indexes for common query patterns.
-- 0041 already creates idx_leave_requests_employee_id, idx_leave_requests_status and
-- idx_leave_requests_approved_by, so those are IF NOT EXISTS no-ops here.
CREATE INDEX IF NOT EXISTS idx_leave_requests_employee_id ON leave.leave_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_leave_requests_employee_status ON leave.leave_requests(employee_id, status);
CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave.leave_requests(status);
CREATE INDEX IF NOT EXISTS idx_leave_requests_start_date ON leave.leave_requests(start_date DESC);
CREATE INDEX IF NOT EXISTS idx_leave_requests_approved_by ON leave.leave_requests(approved_by_id);

-- +goose Down
-- Only the indexes this migration introduced; the others belong to 0041.
DROP INDEX IF EXISTS leave.idx_leave_requests_start_date;
DROP INDEX IF EXISTS leave.idx_leave_requests_employee_status;
