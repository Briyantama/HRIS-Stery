-- ============================================================
-- Migration: 001_create_auth_schema
-- Service:   auth-service
-- Schema:    auth
-- ============================================================
-- Creates the auth schema with full RLS enforcement.
-- Run with hris_admin role (BYPASSRLS). Application role is hris_app.
-- ============================================================

-- Roles (idempotent — run in cluster bootstrap, not per-migration)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'hris_app') THEN
    CREATE ROLE hris_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'hris_admin') THEN
    CREATE ROLE hris_admin LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT BYPASSRLS;
  END IF;
END $$;

-- Schema
CREATE SCHEMA IF NOT EXISTS auth;
GRANT USAGE ON SCHEMA auth TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: tenants
-- No RLS on tenants — tenant lookup happens before context is set.
-- Access restricted to hris_admin for writes; hris_app for reads.
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.tenants (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         TEXT NOT NULL UNIQUE,         -- URL-safe unique identifier
    company_name TEXT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    plan         TEXT NOT NULL DEFAULT 'free', -- subscription plan
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX tenants_slug_idx ON auth.tenants (slug);

GRANT SELECT ON auth.tenants TO hris_app;
GRANT INSERT, UPDATE ON auth.tenants TO hris_admin;

-- ────────────────────────────────────────────────────────────
-- Table: users
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    full_name       TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    email_verified  BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, email)  -- email is unique per tenant, not globally
);

CREATE INDEX users_tenant_id_idx ON auth.users (tenant_id);
CREATE INDEX users_email_idx ON auth.users (email);

ALTER TABLE auth.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.users FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON auth.users
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON auth.users TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: roles
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,           -- e.g. "hr_admin", "manager", "employee"
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT false,  -- system roles cannot be deleted
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, name)
);

CREATE INDEX roles_tenant_id_idx ON auth.roles (tenant_id);

ALTER TABLE auth.roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.roles FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON auth.roles
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON auth.roles TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: permissions
-- Permissions are global (not tenant-scoped) — they are the same for all tenants.
-- Only system can modify; hris_app can read.
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,    -- e.g. "employee:read", "leave:approve"
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

GRANT SELECT ON auth.permissions TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: role_permissions
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.role_permissions (
    role_id       UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

GRANT SELECT, INSERT, DELETE ON auth.role_permissions TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: user_roles
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.user_roles (
    user_id    UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    granted_by UUID,  -- user_id of admin who granted this
    PRIMARY KEY (user_id, role_id)
);

GRANT SELECT, INSERT, DELETE ON auth.user_roles TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: refresh_tokens
-- Stored in Redis as the primary store; this table is a persistent fallback
-- and enables token listing/revocation for security audit.
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,  -- SHA-256 of the raw refresh token
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,           -- NULL = active
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_agent  TEXT,
    ip_address  TEXT
);

CREATE INDEX refresh_tokens_tenant_id_idx ON auth.refresh_tokens (tenant_id);
CREATE INDEX refresh_tokens_user_id_idx ON auth.refresh_tokens (user_id);
CREATE INDEX refresh_tokens_token_hash_idx ON auth.refresh_tokens (token_hash);

ALTER TABLE auth.refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.refresh_tokens FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON auth.refresh_tokens
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON auth.refresh_tokens TO hris_app;

-- ────────────────────────────────────────────────────────────
-- Table: login_attempts  (rate limiting + security monitoring)
-- No RLS — lookup by email before tenant context is known.
-- ────────────────────────────────────────────────────────────
CREATE TABLE auth.login_attempts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL,
    ip_address  TEXT,
    success     BOOLEAN NOT NULL,
    tenant_slug TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX login_attempts_email_idx ON auth.login_attempts (email, attempted_at DESC);
CREATE INDEX login_attempts_ip_idx ON auth.login_attempts (ip_address, attempted_at DESC);

GRANT INSERT ON auth.login_attempts TO hris_app;
GRANT SELECT ON auth.login_attempts TO hris_admin;

-- ────────────────────────────────────────────────────────────
-- Seed: default permissions
-- ────────────────────────────────────────────────────────────
INSERT INTO auth.permissions (name, description) VALUES
    ('employee:read',        'View employee records'),
    ('employee:write',       'Create and update employee records'),
    ('employee:delete',      'Terminate employees'),
    ('attendance:read',      'View attendance records'),
    ('attendance:write',     'Check in/out and override attendance'),
    ('leave:read',           'View leave requests'),
    ('leave:write',          'Submit leave requests'),
    ('leave:approve',        'Approve or reject leave requests'),
    ('notification:read',    'View notifications'),
    ('audit:read',           'View audit log'),
    ('settings:write',       'Manage tenant settings')
ON CONFLICT (name) DO NOTHING;
