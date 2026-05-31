-- ============================================================
-- PostgreSQL cluster bootstrap
-- Run once by the hris_admin superuser via docker-entrypoint-initdb.d
-- ============================================================

-- Application role (no BYPASSRLS — cannot see other tenants' data)
CREATE ROLE hris_app WITH LOGIN PASSWORD 'hris_app_secret'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;

-- Admin role (BYPASSRLS — for migrations and backoffice only)
-- hris_admin was created as POSTGRES_USER — grant BYPASSRLS
ALTER ROLE hris_admin BYPASSRLS;

-- Grant connect on the database
GRANT CONNECT ON DATABASE hris_db TO hris_app;

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- trigram search
CREATE EXTENSION IF NOT EXISTS "btree_gin";  -- GIN indexes for composite searches
