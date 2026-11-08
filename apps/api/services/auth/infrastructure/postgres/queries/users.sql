-- name: InsertUser :exec
INSERT INTO auth.users (id, tenant_id, email, password_hash, full_name, is_active, email_verified, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetUserByID :one
SELECT id, tenant_id, email, password_hash, full_name, is_active, email_verified, last_login_at, last_login_ip_addr, created_at, updated_at
FROM auth.users
WHERE id = $1;

-- name: GetUserByTenantAndEmail :one
SELECT id, tenant_id, email, password_hash, full_name, is_active, email_verified, last_login_at, last_login_ip_addr, created_at, updated_at
FROM auth.users
WHERE tenant_id = $1 AND email = $2;

-- name: UpdateUser :execresult
UPDATE auth.users
SET full_name = $1, is_active = $2, email_verified = $3, last_login_at = $4, last_login_ip_addr = $5, updated_at = $6
WHERE id = $7;

-- name: DeleteUser :execresult
DELETE FROM auth.users WHERE id = $1;

-- name: GetUserRoles :many
SELECT r.id, r.tenant_id, r.name, r.description, r.is_system, r.created_at
FROM auth.roles r
JOIN auth.user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: DeleteUserRoles :exec
DELETE FROM auth.user_roles WHERE user_id = $1;

-- name: InsertUserRole :exec
INSERT INTO auth.user_roles (user_id, role_id, granted_at) VALUES ($1, $2, NOW());
