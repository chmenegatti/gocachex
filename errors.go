package gocachex

import "errors"

var (
	// ErrCacheMiss indicates that the requested key does not exist in the cache
	// or has already expired.
	ErrCacheMiss = errors.New("cache miss")

	// ErrInvalidConfig indicates that the configuration provided is invalid.
	ErrInvalidConfig = errors.New("invalid cache configuration")
)
