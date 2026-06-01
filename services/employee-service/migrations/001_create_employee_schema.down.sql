-- ============================================================
-- Rollback: 001_create_employee_schema
-- ============================================================

DROP TABLE IF EXISTS employee.processed_events;
DROP TABLE IF EXISTS employee.employees;
DROP TABLE IF EXISTS employee.positions;
DROP TABLE IF EXISTS employee.departments;

DROP TYPE IF EXISTS employee.gender;
DROP TYPE IF EXISTS employee.contract_type;
DROP TYPE IF EXISTS employee.employment_status;

DROP SCHEMA IF EXISTS employee;
