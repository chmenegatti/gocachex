package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	ctx := context.Background()
	cache := memory.New[string]()

	// Test Set and Get
	err := cache.Set(ctx, "key1", "value1", time.Minute)
	require.NoError(t, err)

	val, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test Exists
	exists, err := cache.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Delete
	err = cache.Delete(ctx, "key1")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, gocachex.ErrCacheMiss)

	exists, err = cache.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_TTL(t *testing.T) {
	ctx := context.Background()
	// Create cache with active background cleanup
	cache := memory.New[string](memory.WithCleanupInterval[string](50 * time.Millisecond))
	defer cache.Close()

	err := cache.Set(ctx, "key_ttl", "val", 10*time.Millisecond)
	require.NoError(t, err)

	// immediately should exist
	val, err := cache.Get(ctx, "key_ttl")
	require.NoError(t, err)
	assert.Equal(t, "val", val)

	// wait for expiration
	time.Sleep(20 * time.Millisecond)

	_, err = cache.Get(ctx, "key_ttl")
	assert.ErrorIs(t, err, gocachex.ErrCacheMiss)

	// manually trigger another set to ensure cleanup routine doesn't panic and works over time
	err = cache.Set(ctx, "key_long", "longval", time.Hour)
	require.NoError(t, err)

	time.Sleep(60 * time.Millisecond) // wait for at least one cleanup tick
	exists, _ := cache.Exists(ctx, "key_long")
	assert.True(t, exists)
}

func TestMemoryCache_Clear(t *testing.T) {
	ctx := context.Background()
	cache := memory.New[int]()

	cache.Set(ctx, "a", 1, 0)
	cache.Set(ctx, "b", 2, 0)

	err := cache.Clear(ctx)
	require.NoError(t, err)

	existsA, _ := cache.Exists(ctx, "a")
	existsB, _ := cache.Exists(ctx, "b")
	assert.False(t, existsA)
	assert.False(t, existsB)
}
