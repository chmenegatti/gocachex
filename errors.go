package gocachex

import "errors"

var (
	// ErrCacheMiss is returned when a key is not found in the cache.
	ErrCacheMiss = errors.New("cache miss")

	// ErrInvalidConfig is returned when the given cache configuration is invalid.
	ErrInvalidConfig = errors.New("invalid cache configuration")

	// ErrBackendUnavailable is returned when the underlying caching backend is unreachable.
	ErrBackendUnavailable = errors.New("cache backend unavailable")
)
