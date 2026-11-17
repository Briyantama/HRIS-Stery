-- +goose Up
-- Add login tracking columns to users table
ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS last_login_ip_addr TEXT;

-- +goose Down
-- Remove login tracking columns
ALTER TABLE auth.users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE auth.users DROP COLUMN IF EXISTS last_login_ip_addr;
