-- ============================================================
-- Rollback: 001_create_auth_schema
-- ============================================================
-- Drops all tables and schema created by the up migration.
-- Order matters: child tables before parents (FK constraints).
-- ============================================================

DROP TABLE IF EXISTS auth.login_attempts;
DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.user_roles;
DROP TABLE IF EXISTS auth.role_permissions;
DROP TABLE IF EXISTS auth.permissions;
DROP TABLE IF EXISTS auth.roles;
DROP TABLE IF EXISTS auth.users;
DROP TABLE IF EXISTS auth.tenants;

DROP SCHEMA IF EXISTS auth;
