package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/memory"
	"github.com/chmenegatti/gocachex/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusCache_Coverage(t *testing.T) {
	ctx := context.Background()
	baseCache := memory.New[string]()
	promCache := metrics.NewPrometheusCache[string](baseCache, "test_cache")

	// Set
	err := promCache.Set(ctx, "k1", "v1", time.Minute)
	require.NoError(t, err)

	// Get hit
	val, err := promCache.Get(ctx, "k1")
	require.NoError(t, err)
	assert.Equal(t, "v1", val)

	// Get miss
	_, err = promCache.Get(ctx, "k2")
	assert.ErrorIs(t, err, gocachex.ErrCacheMiss)

	// Exists
	exists, err := promCache.Exists(ctx, "k1")
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete
	err = promCache.Delete(ctx, "k1")
	require.NoError(t, err)

	// Clear
	err = promCache.Clear(ctx)
	require.NoError(t, err)

	// Close
	err = promCache.Close()
	require.NoError(t, err)
}
