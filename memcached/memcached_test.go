package memcached_test

import (
	"context"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/memcached"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// We use a mock or skip tests if local memcached isn't running.
// For now, let's write a standard test that requires a running memcached on localhost:11211
// In CI this can be provided as a service.

func setupMemcachedCache[T any](t *testing.T) *memcached.MemcachedCache[T] {
	client := memcache.New("localhost:11211")
	err := client.Ping()
	if err != nil {
		t.Skipf("Skipping memcached test: %v", err)
	}

	cache := memcached.New[T](client)
	return cache
}

func TestMemcachedCache_BasicOperations(t *testing.T) {
	ctx := context.Background()
	cache := setupMemcachedCache[int](t)
	defer cache.Close()

	// Clear before test
	_ = cache.Clear(ctx)

	// Set
	err := cache.Set(ctx, "mc_key1", 42, time.Minute)
	require.NoError(t, err)

	// Get
	val, err := cache.Get(ctx, "mc_key1")
	require.NoError(t, err)
	assert.Equal(t, 42, val)

	// Exists
	exists, err := cache.Exists(ctx, "mc_key1")
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete
	err = cache.Delete(ctx, "mc_key1")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "mc_key1")
	assert.ErrorIs(t, err, gocachex.ErrCacheMiss)
}

func TestMemcachedCache_Clear(t *testing.T) {
	ctx := context.Background()
	cache := setupMemcachedCache[string](t)
	defer cache.Close()

	cache.Set(ctx, "mc_a", "val1", time.Minute)

	err := cache.Clear(ctx)
	require.NoError(t, err)

	exists, _ := cache.Exists(ctx, "mc_a")
	assert.False(t, exists)
}
