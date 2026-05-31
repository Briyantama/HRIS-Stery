-- ============================================================
-- Rollback: 001_create_leave_schema
-- ============================================================

DROP TABLE IF EXISTS leave.processed_events;
DROP TABLE IF EXISTS leave.leave_balances;
DROP TABLE IF EXISTS leave.leave_requests;
DROP TABLE IF EXISTS leave.leave_types;

DROP TYPE IF EXISTS leave.leave_status;

DROP SCHEMA IF EXISTS leave;
