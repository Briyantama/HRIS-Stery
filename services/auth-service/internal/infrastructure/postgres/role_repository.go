package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleRepository implements domain.RoleRepository using PostgreSQL.
type RoleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository creates a new Postgres role repository.
func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

// Create persists a new role.
func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(role.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO auth.roles (id, tenant_id, name, description, is_system, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(ctx, query,
			role.ID().String(),
			role.TenantID().String(),
			role.Name(),
			role.Description(),
			role.IsSystem(),
			role.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a role by ID within a tenant.
func (r *RoleRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.RoleID) (*domain.Role, error) {
	var role *domain.Role
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, name, description, is_system, created_at
			FROM auth.roles
			WHERE id = $1
		`
		var roleID, tenantIDStr, name, description string
		var isSystem bool
		var createdAt time.Time
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(&roleID, &tenantIDStr, &name, &description, &isSystem, &createdAt)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("role not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query role: %w", rowErr)
		}
		role = domain.NewRole(
			domain.MustNewRoleID(roleID),
			domain.MustNewTenantID(tenantIDStr),
			name,
			description,
			isSystem,
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return role, nil
}

// GetByTenantAndName retrieves a role by tenant and name.
func (r *RoleRepository) GetByTenantAndName(ctx context.Context, tenantID domain.TenantID, name string) (*domain.Role, error) {
	var role *domain.Role
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, name, description, is_system, created_at
			FROM auth.roles
			WHERE tenant_id = $1 AND name = $2
		`
		var roleID, tenantIDStr, roleName, description string
		var isSystem bool
		var createdAt time.Time
		rowErr := tx.QueryRow(ctx, query, tenantID.String(), name).Scan(&roleID, &tenantIDStr, &roleName, &description, &isSystem, &createdAt)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("role not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query role by name: %w", rowErr)
		}
		role = domain.NewRole(
			domain.MustNewRoleID(roleID),
			domain.MustNewTenantID(tenantIDStr),
			roleName,
			description,
			isSystem,
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return role, nil
}

// ListByTenant retrieves all roles for a tenant.
func (r *RoleRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Role, error) {
	var roles []*domain.Role
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, name, description, is_system, created_at
			FROM auth.roles
			WHERE tenant_id = $1
		`
		rows, err := tx.Query(ctx, query, tenantID.String())
		if err != nil {
			return fmt.Errorf("query roles: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var roleID, tenantIDStr, name, description string
			var isSystem bool
			var createdAt time.Time
			if err := rows.Scan(&roleID, &tenantIDStr, &name, &description, &isSystem, &createdAt); err != nil {
				return fmt.Errorf("scan role: %w", err)
			}
			roles = append(roles, domain.NewRole(
				domain.MustNewRoleID(roleID),
				domain.MustNewTenantID(tenantIDStr),
				name,
				description,
				isSystem,
			))
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// Update persists changes to a role.
func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(role.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		result, err := tx.Exec(ctx,
			`UPDATE auth.roles SET name = $1, description = $2 WHERE id = $3`,
			role.Name(), role.Description(), role.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update role: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("role not found")
		}
		return nil
	})
}
