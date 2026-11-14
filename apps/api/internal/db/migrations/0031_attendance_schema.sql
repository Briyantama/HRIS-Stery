-- +goose Up
CREATE SCHEMA IF NOT EXISTS attendance;

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
    CONSTRAINT attendance_check_in_before_out CHECK (check_in_at IS NULL OR check_out_at IS NULL OR check_in_at < check_out_at)
);

-- One regular record per employee and day; manual overrides may repeat.
-- (A table UNIQUE constraint cannot carry a WHERE clause, so this is a partial unique index.)
CREATE UNIQUE INDEX attendance_unique_per_day
    ON attendance.attendance_records (tenant_id, employee_id, date) WHERE NOT is_override;

CREATE INDEX idx_attendance_employee_date ON attendance.attendance_records (employee_id, date DESC);
CREATE INDEX idx_attendance_tenant ON attendance.attendance_records (tenant_id);
CREATE INDEX idx_attendance_date_range ON attendance.attendance_records (date DESC);
CREATE INDEX idx_attendance_active_checkins ON attendance.attendance_records (check_in_at DESC)
    WHERE check_in_at IS NOT NULL AND check_out_at IS NULL AND is_override = FALSE;

ALTER TABLE attendance.attendance_records ENABLE ROW LEVEL SECURITY;

CREATE POLICY attendance_tenant_isolation
    ON attendance.attendance_records
    FOR ALL
    USING (tenant_id = (current_setting('app.tenant_id')::UUID));

GRANT USAGE ON SCHEMA attendance TO hris_app;
GRANT SELECT, INSERT, UPDATE ON TABLE attendance.attendance_records TO hris_app;

-- +goose Down
DROP POLICY IF EXISTS attendance_tenant_isolation ON attendance.attendance_records;
DROP INDEX IF EXISTS attendance.idx_attendance_active_checkins;
DROP INDEX IF EXISTS attendance.idx_attendance_date_range;
DROP INDEX IF EXISTS attendance.idx_attendance_tenant;
DROP INDEX IF EXISTS attendance.idx_attendance_employee_date;
DROP INDEX IF EXISTS attendance.attendance_unique_per_day;
DROP TABLE IF EXISTS attendance.attendance_records;
