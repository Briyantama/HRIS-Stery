package cache

import (
	"context"
	"time"

	"github.com/Briyantama/HRIS-Stery/internal/ttlcache"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
)

// DepartmentCacheRepository wraps a DepartmentRepository with in-memory caching.
// Departments are ideal for caching: read-heavy, low-churn, small dataset.
type DepartmentCacheRepository struct {
	wrapped domain.DepartmentRepository
	lists   *ttlcache.Cache[[]*domain.Department]
}

// NewDepartmentCacheRepository wraps a repository with caching (24-hour TTL).
func NewDepartmentCacheRepository(wrapped domain.DepartmentRepository) *DepartmentCacheRepository {
	return &DepartmentCacheRepository{
		wrapped: wrapped,
		lists:   ttlcache.New[[]*domain.Department](24 * time.Hour),
	}
}

// Create persists a new department and invalidates the cache.
func (c *DepartmentCacheRepository) Create(ctx context.Context, department *domain.Department) error {
	err := c.wrapped.Create(ctx, department)
	if err == nil {
		c.lists.Invalidate(department.TenantID().String())
	}
	return err
}

// GetByID retrieves a department by ID (no caching for single lookups).
func (c *DepartmentCacheRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.DepartmentID) (*domain.Department, error) {
	return c.wrapped.GetByID(ctx, tenantID, id)
}

// Update updates a department and invalidates the cache.
func (c *DepartmentCacheRepository) Update(ctx context.Context, department *domain.Department) error {
	err := c.wrapped.Update(ctx, department)
	if err == nil {
		c.lists.Invalidate(department.TenantID().String())
	}
	return err
}

// ListByTenant retrieves all departments for a tenant with caching.
func (c *DepartmentCacheRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Department, error) {
	return c.lists.Load(tenantID.String(), func() ([]*domain.Department, error) {
		return c.wrapped.ListByTenant(ctx, tenantID)
	})
}
