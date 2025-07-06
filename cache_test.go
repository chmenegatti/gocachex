package gocachex

import (
	"context"
	"testing"
	"time"

	"github.com/chmenegatti/gocachex/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryBackend(t *testing.T) {
	cache, err := New(config.Config{
		Backend: "memory",
		Memory: config.MemoryConfig{
			MaxSize:         "1MB",
			EvictionPolicy:  "lru",
			CleanupInterval: time.Second,
		},
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test basic set/get
	err = cache.Set(ctx, "test_key", "test_value", time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)

	// Test exists
	exists, err := cache.Exists(ctx, "test_key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test delete
	err = cache.Delete(ctx, "test_key")
	assert.NoError(t, err)

	exists, err = cache.Exists(ctx, "test_key")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryBackendTTL(t *testing.T) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Set with short TTL
	err = cache.Set(ctx, "ttl_key", "ttl_value", 100*time.Millisecond)
	assert.NoError(t, err)

	// Should exist immediately
	value, err := cache.Get(ctx, "ttl_key")
	assert.NoError(t, err)
	assert.Equal(t, "ttl_value", value)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should not exist anymore
	_, err = cache.Get(ctx, "ttl_key")
	assert.Error(t, err)
}

func TestMemoryBackendMultiOperations(t *testing.T) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test SetMulti
	items := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": map[string]string{"nested": "object"},
	}
	err = cache.SetMulti(ctx, items, time.Minute)
	assert.NoError(t, err)

	// Test GetMulti
	keys := []string{"key1", "key2", "key3", "nonexistent"}
	results, err := cache.GetMulti(ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, results, 3) // Only existing keys
	assert.Equal(t, "value1", results["key1"])
	assert.Equal(t, float64(42), results["key2"]) // JSON unmarshals numbers as float64

	// Test DeleteMulti
	err = cache.DeleteMulti(ctx, []string{"key1", "key2"})
	assert.NoError(t, err)

	// Verify deletion
	exists, _ := cache.Exists(ctx, "key1")
	assert.False(t, exists)
	exists, _ = cache.Exists(ctx, "key2")
	assert.False(t, exists)
	exists, _ = cache.Exists(ctx, "key3")
	assert.True(t, exists)
}

func TestMemoryBackendAtomicOperations(t *testing.T) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Test Increment
	value, err := cache.Increment(ctx, "counter", 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value)

	value, err = cache.Increment(ctx, "counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), value)

	// Test Decrement
	value, err = cache.Decrement(ctx, "counter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(6), value)

	// Test SetNX
	success, err := cache.SetNX(ctx, "counter", 100, time.Minute)
	assert.NoError(t, err)
	assert.False(t, success) // Should fail because key exists

	success, err = cache.SetNX(ctx, "new_key", "new_value", time.Minute)
	assert.NoError(t, err)
	assert.True(t, success) // Should succeed because key doesn't exist
}

func TestMemoryBackendStats(t *testing.T) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Perform some operations
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_, _ = cache.Get(ctx, "key1")    // Hit
	_, _ = cache.Get(ctx, "key3")    // Miss
	_ = cache.Delete(ctx, "key2") // Delete

	// Get stats
	stats, err := cache.Stats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.True(t, stats.Hits > 0)
	assert.True(t, stats.Misses > 0)
	assert.True(t, stats.Sets > 0)
	assert.True(t, stats.Deletes > 0)
}

func TestMemoryBackendHealth(t *testing.T) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	err = cache.Health(ctx)
	assert.NoError(t, err)
}

func TestMemoryBackendClear(t *testing.T) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// Add some data
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)

	// Verify data exists
	exists, _ := cache.Exists(ctx, "key1")
	assert.True(t, exists)

	// Clear cache
	err = cache.Clear(ctx)
	assert.NoError(t, err)

	// Verify data is gone
	exists, _ = cache.Exists(ctx, "key1")
	assert.False(t, exists)
	exists, _ = cache.Exists(ctx, "key2")
	assert.False(t, exists)
}

func TestConfigValidation(t *testing.T) {
	// Test invalid backend
	_, err := New(config.Config{
		Backend: "invalid",
	})
	assert.Error(t, err)

	// Test invalid serializer
	_, err = New(config.Config{
		Backend:    "memory",
		Serializer: "invalid",
	})
	assert.Error(t, err)

	// Test valid config
	_, err = New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	assert.NoError(t, err)
}

func BenchmarkMemorySet(b *testing.B) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(b, err)
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key"
		value := map[string]interface{}{
			"id":   i,
			"name": "Benchmark Item",
			"data": make([]int, 100),
		}
		_ = cache.Set(ctx, key, value, time.Minute)
	}
}

func BenchmarkMemoryGet(b *testing.B) {
	cache, err := New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	require.NoError(b, err)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate cache
	value := map[string]interface{}{
		"id":   1,
		"name": "Benchmark Item",
		"data": make([]int, 100),
	}
	_ = cache.Set(ctx, "benchmark_key", value, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(ctx, "benchmark_key")
	}
}
