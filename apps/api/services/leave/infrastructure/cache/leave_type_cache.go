package cache

import (
	"context"
	"time"

	"github.com/Briyantama/HRIS-Stery/internal/ttlcache"
	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
)

// LeaveTypeCacheRepository wraps a LeaveTypeRepository with in-memory caching.
// Leave types are ideal for caching: read-heavy, low-churn, small dataset.
type LeaveTypeCacheRepository struct {
	wrapped domain.LeaveTypeRepository
	lists   *ttlcache.Cache[[]*domain.LeaveType]
}

// NewLeaveTypeCacheRepository wraps a repository with caching (24-hour TTL).
func NewLeaveTypeCacheRepository(wrapped domain.LeaveTypeRepository) *LeaveTypeCacheRepository {
	return &LeaveTypeCacheRepository{
		wrapped: wrapped,
		lists:   ttlcache.New[[]*domain.LeaveType](24 * time.Hour),
	}
}

// Create persists a new leave type and invalidates the cache.
func (c *LeaveTypeCacheRepository) Create(ctx context.Context, leaveType *domain.LeaveType) error {
	err := c.wrapped.Create(ctx, leaveType)
	if err == nil {
		c.lists.Invalidate(leaveType.TenantID().String())
	}
	return err
}

// GetByID retrieves a leave type by ID (no caching for single lookups).
func (c *LeaveTypeCacheRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.LeaveTypeID) (*domain.LeaveType, error) {
	return c.wrapped.GetByID(ctx, tenantID, id)
}

// ListByTenant retrieves all leave types for a tenant with caching.
func (c *LeaveTypeCacheRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.LeaveType, error) {
	return c.lists.Load(tenantID.String(), func() ([]*domain.LeaveType, error) {
		return c.wrapped.ListByTenant(ctx, tenantID)
	})
}
