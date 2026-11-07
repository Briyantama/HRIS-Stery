-- name: InsertRole :exec
INSERT INTO auth.roles (id, tenant_id, name, description, is_system, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetRoleByID :one
SELECT id, tenant_id, name, description, is_system, created_at
FROM auth.roles
WHERE id = $1;

-- name: GetRoleByTenantAndName :one
SELECT id, tenant_id, name, description, is_system, created_at
FROM auth.roles
WHERE tenant_id = $1 AND name = $2;

-- name: ListRolesByTenant :many
SELECT id, tenant_id, name, description, is_system, created_at
FROM auth.roles
WHERE tenant_id = $1;

-- name: UpdateRole :execresult
UPDATE auth.roles SET name = $1, description = $2 WHERE id = $3;
