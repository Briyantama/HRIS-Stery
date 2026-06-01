package postgres

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionRepository implements domain.PermissionRepository using PostgreSQL.
type PermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPermissionRepository creates a new Postgres permission repository.
func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// Create persists a new permission (global, not tenant-scoped).
func (r *PermissionRepository) Create(ctx context.Context, name, description string) error {
	query := `
		INSERT INTO auth.permissions (name, description, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (name) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, name, description)
	if err != nil {
		return fmt.Errorf("insert permission: %w", err)
	}
	return nil
}

// GetByName retrieves a permission by name.
func (r *PermissionRepository) GetByName(ctx context.Context, name string) (domain.Permission, error) {
	query := `SELECT name FROM auth.permissions WHERE name = $1`
	var permName string
	err := r.pool.QueryRow(ctx, query, name).Scan(&permName)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("permission not found")
	}
	if err != nil {
		return "", fmt.Errorf("query permission: %w", err)
	}
	return domain.Permission(permName), nil
}

// List retrieves all permissions.
func (r *PermissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	query := `SELECT name FROM auth.permissions ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, domain.Permission(name))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate permissions: %w", err)
	}
	return perms, nil
}

// AssignToRole grants a permission to a role.
func (r *PermissionRepository) AssignToRole(ctx context.Context, roleID domain.RoleID, permission domain.Permission) error {
	// First get the permission ID
	permQuery := `SELECT id FROM auth.permissions WHERE name = $1`
	var permID string
	err := r.pool.QueryRow(ctx, permQuery, permission.String()).Scan(&permID)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("permission not found")
	}
	if err != nil {
		return fmt.Errorf("query permission: %w", err)
	}

	query := `
		INSERT INTO auth.role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err = r.pool.Exec(ctx, query, roleID.String(), permID)
	if err != nil {
		return fmt.Errorf("assign permission to role: %w", err)
	}
	return nil
}

// RemoveFromRole revokes a permission from a role.
func (r *PermissionRepository) RemoveFromRole(ctx context.Context, roleID domain.RoleID, permission domain.Permission) error {
	permQuery := `SELECT id FROM auth.permissions WHERE name = $1`
	var permID string
	err := r.pool.QueryRow(ctx, permQuery, permission.String()).Scan(&permID)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("permission not found")
	}
	if err != nil {
		return fmt.Errorf("query permission: %w", err)
	}

	query := `DELETE FROM auth.role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err = r.pool.Exec(ctx, query, roleID.String(), permID)
	if err != nil {
		return fmt.Errorf("remove permission from role: %w", err)
	}
	return nil
}

// GetForRole retrieves all permissions assigned to a role.
func (r *PermissionRepository) GetForRole(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	query := `
		SELECT p.name
		FROM auth.permissions p
		JOIN auth.role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`
	rows, err := r.pool.Query(ctx, query, roleID.String())
	if err != nil {
		return nil, fmt.Errorf("query role permissions: %w", err)
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, domain.Permission(name))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate permissions: %w", err)
	}
	return perms, nil
}
