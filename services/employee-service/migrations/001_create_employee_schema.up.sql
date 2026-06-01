-- ============================================================
-- Migration: 001_create_employee_schema
-- Service:   employee-service
-- Schema:    employee
-- ============================================================

CREATE SCHEMA IF NOT EXISTS employee;
GRANT USAGE ON SCHEMA employee TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: departments
-- ────────────────────────────────────────────────────────────
CREATE TABLE employee.departments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    head_id     UUID,   -- FK to employees.id (set after employees table created)
    parent_id   UUID,   -- self-reference for nested departments
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, name)
);

CREATE INDEX departments_tenant_id_idx ON employee.departments (tenant_id);

ALTER TABLE employee.departments ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee.departments FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON employee.departments
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON employee.departments TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: positions
-- ────────────────────────────────────────────────────────────
CREATE TABLE employee.positions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    department_id UUID REFERENCES employee.departments(id) ON DELETE SET NULL,
    title         TEXT NOT NULL,
    level         TEXT NOT NULL DEFAULT 'staff',  -- junior|staff|senior|lead|manager|director
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX positions_tenant_id_idx ON employee.positions (tenant_id);
CREATE INDEX positions_department_id_idx ON employee.positions (department_id);

ALTER TABLE employee.positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee.positions FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON employee.positions
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON employee.positions TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: employees
-- ────────────────────────────────────────────────────────────
CREATE TYPE employee.employment_status AS ENUM (
    'active', 'inactive', 'terminated', 'on_leave'
);

CREATE TYPE employee.contract_type AS ENUM (
    'permanent', 'fixed_term', 'freelance', 'intern'
);

CREATE TYPE employee.gender AS ENUM (
    'male', 'female', 'unspecified'
);

CREATE TABLE employee.employees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,

    -- Identity
    employee_number TEXT NOT NULL,
    full_name       TEXT NOT NULL,
    email           TEXT NOT NULL,
    phone           TEXT,
    gender          employee.gender NOT NULL DEFAULT 'unspecified',
    birth_date      DATE,

    -- Indonesian legal IDs (stored encrypted via application-level encryption)
    national_id     TEXT,  -- NIK
    tax_id          TEXT,  -- NPWP

    -- Employment
    department_id   UUID REFERENCES employee.departments(id) ON DELETE SET NULL,
    position_id     UUID REFERENCES employee.positions(id) ON DELETE SET NULL,
    manager_id      UUID,   -- self-reference; FK added below
    status          employee.employment_status NOT NULL DEFAULT 'active',
    contract_type   employee.contract_type NOT NULL DEFAULT 'permanent',
    join_date       DATE NOT NULL,
    termination_date DATE,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, employee_number),
    UNIQUE (tenant_id, email)
);

ALTER TABLE employee.employees
    ADD CONSTRAINT employees_manager_id_fk
    FOREIGN KEY (manager_id) REFERENCES employee.employees(id) ON DELETE SET NULL;

ALTER TABLE employee.departments
    ADD CONSTRAINT departments_head_id_fk
    FOREIGN KEY (head_id) REFERENCES employee.employees(id) ON DELETE SET NULL;

CREATE INDEX employees_tenant_id_idx ON employee.employees (tenant_id);
CREATE INDEX employees_status_idx ON employee.employees (tenant_id, status);
CREATE INDEX employees_department_id_idx ON employee.employees (department_id);
CREATE INDEX employees_manager_id_idx ON employee.employees (manager_id);

-- Full-text search index on name and email
CREATE INDEX employees_fts_idx ON employee.employees
    USING gin(to_tsvector('english', full_name || ' ' || email));

ALTER TABLE employee.employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee.employees FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON employee.employees
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON employee.employees TO hris_app;
-- No DELETE — employees are terminated, never deleted (data integrity + audit).

-- ────────────────────────────────────────────────────────────
-- Table: processed_events (for NATS consumer idempotency)
-- ────────────────────────────────────────────────────────────
CREATE TABLE employee.processed_events (
    event_id     UUID PRIMARY KEY,
    event_type   TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX processed_events_event_type_idx ON employee.processed_events (event_type);

GRANT INSERT, SELECT ON employee.processed_events TO hris_app;
