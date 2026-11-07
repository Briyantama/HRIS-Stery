-- name: InsertEntry :exec
INSERT INTO audit.entries (
    entry_id, tenant_id, actor_id, action, resource_type,
    resource_id, description, success, error_message, changes, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetEntryByID :one
SELECT entry_id, tenant_id, actor_id, action, resource_type,
       resource_id, description, success, error_message, changes, created_at
FROM audit.entries
WHERE entry_id = $1 AND tenant_id = $2;

-- name: ListEntriesByTenant :many
SELECT entry_id, tenant_id, actor_id, action, resource_type,
       resource_id, description, success, error_message, changes, created_at
FROM audit.entries
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
