package postgres

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRepository implements domain.TenantRepository using PostgreSQL.
type TenantRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRepository creates a new Postgres tenant repository.
func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// Create persists a new tenant.
// Note: Tenants are not tenant-scoped, so no RLS context needed.
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO auth.tenants (id, slug, company_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		tenant.ID().String(),
		tenant.Slug().String(),
		tenant.CompanyName(),
		tenant.IsActive(),
		tenant.CreatedAt(),
		tenant.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by ID.
func (r *TenantRepository) GetByID(ctx context.Context, id domain.TenantID) (*domain.Tenant, error) {
	query := `
		SELECT id, slug, company_name, is_active, created_at, updated_at
		FROM auth.tenants
		WHERE id = $1
	`
	var (
		tenantID    string
		slug        string
		companyName string
		isActive    bool
	)
	err := r.pool.QueryRow(ctx, query, id.String()).Scan(
		&tenantID,
		&slug,
		&companyName,
		&isActive,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}

	tenant := domain.NewTenant(
		domain.MustNewTenantID(tenantID),
		domain.MustNewTenantSlug(slug),
		companyName,
	)
	if !isActive {
		tenant.Deactivate()
	}
	return tenant, nil
}

// GetBySlug retrieves a tenant by its URL slug.
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT id, slug, company_name, is_active, created_at, updated_at
		FROM auth.tenants
		WHERE slug = $1
	`
	var (
		tenantID    string
		slugValue   string
		companyName string
		isActive    bool
	)
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&tenantID,
		&slugValue,
		&companyName,
		&isActive,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant by slug: %w", err)
	}

	tenant := domain.NewTenant(
		domain.MustNewTenantID(tenantID),
		domain.MustNewTenantSlug(slugValue),
		companyName,
	)
	if !isActive {
		tenant.Deactivate()
	}
	return tenant, nil
}

// Update persists changes to a tenant.
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE auth.tenants
		SET company_name = $1, is_active = $2, updated_at = $3
		WHERE id = $4
	`
	result, err := r.pool.Exec(ctx, query,
		tenant.CompanyName(),
		tenant.IsActive(),
		tenant.UpdatedAt(),
		tenant.ID().String(),
	)
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
	query := `DELETE FROM auth.tenants WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found")
	}
	return nil
}
