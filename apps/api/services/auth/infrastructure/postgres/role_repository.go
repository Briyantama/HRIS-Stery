package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	sqlcdb "github.com/Briyantama/HRIS-Stery/services/auth/infrastructure/postgres/sqlc"
	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"uuid"
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
	return pgtenant.WithTenantTx(ctx, r.pool, role.TenantID().String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		desc := role.Description()
		err := q.InsertRole(ctx, sqlcdb.InsertRoleParams{
			ID:          uuid.MustParse(role.ID().String()),
			TenantID:    uuid.MustParse(role.TenantID().String()),
			Name:        role.Name(),
			Description: &desc,
			IsSystem:    role.IsSystem(),
			CreatedAt:   role.CreatedAt(),
		})
		if err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a role by ID within a tenant.
func (r *RoleRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.RoleID) (*domain.Role, error) {
	var role *domain.Role
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		row, rowErr := q.GetRoleByID(ctx, uuid.MustParse(id.String()))
		if errors.Is(rowErr, pgx.ErrNoRows) {
			return fmt.Errorf("role not found: %w", domain.ErrNotFound)
		}
		if rowErr != nil {
			return fmt.Errorf("query role: %w", rowErr)
		}
		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}
		role = domain.NewRole(
			domain.MustNewRoleID(row.ID.String()),
			domain.MustNewTenantID(row.TenantID.String()),
			row.Name,
			desc,
			row.IsSystem,
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
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		row, rowErr := q.GetRoleByTenantAndName(ctx, sqlcdb.GetRoleByTenantAndNameParams{
			TenantID: uuid.MustParse(tenantID.String()),
			Name:     name,
		})
		if errors.Is(rowErr, pgx.ErrNoRows) {
			return fmt.Errorf("role not found: %w", domain.ErrNotFound)
		}
		if rowErr != nil {
			return fmt.Errorf("query role by name: %w", rowErr)
		}
		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}
		role = domain.NewRole(
			domain.MustNewRoleID(row.ID.String()),
			domain.MustNewTenantID(row.TenantID.String()),
			row.Name,
			desc,
			row.IsSystem,
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
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		rows, err := q.ListRolesByTenant(ctx, uuid.MustParse(tenantID.String()))
		if err != nil {
			return fmt.Errorf("query roles: %w", err)
		}
		for _, row := range rows {
			desc := ""
			if row.Description != nil {
				desc = *row.Description
			}
			roles = append(roles, domain.NewRole(
				domain.MustNewRoleID(row.ID.String()),
				domain.MustNewTenantID(row.TenantID.String()),
				row.Name,
				desc,
				row.IsSystem,
			))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// Update persists changes to a role.
func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return pgtenant.WithTenantTx(ctx, r.pool, role.TenantID().String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		desc := role.Description()
		result, err := q.UpdateRole(ctx, sqlcdb.UpdateRoleParams{
			Name:        role.Name(),
			Description: &desc,
			ID:          uuid.MustParse(role.ID().String()),
		})
		if err != nil {
			return fmt.Errorf("update role: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("role not found: %w", domain.ErrNotFound)
		}
		return nil
	})
}
