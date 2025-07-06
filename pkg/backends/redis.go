package backends

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/chmenegatti/gocachex/pkg/config"
	"github.com/redis/go-redis/v9"
)

// RedisBackend implements a Redis cache backend.
type RedisBackend struct {
	client redis.UniversalClient
	config config.RedisConfig
}

// NewRedisBackend creates a new Redis backend.
func NewRedisBackend(cfg config.RedisConfig) (*RedisBackend, error) {
	var client redis.UniversalClient

	if cfg.Cluster.Enabled {
		// Cluster mode
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:          cfg.Addresses,
			Password:       cfg.Password,
			MaxRedirects:   cfg.Cluster.MaxRedirects,
			ReadOnly:       cfg.Cluster.ReadOnly,
			RouteByLatency: cfg.Cluster.RouteByLatency,
			RouteRandomly:  cfg.Cluster.RouteRandomly,
			DialTimeout:    cfg.DialTimeout,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			PoolSize:       cfg.PoolSize, PoolTimeout: cfg.PoolTimeout,
			ConnMaxIdleTime: cfg.IdleTimeout,
			MaxRetries:      cfg.MaxRetries,
			MinRetryBackoff: cfg.MinRetryBackoff,
			MaxRetryBackoff: cfg.MaxRetryBackoff,
		})
	} else if len(cfg.Addresses) > 1 {
		// Sentinel mode
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    "master", // This should be configurable
			SentinelAddrs: cfg.Addresses,
			Password:      cfg.Password,
			DB:            cfg.DB,
			DialTimeout:   cfg.DialTimeout,
			ReadTimeout:   cfg.ReadTimeout,
			WriteTimeout:  cfg.WriteTimeout,
			PoolSize:      cfg.PoolSize, PoolTimeout: cfg.PoolTimeout,
			ConnMaxIdleTime: cfg.IdleTimeout,
			MaxRetries:      cfg.MaxRetries,
			MinRetryBackoff: cfg.MinRetryBackoff,
			MaxRetryBackoff: cfg.MaxRetryBackoff,
		})
	} else {
		// Single instance mode
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Addresses[0],
			Password:     cfg.Password,
			DB:           cfg.DB,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolSize:     cfg.PoolSize, PoolTimeout: cfg.PoolTimeout,
			ConnMaxIdleTime: cfg.IdleTimeout,
			MaxRetries:      cfg.MaxRetries,
			MinRetryBackoff: cfg.MinRetryBackoff,
			MaxRetryBackoff: cfg.MaxRetryBackoff,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisBackend{
		client: client,
		config: cfg,
	}, nil
}

// Get retrieves a value from Redis.
func (r *RedisBackend) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found")
		}
		return nil, err
	}
	return []byte(val), nil
}

// Set stores a value in Redis.
func (r *RedisBackend) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a value from Redis.
func (r *RedisBackend) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis.
func (r *RedisBackend) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// GetMulti retrieves multiple values from Redis.
func (r *RedisBackend) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, val := range values {
		if val != nil {
			if str, ok := val.(string); ok {
				result[keys[i]] = []byte(str)
			}
		}
	}

	return result, nil
}

// SetMulti stores multiple values in Redis.
func (r *RedisBackend) SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	pipe := r.client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// DeleteMulti removes multiple values from Redis.
func (r *RedisBackend) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// Increment atomically increments a numeric value in Redis.
func (r *RedisBackend) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}

// Decrement atomically decrements a numeric value in Redis.
func (r *RedisBackend) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.DecrBy(ctx, key, delta).Result()
}

// Expire sets a timeout on a key in Redis.
func (r *RedisBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// TTL returns the remaining time to live of a key in Redis.
func (r *RedisBackend) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Clear removes all keys from the Redis database.
func (r *RedisBackend) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Stats returns Redis statistics.
func (r *RedisBackend) Stats(ctx context.Context) (*Stats, error) {
	info, err := r.client.Info(ctx, "stats", "memory", "keyspace").Result()
	if err != nil {
		return nil, err
	}

	stats := &Stats{}

	// Parse Redis INFO output
	lines := parseRedisInfo(info)

	// Parse stats
	if val, ok := lines["keyspace_hits"]; ok {
		stats.Hits, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := lines["keyspace_misses"]; ok {
		stats.Misses, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := lines["used_memory"]; ok {
		stats.MemoryUsage, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := lines["uptime_in_seconds"]; ok {
		stats.Uptime, _ = strconv.ParseInt(val, 10, 64)
	}

	// Get key count from keyspace info
	dbInfo, err := r.client.Info(ctx, "keyspace").Result()
	if err == nil {
		dbLines := parseRedisInfo(dbInfo)
		if val, ok := dbLines[fmt.Sprintf("db%d", r.config.DB)]; ok {
			// Parse "keys=N,expires=M,avg_ttl=X" format
			if keyCount := parseKeyCount(val); keyCount > 0 {
				stats.KeyCount = keyCount
			}
		}
	}

	return stats, nil
}

// Health checks the health of the Redis connection.
func (r *RedisBackend) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (r *RedisBackend) Close() error {
	return r.client.Close()
}

// parseRedisInfo parses Redis INFO command output into a map.
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := splitLines(info)

	for _, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			if colon := findChar(line, ':'); colon != -1 {
				key := line[:colon]
				value := line[colon+1:]
				result[key] = value
			}
		}
	}

	return result
}

// parseKeyCount parses the key count from db info string.
func parseKeyCount(dbInfo string) int64 {
	// Parse "keys=N,expires=M,avg_ttl=X" format
	parts := splitString(dbInfo, ',')
	for _, part := range parts {
		if len(part) > 5 && part[:5] == "keys=" {
			if count, err := strconv.ParseInt(part[5:], 10, 64); err == nil {
				return count
			}
		}
	}
	return 0
}

// Helper functions for string parsing
func splitLines(s string) []string {
	var result []string
	var current string

	for _, char := range s {
		if char == '\n' || char == '\r' {
			if len(current) > 0 {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}

func splitString(s string, delimiter rune) []string {
	var result []string
	var current string

	for _, char := range s {
		if char == delimiter {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}

func findChar(s string, target rune) int {
	for i, char := range s {
		if char == target {
			return i
		}
	}
	return -1
}
