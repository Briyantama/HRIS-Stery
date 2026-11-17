-- +goose Up
-- Concurrent "Unassigned" creation by the user-registered consumer must not produce duplicate positions.
CREATE UNIQUE INDEX positions_tenant_title_key ON employee.positions (tenant_id, title);

-- +goose Down
DROP INDEX employee.positions_tenant_title_key;
