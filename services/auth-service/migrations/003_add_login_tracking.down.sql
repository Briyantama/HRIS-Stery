-- Remove login tracking columns
ALTER TABLE auth.users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE auth.users DROP COLUMN IF EXISTS last_login_ip_addr;
