package cache

import (
	"context"
	"sync"
	"time"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

type positionCacheEntry struct {
	positions []*domain.Position
	expiresAt time.Time
}

// PositionCacheRepository wraps a PositionRepository with in-memory caching.
// Positions are ideal for caching: read-heavy, low-churn, small dataset.
type PositionCacheRepository struct {
	wrapped domain.PositionRepository
	mu      sync.RWMutex
	cache   map[string]positionCacheEntry
	ttl     time.Duration
}

// NewPositionCacheRepository wraps a repository with caching (24-hour TTL).
func NewPositionCacheRepository(wrapped domain.PositionRepository) *PositionCacheRepository {
	return &PositionCacheRepository{
		wrapped: wrapped,
		cache:   make(map[string]positionCacheEntry),
		ttl:     24 * time.Hour,
	}
}

// Create persists a new position and invalidates the cache.
func (c *PositionCacheRepository) Create(ctx context.Context, position *domain.Position) error {
	err := c.wrapped.Create(ctx, position)
	if err == nil {
		c.invalidateListCache(position.TenantID().String())
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
		c.invalidateListCache(position.TenantID().String())
	}
	return err
}

// ListByTenant retrieves all positions for a tenant with caching.
func (c *PositionCacheRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Position, error) {
	tenantKey := tenantID.String()

	c.mu.RLock()
	if entry, ok := c.cache[tenantKey]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.positions, nil
	}
	c.mu.RUnlock()

	positions, err := c.wrapped.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[tenantKey] = positionCacheEntry{
		positions: positions,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return positions, nil
}

// invalidateListCache clears the cached list for a tenant.
func (c *PositionCacheRepository) invalidateListCache(tenantKey string) {
	c.mu.Lock()
	delete(c.cache, tenantKey)
	c.mu.Unlock()
}
