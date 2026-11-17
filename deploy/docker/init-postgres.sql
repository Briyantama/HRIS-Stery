-- ============================================================
-- PostgreSQL cluster bootstrap
-- Run once as the superuser by deploy/docker/init-postgres.sh (docker-entrypoint-initdb.d),
-- which passes the app password as the psql variable :app_password.
-- Manual run: psql -v app_password='...' -f init-postgres.sql
-- ============================================================

-- Application role (no BYPASSRLS — cannot see other tenants' data)
CREATE ROLE hris_app WITH LOGIN PASSWORD :'app_password'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;

-- Owner/migration role (BYPASSRLS — migrations and backoffice only)
ALTER ROLE :"POSTGRES_USER" BYPASSRLS;

-- Grant connect on the database
GRANT CONNECT ON DATABASE :"POSTGRES_DB" TO hris_app;

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- trigram search
CREATE EXTENSION IF NOT EXISTS "btree_gin";  -- GIN indexes for composite searches
