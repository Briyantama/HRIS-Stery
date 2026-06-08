-- Add performance indexes for common query patterns

-- Index for filtering employees by department
CREATE INDEX idx_employees_department_id ON employee.employees(department_id);

-- Index for filtering employees by status
CREATE INDEX idx_employees_status ON employee.employees(status);

-- Composite index for common list query filters
CREATE INDEX idx_employees_dept_status ON employee.employees(department_id, status);

-- Index for manager lookups (direct reports)
CREATE INDEX idx_employees_manager_id ON employee.employees(manager_id);
