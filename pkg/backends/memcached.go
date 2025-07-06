package backends

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/chmenegatti/gocachex/pkg/config"
)

// MemcachedBackend implements a Memcached cache backend.
type MemcachedBackend struct {
	client *memcache.Client
	config config.MemcachedConfig
}

// NewMemcachedBackend creates a new Memcached backend.
func NewMemcachedBackend(cfg config.MemcachedConfig) (*MemcachedBackend, error) {
	client := memcache.New(cfg.Servers...)
	client.Timeout = cfg.Timeout
	client.MaxIdleConns = cfg.MaxIdleConns

	// Test connection
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Memcached: %w", err)
	}

	return &MemcachedBackend{
		client: client,
		config: cfg,
	}, nil
}

// Get retrieves a value from Memcached.
func (m *MemcachedBackend) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, fmt.Errorf("key not found")
		}
		return nil, err
	}
	return item.Value, nil
}

// Set stores a value in Memcached.
func (m *MemcachedBackend) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	item := &memcache.Item{
		Key:   key,
		Value: value,
	}

	if ttl > 0 {
		item.Expiration = int32(ttl.Seconds())
	}

	return m.client.Set(item)
}

// Delete removes a value from Memcached.
func (m *MemcachedBackend) Delete(ctx context.Context, key string) error {
	err := m.client.Delete(key)
	if err == memcache.ErrCacheMiss {
		return nil // Key didn't exist, consider it successful
	}
	return err
}

// Exists checks if a key exists in Memcached.
func (m *MemcachedBackend) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetMulti retrieves multiple values from Memcached.
func (m *MemcachedBackend) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	items, err := m.client.GetMulti(keys)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for key, item := range items {
		result[key] = item.Value
	}

	return result, nil
}

// SetMulti stores multiple values in Memcached.
func (m *MemcachedBackend) SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	for key, value := range items {
		if err := m.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMulti removes multiple values from Memcached.
func (m *MemcachedBackend) DeleteMulti(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Increment atomically increments a numeric value in Memcached.
func (m *MemcachedBackend) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	newValue, err := m.client.Increment(key, uint64(delta))
	if err != nil {
		// If key doesn't exist, create it with delta value
		if err == memcache.ErrCacheMiss {
			if err := m.Set(ctx, key, []byte(strconv.FormatInt(delta, 10)), 0); err != nil {
				return 0, err
			}
			return delta, nil
		}
		return 0, err
	}
	return int64(newValue), nil
}

// Decrement atomically decrements a numeric value in Memcached.
func (m *MemcachedBackend) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	newValue, err := m.client.Decrement(key, uint64(delta))
	if err != nil {
		// If key doesn't exist, create it with negative delta value
		if err == memcache.ErrCacheMiss {
			negDelta := -delta
			if err := m.Set(ctx, key, []byte(strconv.FormatInt(negDelta, 10)), 0); err != nil {
				return 0, err
			}
			return negDelta, nil
		}
		return 0, err
	}
	return int64(newValue), nil
}

// Expire sets a timeout on a key (not directly supported by Memcached).
func (m *MemcachedBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// Memcached doesn't support changing expiration without resetting value
	// We need to get the value and set it again with new TTL
	value, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	return m.Set(ctx, key, value, ttl)
}

// TTL returns the remaining time to live of a key (not supported by Memcached).
func (m *MemcachedBackend) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, fmt.Errorf("TTL operation not supported by Memcached")
}

// Clear removes all keys from Memcached.
func (m *MemcachedBackend) Clear(ctx context.Context) error {
	return m.client.FlushAll()
}

// Stats returns Memcached statistics.
func (m *MemcachedBackend) Stats(ctx context.Context) (*Stats, error) {
	// Note: The gomemcache library doesn't expose a simple Stats() method
	// This is a placeholder implementation
	return &Stats{
		Hits:        0,
		Misses:      0,
		Sets:        0,
		Deletes:     0,
		Evictions:   0,
		KeyCount:    0,
		MemoryUsage: 0,
		Uptime:      0,
	}, nil
}

// Health checks the health of the Memcached connection.
func (m *MemcachedBackend) Health(ctx context.Context) error {
	return m.client.Ping()
}

// Close closes the Memcached connection.
func (m *MemcachedBackend) Close() error {
	// Memcached client doesn't have a close method
	return nil
}
