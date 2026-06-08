package cache

import (
	"context"
	"sync"
	"time"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

type departmentCacheEntry struct {
	departments []*domain.Department
	expiresAt   time.Time
}

// DepartmentCacheRepository wraps a DepartmentRepository with in-memory caching.
// Departments are ideal for caching: read-heavy, low-churn, small dataset.
type DepartmentCacheRepository struct {
	wrapped domain.DepartmentRepository
	mu      sync.RWMutex
	cache   map[string]departmentCacheEntry
	ttl     time.Duration
}

// NewDepartmentCacheRepository wraps a repository with caching (24-hour TTL).
func NewDepartmentCacheRepository(wrapped domain.DepartmentRepository) *DepartmentCacheRepository {
	return &DepartmentCacheRepository{
		wrapped: wrapped,
		cache:   make(map[string]departmentCacheEntry),
		ttl:     24 * time.Hour,
	}
}

// Create persists a new department and invalidates the cache.
func (c *DepartmentCacheRepository) Create(ctx context.Context, department *domain.Department) error {
	err := c.wrapped.Create(ctx, department)
	if err == nil {
		c.invalidateListCache(department.TenantID().String())
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
		c.invalidateListCache(department.TenantID().String())
	}
	return err
}

// ListByTenant retrieves all departments for a tenant with caching.
func (c *DepartmentCacheRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Department, error) {
	tenantKey := tenantID.String()

	c.mu.RLock()
	if entry, ok := c.cache[tenantKey]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.departments, nil
	}
	c.mu.RUnlock()

	departments, err := c.wrapped.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[tenantKey] = departmentCacheEntry{
		departments: departments,
		expiresAt:   time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return departments, nil
}

// invalidateListCache clears the cached list for a tenant.
func (c *DepartmentCacheRepository) invalidateListCache(tenantKey string) {
	c.mu.Lock()
	delete(c.cache, tenantKey)
	c.mu.Unlock()
}
