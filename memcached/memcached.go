package memcached

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/chmenegatti/gocachex"
)

// MemcachedCache is a Memcached-backed, generic cache implementation.
type MemcachedCache[T any] struct {
	client *memcache.Client
}

// New creates a new Memcached cache backend.
func New[T any](client *memcache.Client) *MemcachedCache[T] {
	return &MemcachedCache[T]{
		client: client,
	}
}

// Get retrieves a value from the cache.
func (c *MemcachedCache[T]) Get(ctx context.Context, key string) (T, error) {
	item, err := c.client.Get(key)
	if err == memcache.ErrCacheMiss {
		var zero T
		return zero, gocachex.ErrCacheMiss
	} else if err != nil {
		var zero T
		return zero, err
	}

	var parsed T
	if err := json.Unmarshal(item.Value, &parsed); err != nil {
		var zero T
		return zero, err
	}

	return parsed, nil
}

// Set stores a value in the cache with the given TTL.
// Memcached ignores contexts.
func (c *MemcachedCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	expiration := int32(ttl.Seconds())

	return c.client.Set(&memcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: expiration,
	})
}

// Delete removes a value from the cache.
func (c *MemcachedCache[T]) Delete(ctx context.Context, key string) error {
	err := c.client.Delete(key)
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

// Exists checks if a key exists in the cache.
// Memcached doesn't have an EXISTS command, so we do a Get.
func (c *MemcachedCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Clear purges all keys from the cache.
func (c *MemcachedCache[T]) Clear(ctx context.Context) error {
	return c.client.FlushAll()
}

// Close closes the Memcached connections.
// gomemcache doesn't have a Close method, we just do nothing.
func (c *MemcachedCache[T]) Close() error {
	return nil
}

// compiler check to ensure MemcachedCache implements gocachex.Cache
var _ gocachex.Cache[string] = (*MemcachedCache[string])(nil)
