-- Add performance indexes for common query patterns

-- Index for finding attendance by employee and date (most common query)
CREATE INDEX idx_attendance_employee_date ON attendance.attendance_records(employee_id, date DESC);

-- Index for filtering by status
CREATE INDEX idx_attendance_status ON attendance.attendance_records(status);

-- Index for date range queries
CREATE INDEX idx_attendance_date ON attendance.attendance_records(date DESC);
