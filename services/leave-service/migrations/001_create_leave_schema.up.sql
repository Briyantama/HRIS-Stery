-- Create leave schema
CREATE SCHEMA IF NOT EXISTS leave;

-- Create leave_types table
CREATE TABLE leave.leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    max_days_per_year NUMERIC(5, 2) NOT NULL,
    requires_document BOOLEAN NOT NULL DEFAULT false,
    is_paid BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index on tenant_id for RLS
CREATE INDEX idx_leave_types_tenant_id ON leave.leave_types(tenant_id);

-- Create unique index on tenant + code
CREATE UNIQUE INDEX idx_leave_types_tenant_code ON leave.leave_types(tenant_id, code) WHERE is_active = true;

-- Enable RLS on leave_types
ALTER TABLE leave.leave_types ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for leave_types
CREATE POLICY leave_types_tenant_isolation ON leave.leave_types
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- Create leave_requests table
CREATE TABLE leave.leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_count INTEGER NOT NULL CHECK (days_count > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED')),
    reason TEXT,
    rejection_reason TEXT,
    approved_by_id UUID,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create indexes on leave_requests
CREATE INDEX idx_leave_requests_tenant_id ON leave.leave_requests(tenant_id);
CREATE INDEX idx_leave_requests_employee_id ON leave.leave_requests(tenant_id, employee_id);
CREATE INDEX idx_leave_requests_status ON leave.leave_requests(tenant_id, status);
CREATE INDEX idx_leave_requests_dates ON leave.leave_requests(tenant_id, start_date, end_date);
CREATE INDEX idx_leave_requests_approved_by ON leave.leave_requests(tenant_id, approved_by_id);

-- Enable RLS on leave_requests
ALTER TABLE leave.leave_requests ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for leave_requests
CREATE POLICY leave_requests_tenant_isolation ON leave.leave_requests
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- Create leave_balances table
CREATE TABLE leave.leave_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL,
    year INTEGER NOT NULL,
    entitled_days NUMERIC(5, 2) NOT NULL CHECK (entitled_days >= 0),
    used_days NUMERIC(5, 2) NOT NULL DEFAULT 0 CHECK (used_days >= 0),
    pending_days NUMERIC(5, 2) NOT NULL DEFAULT 0 CHECK (pending_days >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, employee_id, leave_type_id, year)
);

-- Create indexes on leave_balances
CREATE INDEX idx_leave_balances_tenant_id ON leave.leave_balances(tenant_id);
CREATE INDEX idx_leave_balances_employee_id ON leave.leave_balances(tenant_id, employee_id);
CREATE INDEX idx_leave_balances_year ON leave.leave_balances(tenant_id, year);

-- Enable RLS on leave_balances
ALTER TABLE leave.leave_balances ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for leave_balances
CREATE POLICY leave_balances_tenant_isolation ON leave.leave_balances
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- Create processed_events table (NO RLS - internal idempotency table)
CREATE TABLE leave.processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index on processed_at for cleanup
CREATE INDEX idx_processed_events_processed_at ON leave.processed_events(processed_at);

-- Grant permissions to app role
GRANT ALL PRIVILEGES ON SCHEMA leave TO app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA leave TO app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA leave TO app;
