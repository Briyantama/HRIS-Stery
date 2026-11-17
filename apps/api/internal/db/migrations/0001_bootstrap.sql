-- +goose Up
-- Schemas and roles every later migration relies on. Roles are NOLOGIN here; the deployment
-- (deploy/docker/init-postgres.sql) creates hris_app with LOGIN first, which this block then leaves alone.
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS employee;
CREATE SCHEMA IF NOT EXISTS attendance;
CREATE SCHEMA IF NOT EXISTS leave;
CREATE SCHEMA IF NOT EXISTS notification;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS document;

-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hris_app') THEN
    CREATE ROLE hris_app NOLOGIN;
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hris_admin') THEN
    CREATE ROLE hris_admin NOLOGIN BYPASSRLS;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- Roles are cluster-wide and may be shared with other databases, so they are kept.
DROP SCHEMA IF EXISTS document;
DROP SCHEMA IF EXISTS audit;
DROP SCHEMA IF EXISTS notification;
DROP SCHEMA IF EXISTS leave;
DROP SCHEMA IF EXISTS attendance;
DROP SCHEMA IF EXISTS employee;
DROP SCHEMA IF EXISTS auth;
