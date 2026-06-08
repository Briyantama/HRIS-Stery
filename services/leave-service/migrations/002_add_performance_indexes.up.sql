-- Add performance indexes for common query patterns

-- Index for finding leave requests by employee
CREATE INDEX idx_leave_requests_employee_id ON leave.leave_requests(employee_id);

-- Composite index for employee and status filtering
CREATE INDEX idx_leave_requests_employee_status ON leave.leave_requests(employee_id, status);

-- Index for filtering by status
CREATE INDEX idx_leave_requests_status ON leave.leave_requests(status);

-- Index for date range queries
CREATE INDEX idx_leave_requests_start_date ON leave.leave_requests(start_date DESC);

-- Index for approver lookups
CREATE INDEX idx_leave_requests_approved_by ON leave.leave_requests(approved_by_id);
