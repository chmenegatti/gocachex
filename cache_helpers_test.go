package gocachex_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemember(t *testing.T) {
	ctx := context.Background()
	cache := memory.New[string]()
	var calls int32

	loader := func() (string, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond) // simulate slow db
		return "loaded_value", nil
	}

	// Test 1: Cache Miss - should load
	val, err := gocachex.Remember(ctx, cache, "key1", time.Minute, loader)
	require.NoError(t, err)
	assert.Equal(t, "loaded_value", val)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))

	// Test 2: Cache Hit - should not load
	val, err = gocachex.Remember(ctx, cache, "key1", time.Minute, loader)
	require.NoError(t, err)
	assert.Equal(t, "loaded_value", val)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls)) // still 1

	// Test 3: Stampede Protection
	atomic.StoreInt32(&calls, 0)
	cache.Clear(ctx)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := gocachex.Remember(ctx, cache, "stampede_key", time.Minute, loader)
			assert.NoError(t, e)
			assert.Equal(t, "loaded_value", v)
		}()
	}
	wg.Wait()

	// Even with 10 concurrent requests, the loader should only be called once
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestRemember_LoaderError(t *testing.T) {
	ctx := context.Background()
	cache := memory.New[string]()

	expectedErr := errors.New("db down")
	loader := func() (string, error) {
		return "", expectedErr
	}

	val, err := gocachex.Remember(ctx, cache, "err_key", time.Minute, loader)
	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, val)
}
