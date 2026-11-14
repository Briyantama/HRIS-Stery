-- +goose Up
-- Tenant self-registration inserts into auth.tenants as hris_app.
GRANT INSERT ON auth.tenants TO hris_app;

-- +goose Down
REVOKE INSERT ON auth.tenants FROM hris_app;
