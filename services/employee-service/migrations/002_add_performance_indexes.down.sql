-- Rollback performance indexes

DROP INDEX IF EXISTS idx_employees_dept_status;
DROP INDEX IF EXISTS idx_employees_manager_id;
DROP INDEX IF EXISTS idx_employees_status;
DROP INDEX IF EXISTS idx_employees_department_id;
