package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	sqlcdb "github.com/Briyantama/HRIS-Stery/services/auth/infrastructure/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"uuid"
)

// TenantRepository implements domain.TenantRepository using PostgreSQL.
// Tenants are not tenant-scoped, so queries run directly on the pool without RLS.
type TenantRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRepository creates a new Postgres tenant repository.
func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// Create persists a new tenant.
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	q := sqlcdb.New(r.pool)
	err := q.InsertTenant(ctx, sqlcdb.InsertTenantParams{
		ID:          uuid.MustParse(tenant.ID().String()),
		Slug:        tenant.Slug().String(),
		CompanyName: tenant.CompanyName(),
		IsActive:    tenant.IsActive(),
		CreatedAt:   tenant.CreatedAt(),
		UpdatedAt:   tenant.UpdatedAt(),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("insert tenant: %w", domain.ErrConflict)
		}
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by ID.
func (r *TenantRepository) GetByID(ctx context.Context, id domain.TenantID) (*domain.Tenant, error) {
	q := sqlcdb.New(r.pool)
	row, err := q.GetTenantByID(ctx, uuid.MustParse(id.String()))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("tenant not found: %w", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}
	return buildTenant(row.ID, row.Slug, row.CompanyName, row.IsActive), nil
}

// GetBySlug retrieves a tenant by its URL slug.
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	q := sqlcdb.New(r.pool)
	row, err := q.GetTenantBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("tenant not found: %w", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant by slug: %w", err)
	}
	return buildTenant(row.ID, row.Slug, row.CompanyName, row.IsActive), nil
}

// Update persists changes to a tenant.
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	q := sqlcdb.New(r.pool)
	result, err := q.UpdateTenant(ctx, sqlcdb.UpdateTenantParams{
		CompanyName: tenant.CompanyName(),
		IsActive:    tenant.IsActive(),
		UpdatedAt:   tenant.UpdatedAt(),
		ID:          uuid.MustParse(tenant.ID().String()),
	})
	if err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found")
	}
	return nil
}

// Delete removes a tenant and all its data (cascading FK constraints).
func (r *TenantRepository) Delete(ctx context.Context, id domain.TenantID) error {
	q := sqlcdb.New(r.pool)
	result, err := q.DeleteTenant(ctx, uuid.MustParse(id.String()))
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found")
	}
	return nil
}

func buildTenant(id uuid.UUID, slug, companyName string, isActive bool) *domain.Tenant {
	tenant := domain.NewTenant(
		domain.MustNewTenantID(id.String()),
		domain.MustNewTenantSlug(slug),
		companyName,
	)
	if !isActive {
		tenant.Deactivate()
	}
	return tenant
}
