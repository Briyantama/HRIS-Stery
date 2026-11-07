package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	sqlcdb "github.com/hris-stery/hris-stery/services/auth-service/internal/infrastructure/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionRepository implements domain.PermissionRepository using PostgreSQL.
// Permissions are global (not tenant-scoped), so queries run directly on the pool.
type PermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPermissionRepository creates a new Postgres permission repository.
func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// Create persists a new permission (global, not tenant-scoped).
func (r *PermissionRepository) Create(ctx context.Context, name, description string) error {
	q := sqlcdb.New(r.pool)
	err := q.InsertPermission(ctx, sqlcdb.InsertPermissionParams{
		Name:        name,
		Description: &description,
	})
	if err != nil {
		return fmt.Errorf("insert permission: %w", err)
	}
	return nil
}

// GetByName retrieves a permission by name.
func (r *PermissionRepository) GetByName(ctx context.Context, name string) (domain.Permission, error) {
	q := sqlcdb.New(r.pool)
	permName, err := q.GetPermissionByName(ctx, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("permission not found")
	}
	if err != nil {
		return "", fmt.Errorf("query permission: %w", err)
	}
	return domain.Permission(permName), nil
}

// List retrieves all permissions.
func (r *PermissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	q := sqlcdb.New(r.pool)
	names, err := q.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}
	perms := make([]domain.Permission, len(names))
	for i, n := range names {
		perms[i] = domain.Permission(n)
	}
	return perms, nil
}

// AssignToRole grants a permission to a role.
func (r *PermissionRepository) AssignToRole(ctx context.Context, roleID domain.RoleID, permission domain.Permission) error {
	q := sqlcdb.New(r.pool)
	permID, err := q.GetPermissionIDByName(ctx, permission.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("permission not found")
	}
	if err != nil {
		return fmt.Errorf("query permission: %w", err)
	}
	err = q.InsertRolePermission(ctx, sqlcdb.InsertRolePermissionParams{
		RoleID:       uuid.MustParse(roleID.String()),
		PermissionID: permID,
	})
	if err != nil {
		return fmt.Errorf("assign permission to role: %w", err)
	}
	return nil
}

// RemoveFromRole revokes a permission from a role.
func (r *PermissionRepository) RemoveFromRole(ctx context.Context, roleID domain.RoleID, permission domain.Permission) error {
	q := sqlcdb.New(r.pool)
	permID, err := q.GetPermissionIDByName(ctx, permission.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("permission not found")
	}
	if err != nil {
		return fmt.Errorf("query permission: %w", err)
	}
	err = q.DeleteRolePermission(ctx, sqlcdb.DeleteRolePermissionParams{
		RoleID:       uuid.MustParse(roleID.String()),
		PermissionID: permID,
	})
	if err != nil {
		return fmt.Errorf("remove permission from role: %w", err)
	}
	return nil
}

// GetForRole retrieves all permissions assigned to a role.
func (r *PermissionRepository) GetForRole(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	q := sqlcdb.New(r.pool)
	names, err := q.GetPermissionsForRole(ctx, uuid.MustParse(roleID.String()))
	if err != nil {
		return nil, fmt.Errorf("query role permissions: %w", err)
	}
	perms := make([]domain.Permission, len(names))
	for i, n := range names {
		perms[i] = domain.Permission(n)
	}
	return perms, nil
}

// GetForRoles retrieves all unique permissions assigned to multiple roles in a single query.
func (r *PermissionRepository) GetForRoles(ctx context.Context, roleIDs []domain.RoleID) ([]domain.Permission, error) {
	if len(roleIDs) == 0 {
		return []domain.Permission{}, nil
	}
	q := sqlcdb.New(r.pool)
	uuids := make([]uuid.UUID, len(roleIDs))
	for i, id := range roleIDs {
		uuids[i] = uuid.MustParse(id.String())
	}
	names, err := q.GetPermissionsForRoles(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("query permissions for roles: %w", err)
	}
	perms := make([]domain.Permission, len(names))
	for i, n := range names {
		perms[i] = domain.Permission(n)
	}
	return perms, nil
}
