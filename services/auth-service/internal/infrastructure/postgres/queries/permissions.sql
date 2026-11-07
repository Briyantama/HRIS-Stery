-- name: InsertPermission :exec
INSERT INTO auth.permissions (name, description, created_at)
VALUES ($1, $2, NOW())
ON CONFLICT (name) DO NOTHING;

-- name: GetPermissionByName :one
SELECT name FROM auth.permissions WHERE name = $1;

-- name: ListPermissions :many
SELECT name FROM auth.permissions ORDER BY name;

-- name: GetPermissionIDByName :one
SELECT id FROM auth.permissions WHERE name = $1;

-- name: InsertRolePermission :exec
INSERT INTO auth.role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: DeleteRolePermission :exec
DELETE FROM auth.role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: GetPermissionsForRole :many
SELECT p.name
FROM auth.permissions p
JOIN auth.role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.name;

-- name: GetPermissionsForRoles :many
SELECT DISTINCT p.name
FROM auth.permissions p
JOIN auth.role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = ANY($1::uuid[])
ORDER BY p.name;
