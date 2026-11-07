-- name: InsertTenant :exec
INSERT INTO auth.tenants (id, slug, company_name, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTenantByID :one
SELECT id, slug, company_name, is_active, created_at, updated_at
FROM auth.tenants
WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT id, slug, company_name, is_active, created_at, updated_at
FROM auth.tenants
WHERE slug = $1;

-- name: UpdateTenant :execresult
UPDATE auth.tenants
SET company_name = $1, is_active = $2, updated_at = $3
WHERE id = $4;

-- name: DeleteTenant :execresult
DELETE FROM auth.tenants WHERE id = $1;
