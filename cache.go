package gocachex

import (
	"context"
	"time"
)

// Cache represents a unified, type-safe caching interface using Generics.
// Any backend implementation (memory, redis, memcached) must satisfy this interface.
//
// By using 1.18+ Generics, we skip interface{} casting, improving
// developer experience and achieving zero-allocation returns for in-memory layers.
type Cache[T any] interface {
	// Set stores a value in the cache with the given TTL.
	// A TTL of 0 means the item never expires.
	Set(ctx context.Context, key string, value T, ttl time.Duration) error

	// Get retrieves a value from the cache.
	Get(ctx context.Context, key string) (T, error)

	// Delete removes a value from the cache.
	Delete(ctx context.Context, key string) error
}
