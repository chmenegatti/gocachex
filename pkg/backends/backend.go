// Package backends provides the interface and implementations for different cache backends.
package backends

import (
	"context"
	"fmt"
	"time"

	"github.com/chmenegatti/gocachex/pkg/config"
)

// Backend represents a cache backend interface that all implementations must satisfy.
type Backend interface {
	// Basic operations
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)
	SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error
	DeleteMulti(ctx context.Context, keys []string) error

	// Atomic operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Advanced operations
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Management operations
	Clear(ctx context.Context) error
	Stats(ctx context.Context) (*Stats, error)
	Health(ctx context.Context) error
	Close() error
}

// Stats represents backend statistics.
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

// Serializer represents a data serializer interface.
type Serializer interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
	ContentType() string
}

// Compressor represents a data compressor interface.
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	Algorithm() string
}

// New creates a new backend instance based on the configuration.
func New(backendType string, cfg config.Config) (Backend, error) {
	switch backendType {
	case "memory":
		return NewMemoryBackend(cfg.Memory)
	case "redis":
		return NewRedisBackend(cfg.Redis)
	case "memcached":
		return NewMemcachedBackend(cfg.Memcached)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", backendType)
	}
}

// NewSerializer creates a new serializer based on the type.
func NewSerializer(serializerType string) (Serializer, error) {
	switch serializerType {
	case "json":
		return &JSONSerializer{}, nil
	case "gob":
		return &GobSerializer{}, nil
	case "msgpack":
		return &MsgPackSerializer{}, nil
	default:
		return nil, fmt.Errorf("unsupported serializer type: %s", serializerType)
	}
}

// NewCompressor creates a new compressor based on the algorithm.
func NewCompressor(algorithm string) (Compressor, error) {
	switch algorithm {
	case "gzip":
		return &GzipCompressor{}, nil
	case "lz4":
		return &LZ4Compressor{}, nil
	case "snappy":
		return &SnappyCompressor{}, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}
