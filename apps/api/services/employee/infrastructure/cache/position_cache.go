package cache

import (
	"context"
	"time"

	"github.com/Briyantama/HRIS-Stery/internal/ttlcache"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
)

// PositionCacheRepository wraps a PositionRepository with in-memory caching.
// Positions are ideal for caching: read-heavy, low-churn, small dataset.
type PositionCacheRepository struct {
	wrapped domain.PositionRepository
	lists   *ttlcache.Cache[[]*domain.Position]
}

// NewPositionCacheRepository wraps a repository with caching (24-hour TTL).
func NewPositionCacheRepository(wrapped domain.PositionRepository) *PositionCacheRepository {
	return &PositionCacheRepository{
		wrapped: wrapped,
		lists:   ttlcache.New[[]*domain.Position](24 * time.Hour),
	}
}

// Create persists a new position and invalidates the cache.
func (c *PositionCacheRepository) Create(ctx context.Context, position *domain.Position) error {
	err := c.wrapped.Create(ctx, position)
	if err == nil {
		c.lists.Invalidate(position.TenantID().String())
	}
	return err
}

// GetByID retrieves a position by ID (no caching for single lookups).
func (c *PositionCacheRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.PositionID) (*domain.Position, error) {
	return c.wrapped.GetByID(ctx, tenantID, id)
}

// Update updates a position and invalidates the cache.
func (c *PositionCacheRepository) Update(ctx context.Context, position *domain.Position) error {
	err := c.wrapped.Update(ctx, position)
	if err == nil {
		c.lists.Invalidate(position.TenantID().String())
	}
	return err
}

// ListByTenant retrieves all positions for a tenant with caching.
func (c *PositionCacheRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Position, error) {
	return c.lists.Load(tenantID.String(), func() ([]*domain.Position, error) {
		return c.wrapped.ListByTenant(ctx, tenantID)
	})
}
