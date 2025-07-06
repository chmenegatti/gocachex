package gocachex

import (
	"context"
	"fmt"
	"time"

	"github.com/chmenegatti/gocachex/pkg/backends"
	"github.com/chmenegatti/gocachex/pkg/config"
	"github.com/chmenegatti/gocachex/pkg/sharding"
)

// NoOpSpan is a no-operation span for when tracing is disabled.
type NoOpSpan struct{}

func (n NoOpSpan) End() {}

// Helper methods for CacheClient

// initHierarchicalCache initializes hierarchical (L1/L2) cache.
func (c *CacheClient) initHierarchicalCache() error {
	// Initialize L1 cache
	l1Cache, err := New(config.Config{
		Backend:     c.config.L1.Backend,
		Memory:      c.config.L1.Memory,
		Redis:       c.config.L1.Redis,
		Memcached:   c.config.L1.Memcached,
		Serializer:  c.config.Serializer,
		Compression: c.config.Compression,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize L1 cache: %w", err)
	}
	c.l1Cache = l1Cache

	// Initialize L2 cache
	l2Cache, err := New(config.Config{
		Backend:     c.config.L2.Backend,
		Memory:      c.config.L2.Memory,
		Redis:       c.config.L2.Redis,
		Memcached:   c.config.L2.Memcached,
		Serializer:  c.config.Serializer,
		Compression: c.config.Compression,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize L2 cache: %w", err)
	}
	c.l2Cache = l2Cache

	return nil
}

// initSharding initializes distributed cache sharding.
func (c *CacheClient) initSharding() error {
	sharder := sharding.NewSharder(c.config.Sharding)

	// Create shards based on configuration
	shardCount := c.config.Sharding.Shards
	if shardCount <= 0 {
		shardCount = 3 // Default shard count
	}

	for i := 0; i < shardCount; i++ {
		backend, err := backends.New(c.config.Backend, *c.config)
		if err != nil {
			return fmt.Errorf("failed to create shard %d: %w", i, err)
		}
		c.shards = append(c.shards, backend)
		if err := sharder.AddShard(backend); err != nil {
			return fmt.Errorf("failed to add shard %d: %w", i, err)
		}
	}

	return nil
}

// startSpan starts a tracing span if tracing is enabled.
func (c *CacheClient) startSpan(ctx context.Context, operationName string) (context.Context, interface{ End() }) {
	return ctx, NoOpSpan{}
}

// getSingle gets a value from a single backend.
func (c *CacheClient) getSingle(ctx context.Context, key string) (interface{}, error) {
	// Get raw data from backend
	data, err := c.backend.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Decompress if needed
	if c.compressor != nil {
		data, err = c.compressor.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress data: %w", err)
		}
	}

	// Deserialize
	var result interface{}
	if err := c.serializer.Deserialize(data, &result); err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return result, nil
}

// setSingle sets a value in a single backend.
func (c *CacheClient) setSingle(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serialize
	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	// Compress if needed
	if c.compressor != nil {
		data, err = c.compressor.Compress(data)
		if err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
	}

	// Store in backend
	return c.backend.Set(ctx, key, data, ttl)
}

// deleteSingle deletes a value from a single backend.
func (c *CacheClient) deleteSingle(ctx context.Context, key string) error {
	return c.backend.Delete(ctx, key)
}

// getHierarchical gets a value from hierarchical cache (L1/L2).
func (c *CacheClient) getHierarchical(ctx context.Context, key string) (interface{}, error) {
	// Try L1 cache first
	value, err := c.l1Cache.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// Try L2 cache
	value, err = c.l2Cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Found in L2, promote to L1
	// Store in L1 with appropriate TTL
	l1TTL := c.config.L1.TTL
	if l1TTL == 0 {
		l1TTL = 5 * time.Minute // Default L1 TTL
	}
	if err := c.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
		// Log error but don't fail the operation since we have the value
		// In a real implementation, you might want to log this
		_ = err // Acknowledge the error to satisfy linter
	}

	return value, nil
}

// setHierarchical sets a value in hierarchical cache (L1/L2).
func (c *CacheClient) setHierarchical(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Set in both L1 and L2
	if err := c.l1Cache.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("failed to set in L1 cache: %w", err)
	}

	if err := c.l2Cache.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("failed to set in L2 cache: %w", err)
	}

	return nil
}

// deleteHierarchical deletes a value from hierarchical cache (L1/L2).
func (c *CacheClient) deleteHierarchical(ctx context.Context, key string) error {
	// Delete from both L1 and L2
	err1 := c.l1Cache.Delete(ctx, key)
	err2 := c.l2Cache.Delete(ctx, key)

	// Return error if both failed
	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to delete from both caches: L1=%v, L2=%v", err1, err2)
	}

	return nil
}

// existsHierarchical checks if a key exists in hierarchical cache.
func (c *CacheClient) existsHierarchical(ctx context.Context, key string) (bool, error) {
	// Check L1 first
	if exists, err := c.l1Cache.Exists(ctx, key); err == nil && exists {
		return true, nil
	}

	// Check L2
	return c.l2Cache.Exists(ctx, key)
}

// getDistributed gets a value from distributed cache.
func (c *CacheClient) getDistributed(ctx context.Context, key string) (interface{}, error) {
	shard := c.getShard(key)
	if shard == nil {
		return nil, fmt.Errorf("no shard available for key: %s", key)
	}

	// Get raw data from shard
	data, err := shard.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Decompress if needed
	if c.compressor != nil {
		data, err = c.compressor.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress data: %w", err)
		}
	}

	// Deserialize
	var result interface{}
	if err := c.serializer.Deserialize(data, &result); err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return result, nil
}

// setDistributed sets a value in distributed cache.
func (c *CacheClient) setDistributed(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(key)
	if shard == nil {
		return fmt.Errorf("no shard available for key: %s", key)
	}

	// Serialize
	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	// Compress if needed
	if c.compressor != nil {
		data, err = c.compressor.Compress(data)
		if err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
	}

	// Store in shard
	return shard.Set(ctx, key, data, ttl)
}

// deleteDistributed deletes a value from distributed cache.
func (c *CacheClient) deleteDistributed(ctx context.Context, key string) error {
	shard := c.getShard(key)
	if shard == nil {
		return fmt.Errorf("no shard available for key: %s", key)
	}

	return shard.Delete(ctx, key)
}

// existsDistributed checks if a key exists in distributed cache.
func (c *CacheClient) existsDistributed(ctx context.Context, key string) (bool, error) {
	shard := c.getShard(key)
	if shard == nil {
		return false, fmt.Errorf("no shard available for key: %s", key)
	}

	return shard.Exists(ctx, key)
}

// getShard returns the appropriate shard for a given key.
func (c *CacheClient) getShard(key string) backends.Backend {
	if len(c.shards) == 0 {
		return nil
	}

	if len(c.shards) == 1 {
		return c.shards[0]
	}

	// Simple hash-based sharding
	index := sharding.ShardKey(key, len(c.shards))
	return c.shards[index]
}

// statsHierarchical returns stats for hierarchical cache.
func (c *CacheClient) statsHierarchical(ctx context.Context) (*Stats, error) {
	l1Stats, err := c.l1Cache.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get L1 stats: %w", err)
	}

	l2Stats, err := c.l2Cache.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get L2 stats: %w", err)
	}

	// Combine stats
	return &Stats{
		Hits:        l1Stats.Hits + l2Stats.Hits,
		Misses:      l1Stats.Misses + l2Stats.Misses,
		Sets:        l1Stats.Sets + l2Stats.Sets,
		Deletes:     l1Stats.Deletes + l2Stats.Deletes,
		Evictions:   l1Stats.Evictions + l2Stats.Evictions,
		KeyCount:    l1Stats.KeyCount + l2Stats.KeyCount,
		MemoryUsage: l1Stats.MemoryUsage + l2Stats.MemoryUsage,
		Uptime:      max(l1Stats.Uptime, l2Stats.Uptime),
	}, nil
}

// statsDistributed returns stats for distributed cache.
func (c *CacheClient) statsDistributed(ctx context.Context) (*Stats, error) {
	combinedStats := &Stats{}

	for _, shard := range c.shards {
		shardStats, err := shard.Stats(ctx)
		if err != nil {
			continue // Skip failed shards
		}

		combinedStats.Hits += shardStats.Hits
		combinedStats.Misses += shardStats.Misses
		combinedStats.Sets += shardStats.Sets
		combinedStats.Deletes += shardStats.Deletes
		combinedStats.Evictions += shardStats.Evictions
		combinedStats.KeyCount += shardStats.KeyCount
		combinedStats.MemoryUsage += shardStats.MemoryUsage
		if shardStats.Uptime > combinedStats.Uptime {
			combinedStats.Uptime = shardStats.Uptime
		}
	}

	return combinedStats, nil
}

// max returns the maximum of two int64 values.
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// TracingInterface defines the interface for tracing operations.
type TracingInterface interface {
	StartSpan(ctx context.Context, operationName string) (context.Context, interface{ End() })
	IsEnabled() bool
}
