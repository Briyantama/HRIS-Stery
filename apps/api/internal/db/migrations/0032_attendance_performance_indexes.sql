-- +goose Up
-- idx_attendance_employee_date already exists from 0031; IF NOT EXISTS keeps this step a no-op for it.
CREATE INDEX IF NOT EXISTS idx_attendance_employee_date ON attendance.attendance_records (employee_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_attendance_status ON attendance.attendance_records (status);
CREATE INDEX IF NOT EXISTS idx_attendance_date ON attendance.attendance_records (date DESC);

-- +goose Down
-- idx_attendance_employee_date belongs to 0031 and is dropped there.
DROP INDEX IF EXISTS attendance.idx_attendance_date;
DROP INDEX IF EXISTS attendance.idx_attendance_status;
