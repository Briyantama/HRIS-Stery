package cache

import (
	"context"
	"sync"
	"time"

	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

type leaveTypeCacheEntry struct {
	types     []*domain.LeaveType
	expiresAt time.Time
}

// LeaveTypeCacheRepository wraps a LeaveTypeRepository with in-memory caching.
// Leave types are ideal for caching: read-heavy, low-churn, small dataset.
type LeaveTypeCacheRepository struct {
	wrapped domain.LeaveTypeRepository
	mu      sync.RWMutex
	cache   map[string]leaveTypeCacheEntry
	ttl     time.Duration
}

// NewLeaveTypeCacheRepository wraps a repository with caching (24-hour TTL).
func NewLeaveTypeCacheRepository(wrapped domain.LeaveTypeRepository) *LeaveTypeCacheRepository {
	return &LeaveTypeCacheRepository{
		wrapped: wrapped,
		cache:   make(map[string]leaveTypeCacheEntry),
		ttl:     24 * time.Hour,
	}
}

// Create persists a new leave type and invalidates the cache.
func (c *LeaveTypeCacheRepository) Create(ctx context.Context, leaveType *domain.LeaveType) error {
	err := c.wrapped.Create(ctx, leaveType)
	if err == nil {
		c.invalidateListCache(leaveType.TenantID().String())
	}
	return err
}

// GetByID retrieves a leave type by ID (no caching for single lookups).
func (c *LeaveTypeCacheRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.LeaveTypeID) (*domain.LeaveType, error) {
	return c.wrapped.GetByID(ctx, tenantID, id)
}

// ListByTenant retrieves all leave types for a tenant with caching.
func (c *LeaveTypeCacheRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.LeaveType, error) {
	tenantKey := tenantID.String()

	c.mu.RLock()
	if entry, ok := c.cache[tenantKey]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.types, nil
	}
	c.mu.RUnlock()

	types, err := c.wrapped.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[tenantKey] = leaveTypeCacheEntry{
		types:     types,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return types, nil
}

// invalidateListCache clears the cached list for a tenant.
func (c *LeaveTypeCacheRepository) invalidateListCache(tenantKey string) {
	c.mu.Lock()
	delete(c.cache, tenantKey)
	c.mu.Unlock()
}
