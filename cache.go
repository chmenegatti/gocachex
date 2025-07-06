// Package gocachex provides a plug-and-play distributed cache library for Go.
//
// GoCacheX supports multiple backends (Redis, Memcached, In-Memory), advanced features
// like hierarchical caching, sharding, compression, and comprehensive monitoring.
//
// Basic usage:
//
//	cache := gocachex.New(gocachex.Config{
//		Backend: "memory",
//	})
//
//	ctx := context.Background()
//	cache.Set(ctx, "key", "value", time.Minute)
//	value, err := cache.Get(ctx, "key")
//
// Advanced usage with Redis:
//
//	cache := gocachex.New(gocachex.Config{
//		Backend: "redis",
//		Redis: gocachex.RedisConfig{
//			Addresses: []string{"localhost:6379"},
//		},
//		Compression: true,
//		Distributed: true,
//	})
package gocachex

import (
	"context"
	"fmt"
	"time"

	"github.com/chmenegatti/gocachex/pkg/backends"
	"github.com/chmenegatti/gocachex/pkg/config"
)

// Cache represents the main cache interface that all backends must implement.
// It provides a unified API for cache operations across different storage backends.
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	DeleteMulti(ctx context.Context, keys []string) error

	// Atomic operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Advanced operations
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}) (interface{}, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Management operations
	Clear(ctx context.Context) error
	Stats(ctx context.Context) (*Stats, error)
	Health(ctx context.Context) error
	Close() error
}

// Stats represents cache statistics and metrics.
type Stats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Sets        int64 `json:"sets"`
	Deletes     int64 `json:"deletes"`
	Evictions   int64 `json:"evictions"`
	KeyCount    int64 `json:"key_count"`
	MemoryUsage int64 `json:"memory_usage"`
	Uptime      int64 `json:"uptime"`
}

// HitRatio calculates the cache hit ratio.
func (s *Stats) HitRatio() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// CacheClient is the main implementation of the Cache interface.
// It provides a unified interface to different cache backends with additional features
// like compression, serialization, monitoring, and distributed operations.
type CacheClient struct {
	backend backends.Backend
	config  *config.Config
	// metrics    *metrics.Collector
	// tracer     *tracing.Tracer
	shards     []backends.Backend
	l1Cache    Cache
	l2Cache    Cache
	serializer backends.Serializer
	compressor backends.Compressor
}

// New creates a new cache client with the given configuration.
// It initializes the appropriate backend, sets up monitoring, and configures
// additional features like compression and hierarchical caching.
func New(cfg config.Config) (Cache, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client := &CacheClient{
		config: &cfg,
	}

	// Initialize metrics collector
	// if cfg.Prometheus.Enabled {
	// 	client.metrics = metrics.New(cfg.Prometheus)
	// }

	// Initialize tracer
	// if cfg.Tracing.Enabled {
	// 	tracer, err := tracing.New(cfg.Tracing)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	// 	}
	// 	client.tracer = tracer
	// }

	// Initialize serializer
	serializer, err := backends.NewSerializer(cfg.Serializer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize serializer: %w", err)
	}
	client.serializer = serializer

	// Initialize compressor if enabled
	if cfg.Compression {
		compressor, err := backends.NewCompressor(cfg.CompressionAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize compressor: %w", err)
		}
		client.compressor = compressor
	}

	// Initialize hierarchical cache if enabled
	if cfg.Hierarchical {
		if err := client.initHierarchicalCache(); err != nil {
			return nil, fmt.Errorf("failed to initialize hierarchical cache: %w", err)
		}
		return client, nil
	}

	// Initialize main backend
	backend, err := backends.New(cfg.Backend, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize backend: %w", err)
	}
	client.backend = backend

	// Initialize sharding if distributed
	if cfg.Distributed {
		if err := client.initSharding(); err != nil {
			return nil, fmt.Errorf("failed to initialize sharding: %w", err)
		}
	}

	return client, nil
}

// Get retrieves a value from the cache.
func (c *CacheClient) Get(ctx context.Context, key string) (interface{}, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.get")
	defer span.End()

	// Hierarchical cache check
	if c.config.Hierarchical {
		return c.getHierarchical(ctx, key)
	}

	// Distributed cache check
	if c.config.Distributed {
		return c.getDistributed(ctx, key)
	}

	// Single backend get
	return c.getSingle(ctx, key)
}

// Set stores a value in the cache with the specified TTL.
func (c *CacheClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.set")
	defer span.End()

	// Hierarchical cache set
	if c.config.Hierarchical {
		return c.setHierarchical(ctx, key, value, ttl)
	}

	// Distributed cache set
	if c.config.Distributed {
		return c.setDistributed(ctx, key, value, ttl)
	}

	// Single backend set
	return c.setSingle(ctx, key, value, ttl)
}

// Delete removes a value from the cache.
func (c *CacheClient) Delete(ctx context.Context, key string) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.delete")
	defer span.End()

	// Hierarchical cache delete
	if c.config.Hierarchical {
		return c.deleteHierarchical(ctx, key)
	}

	// Distributed cache delete
	if c.config.Distributed {
		return c.deleteDistributed(ctx, key)
	}

	// Single backend delete
	return c.deleteSingle(ctx, key)
}

// Exists checks if a key exists in the cache.
func (c *CacheClient) Exists(ctx context.Context, key string) (bool, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.exists")
	defer span.End()

	// Hierarchical cache check
	if c.config.Hierarchical {
		return c.existsHierarchical(ctx, key)
	}

	// Distributed cache check
	if c.config.Distributed {
		return c.existsDistributed(ctx, key)
	}

	// Single backend check
	return c.backend.Exists(ctx, key)
}

// GetMulti retrieves multiple values from the cache.
func (c *CacheClient) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.get_multi")
	defer span.End()

	result := make(map[string]interface{})

	// For hierarchical or distributed cache, we need to handle each key individually
	if c.config.Hierarchical || c.config.Distributed {
		for _, key := range keys {
			value, err := c.Get(ctx, key)
			if err == nil {
				result[key] = value
			}
		}
		return result, nil
	}

	// Single backend get multi
	rawResult, err := c.backend.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]interface{})
	for key, value := range rawResult {
		// Deserialize each value
		var deserializedValue interface{}
		if err := c.serializer.Deserialize(value, &deserializedValue); err == nil {
			resultMap[key] = deserializedValue
		}
	}

	return resultMap, nil
}

// SetMulti stores multiple values in the cache.
func (c *CacheClient) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.set_multi")
	defer span.End()

	// For hierarchical or distributed cache, we need to handle each item individually
	if c.config.Hierarchical || c.config.Distributed {
		for key, value := range items {
			if err := c.Set(ctx, key, value, ttl); err != nil {
				return err
			}
		}
		return nil
	}

	// Single backend set multi
	serializedItems := make(map[string][]byte)
	for key, value := range items {
		serializedValue, err := c.serializer.Serialize(value)
		if err != nil {
			return err
		}
		serializedItems[key] = serializedValue
	}
	return c.backend.SetMulti(ctx, serializedItems, ttl)
}

// DeleteMulti removes multiple values from the cache.
func (c *CacheClient) DeleteMulti(ctx context.Context, keys []string) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.delete_multi")
	defer span.End()

	// For hierarchical or distributed cache, we need to handle each key individually
	if c.config.Hierarchical || c.config.Distributed {
		for _, key := range keys {
			if err := c.Delete(ctx, key); err != nil {
				return err
			}
		}
		return nil
	}

	// Single backend delete multi
	return c.backend.DeleteMulti(ctx, keys)
}

// Increment atomically increments a numeric value.
func (c *CacheClient) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.increment")
	defer span.End()

	// For hierarchical cache, use L2 for atomic operations
	if c.config.Hierarchical {
		return c.l2Cache.Increment(ctx, key, delta)
	}

	// For distributed cache, use the appropriate shard
	if c.config.Distributed {
		shard := c.getShard(key)
		return shard.Increment(ctx, key, delta)
	}

	// Single backend increment
	return c.backend.Increment(ctx, key, delta)
}

// Decrement atomically decrements a numeric value.
func (c *CacheClient) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.decrement")
	defer span.End()

	// For hierarchical cache, use L2 for atomic operations
	if c.config.Hierarchical {
		return c.l2Cache.Decrement(ctx, key, delta)
	}

	// For distributed cache, use the appropriate shard
	if c.config.Distributed {
		shard := c.getShard(key)
		return shard.Decrement(ctx, key, delta)
	}

	// Single backend decrement
	return c.backend.Decrement(ctx, key, delta)
}

// SetNX sets a value only if the key doesn't exist.
func (c *CacheClient) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.setnx")
	defer span.End()

	// Check if key exists first
	exists, err := c.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	// Set the value
	err = c.Set(ctx, key, value, ttl)
	return err == nil, err
}

// GetSet atomically sets a value and returns the old value.
func (c *CacheClient) GetSet(ctx context.Context, key string, value interface{}) (interface{}, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.getset")
	defer span.End()

	// Get the old value
	oldValue, _ := c.Get(ctx, key)

	// Set the new value
	err := c.Set(ctx, key, value, 0) // Use default TTL
	if err != nil {
		return nil, err
	}

	return oldValue, nil
}

// Expire sets a timeout on a key.
func (c *CacheClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.expire")
	defer span.End()

	// For hierarchical or distributed cache, this operation might not be supported
	if c.config.Hierarchical || c.config.Distributed {
		return fmt.Errorf("expire operation not supported in hierarchical/distributed mode")
	}

	return c.backend.Expire(ctx, key, ttl)
}

// TTL returns the remaining time to live of a key.
func (c *CacheClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.ttl")
	defer span.End()

	// For hierarchical or distributed cache, this operation might not be supported
	if c.config.Hierarchical || c.config.Distributed {
		return 0, fmt.Errorf("ttl operation not supported in hierarchical/distributed mode")
	}

	return c.backend.TTL(ctx, key)
}

// Clear removes all keys from the cache.
func (c *CacheClient) Clear(ctx context.Context) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.clear")
	defer span.End()

	// Hierarchical cache clear
	if c.config.Hierarchical {
		if err := c.l1Cache.Clear(ctx); err != nil {
			return err
		}
		return c.l2Cache.Clear(ctx)
	}

	// Distributed cache clear
	if c.config.Distributed {
		for _, shard := range c.shards {
			if err := shard.Clear(ctx); err != nil {
				return err
			}
		}
		return nil
	}

	// Single backend clear
	return c.backend.Clear(ctx)
}

// Stats returns cache statistics.
func (c *CacheClient) Stats(ctx context.Context) (*Stats, error) {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.stats")
	defer span.End()

	// Hierarchical cache stats
	if c.config.Hierarchical {
		return c.statsHierarchical(ctx)
	}

	// Distributed cache stats
	if c.config.Distributed {
		return c.statsDistributed(ctx)
	}

	// Single backend stats
	backendStats, err := c.backend.Stats(ctx)
	if err != nil {
		return nil, err
	}

	return &Stats{
		Hits:        backendStats.Hits,
		Misses:      backendStats.Misses,
		Sets:        backendStats.Sets,
		Deletes:     backendStats.Deletes,
		Evictions:   backendStats.Evictions,
		KeyCount:    backendStats.KeyCount,
		MemoryUsage: backendStats.MemoryUsage,
		Uptime:      backendStats.Uptime,
	}, nil
}

// Health checks the health of the cache backend.
func (c *CacheClient) Health(ctx context.Context) error {
	// Start tracing span
	ctx, span := c.startSpan(ctx, "cache.health")
	defer span.End()

	// Hierarchical cache health
	if c.config.Hierarchical {
		if err := c.l1Cache.Health(ctx); err != nil {
			return err
		}
		return c.l2Cache.Health(ctx)
	}

	// Distributed cache health
	if c.config.Distributed {
		for _, shard := range c.shards {
			if err := shard.Health(ctx); err != nil {
				return err
			}
		}
		return nil
	}

	// Single backend health
	return c.backend.Health(ctx)
}

// Close closes the cache client and releases resources.
func (c *CacheClient) Close() error {
	var errors []error

	// Close hierarchical caches
	if c.config.Hierarchical {
		if err := c.l1Cache.Close(); err != nil {
			errors = append(errors, err)
		}
		if err := c.l2Cache.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	// Close distributed shards
	if c.config.Distributed {
		for _, shard := range c.shards {
			if err := shard.Close(); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Close main backend
	if c.backend != nil {
		if err := c.backend.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	// Close tracer
	// if c.tracer != nil {
	// 	if err := c.tracer.Close(); err != nil {
	// 		errors = append(errors, err)
	// 	}
	// }

	if len(errors) > 0 {
		return fmt.Errorf("errors closing cache: %v", errors)
	}

	return nil
}
