package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/chmenegatti/gocachex"
	goredis "github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-backed, generic cache implementation.
type RedisCache[T any] struct {
	client goredis.UniversalClient
}

// New creates a new Redis cache backend.
func New[T any](client goredis.UniversalClient) *RedisCache[T] {
	return &RedisCache[T]{
		client: client,
	}
}

// Get retrieves a value from the cache.
func (c *RedisCache[T]) Get(ctx context.Context, key string) (T, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		var zero T
		return zero, gocachex.ErrCacheMiss
	} else if err != nil {
		var zero T
		return zero, err
	}

	var parsed T
	if err := json.Unmarshal(val, &parsed); err != nil {
		var zero T
		return zero, err
	}

	return parsed, nil
}

// Set stores a value in the cache with the given TTL.
func (c *RedisCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, bytes, ttl).Err()
}

// Delete removes a value from the cache.
func (c *RedisCache[T]) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in the cache.
func (c *RedisCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Clear purges all keys from the cache.
// Note: FLUSHDB is a dangerous operation. It is exposed to satisfy the Cache interface.
func (c *RedisCache[T]) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Close closes the Redis client.
func (c *RedisCache[T]) Close() error {
	return c.client.Close()
}

// compiler check to ensure RedisCache implements gocachex.Cache
var _ gocachex.Cache[string] = (*RedisCache[string])(nil)
