package memory

import (
	"context"
	"sync"
	"time"

	"github.com/chmenegatti/gocachex"
)

// Item represents a cached item with an expiration time.
type Item[T any] struct {
	Value      T
	Expiration int64
}

// MemoryCache is a thread-safe, generic in-memory cache implementation.
type MemoryCache[T any] struct {
	mu    sync.RWMutex
	items map[string]Item[T]

	// done is used to signal the cleanup goroutine to stop
	done chan struct{}
}

// Option configures the MemoryCache.
type Option[T any] func(*MemoryCache[T])

// WithCleanupInterval configures the interval at which expired items are automatically removed.
func WithCleanupInterval[T any](interval time.Duration) Option[T] {
	return func(c *MemoryCache[T]) {
		if interval > 0 {
			go c.startCleanupTimer(interval)
		}
	}
}

// New creates a new in-memory cache.
func New[T any](options ...Option[T]) *MemoryCache[T] {
	c := &MemoryCache[T]{
		items: make(map[string]Item[T]),
		done:  make(chan struct{}),
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// Get retrieves a value from the cache.
func (c *MemoryCache[T]) Get(ctx context.Context, key string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, gocachex.ErrCacheMiss
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		var zero T
		return zero, gocachex.ErrCacheMiss
	}

	return item.Value, nil
}

// Set stores a value in the cache with the given TTL.
// A TTL of 0 means the item never expires.
func (c *MemoryCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item[T]{
		Value:      value,
		Expiration: expiration,
	}
	return nil
}

// Delete removes a value from the cache.
func (c *MemoryCache[T]) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

// Exists checks if a key exists in the cache and has not expired.
func (c *MemoryCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return false, nil
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return false, nil
	}
	return true, nil
}

// Clear purges all keys from the cache.
func (c *MemoryCache[T]) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item[T])
	return nil
}

// Close stops the cleanup goroutine (if running).
func (c *MemoryCache[T]) Close() error {
	close(c.done)
	return nil
}

// startCleanupTimer periodically purges expired items.
func (c *MemoryCache[T]) startCleanupTimer(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

// cleanup removes all expired items from the cache.
func (c *MemoryCache[T]) cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, key)
		}
	}
}

// compiler check to ensure MemoryCache implements gocachex.Cache
var _ gocachex.Cache[string] = (*MemoryCache[string])(nil)
