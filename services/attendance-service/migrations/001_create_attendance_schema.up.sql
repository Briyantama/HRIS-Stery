-- Create attendance schema
CREATE SCHEMA IF NOT EXISTS attendance AUTHORIZATION postgres;

-- Create attendance_records table
CREATE TABLE attendance.attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    date DATE NOT NULL,
    check_in_at TIMESTAMPTZ,
    check_out_at TIMESTAMPTZ,
    check_in_latitude DOUBLE PRECISION,
    check_in_longitude DOUBLE PRECISION,
    check_out_latitude DOUBLE PRECISION,
    check_out_longitude DOUBLE PRECISION,
    status VARCHAR(50) NOT NULL DEFAULT 'ABSENT',
    work_duration_mins INTEGER,
    notes TEXT DEFAULT '',
    created_by VARCHAR(255) NOT NULL,
    is_override BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT attendance_valid_status CHECK (status IN ('PRESENT', 'ABSENT', 'LATE', 'HALF_DAY', 'ON_LEAVE')),
    CONSTRAINT attendance_check_in_before_out CHECK (check_in_at IS NULL OR check_out_at IS NULL OR check_in_at < check_out_at),
    CONSTRAINT attendance_unique_per_day UNIQUE (tenant_id, employee_id, date) WHERE NOT is_override
);

-- Create indexes for common queries
CREATE INDEX idx_attendance_employee_date ON attendance.attendance_records (employee_id, date DESC);
CREATE INDEX idx_attendance_tenant ON attendance.attendance_records (tenant_id);
CREATE INDEX idx_attendance_date_range ON attendance.attendance_records (date DESC);
CREATE INDEX idx_attendance_active_checkins ON attendance.attendance_records (check_in_at DESC)
    WHERE check_in_at IS NOT NULL AND check_out_at IS NULL AND is_override = FALSE;

-- Enable Row-Level Security (RLS)
ALTER TABLE attendance.attendance_records ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Tenants can only see their own records
CREATE POLICY attendance_tenant_isolation
    ON attendance.attendance_records
    FOR ALL
    USING (tenant_id = (current_setting('app.tenant_id')::UUID));

-- GRANT permissions to app role
GRANT USAGE ON SCHEMA attendance TO app;
GRANT SELECT, INSERT, UPDATE ON TABLE attendance.attendance_records TO app;
