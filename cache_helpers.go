package gocachex

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"
)

// group is the singleflight Group used to prevent cache stampedes.
var group singleflight.Group

// Remember is a cache-aside helper that retrieves a value from the cache.
// If the key is not in the cache, it calls the provided loader function,
// stores the result in the cache, and returns it.
// It uses singleflight to ensure that concurrent calls for the same key
// only execute the loader function once (cache stampede protection).
func Remember[T any](
	ctx context.Context,
	cache Cache[T],
	key string,
	ttl time.Duration,
	loader func() (T, error),
) (T, error) {
	// 1. Try to get from cache
	val, err := cache.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// Only proceed if it's a cache miss; otherwise, return the original error.
	if !errors.Is(err, ErrCacheMiss) && err.Error() != ErrCacheMiss.Error() {
		var zero T
		return zero, err
	}

	// 2. Cache miss, use singleflight to load the data exactly once
	result, err, _ := group.Do(key, func() (interface{}, error) {
		// Try to read again in case another goroutine loaded it
		val, err := cache.Get(ctx, key)
		if err == nil {
			return val, nil // another goroutine won the race to load
		}

		// Fallback to calling the loader
		loadedValue, loadErr := loader()
		if loadErr != nil {
			return nil, loadErr
		}

		// Store in cache
		setErr := cache.Set(ctx, key, loadedValue, ttl)
		if setErr != nil {
			// Even if we fail to strictly cache it, we successfully loaded it,
			// so we might want to return the value and error, but generally
			// failing to Set should probably bubble up, or just be logged.
			// Depending on strictness, we return setErr here.
			return loadedValue, setErr
		}

		return loadedValue, nil
	})

	if err != nil {
		var zero T
		return zero, err
	}

	return result.(T), nil
}
