package gocachex

import (
	"context"
	"time"
)

// Cache represents a unified, type-safe caching interface using Generics.
// Any backend implementation (memory, redis, memcached) must satisfy this interface.
type Cache[T any] interface {
	// Get retrieves a value from the cache.
	Get(ctx context.Context, key string) (T, error)

	// Set stores a value in the cache with the given TTL.
	Set(ctx context.Context, key string, value T, ttl time.Duration) error

	// Delete removes a value from the cache.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Clear purges all keys from the cache.
	Clear(ctx context.Context) error

	// Close closes the cache backend connection and cleans up resources.
	Close() error
}
