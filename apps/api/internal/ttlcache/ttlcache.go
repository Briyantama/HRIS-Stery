// Package ttlcache is a small per-key in-memory cache with a fixed TTL.
package ttlcache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// Cache holds one value per string key until its TTL runs out.
type Cache[V any] struct {
	mu      sync.RWMutex
	entries map[string]entry[V]
	ttl     time.Duration
}

// New returns an empty cache whose entries live for ttl.
func New[V any](ttl time.Duration) *Cache[V] {
	return &Cache[V]{entries: make(map[string]entry[V]), ttl: ttl}
}

// Load returns the cached value for key, or calls fetch and caches the result.
// Fetch errors are returned and never cached.
func (c *Cache[V]) Load(key string, fetch func() (V, error)) (V, error) {
	c.mu.RLock()
	if e, ok := c.entries[key]; ok && time.Now().Before(e.expiresAt) {
		c.mu.RUnlock()
		return e.value, nil
	}
	c.mu.RUnlock()

	v, err := fetch()
	if err != nil {
		var zero V
		return zero, err
	}

	c.mu.Lock()
	c.entries[key] = entry[V]{value: v, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return v, nil
}

// Invalidate drops the cached value for key.
func (c *Cache[V]) Invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}
