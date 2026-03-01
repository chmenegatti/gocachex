// Package config provides configuration structures and validation for GoCacheX.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config represents the main configuration for GoCacheX.
type Config struct {
	// Backend specifies the cache backend to use: "memory", "redis", "memcached"
	Backend string `json:"backend"`

	// Compression enables data compression
	Compression bool `json:"compression"`

	// CompressionAlgorithm specifies the compression algorithm: "gzip", "lz4", "snappy"
	CompressionAlgorithm string `json:"compression_algorithm"`

	// Serializer specifies the serialization format: "json", "gob", "msgpack"
	Serializer string `json:"serializer"`

	// Distributed enables distributed cache mode
	Distributed bool `json:"distributed"`

	// Hierarchical enables hierarchical caching (L1/L2)
	Hierarchical bool `json:"hierarchical"`

	// Memory configuration
	Memory MemoryConfig `json:"memory,omitempty"`

	// Redis configuration
	Redis RedisConfig `json:"redis,omitempty"`

	// Memcached configuration
	Memcached MemcachedConfig `json:"memcached,omitempty"`

	// gRPC configuration for distributed mode
	GRPC GRPCConfig `json:"grpc,omitempty"`

	// L1 cache configuration (hierarchical mode)
	L1 CacheConfig `json:"l1,omitempty"`

	// L2 cache configuration (hierarchical mode)
	L2 CacheConfig `json:"l2,omitempty"`

	// Prometheus metrics configuration
	Prometheus PrometheusConfig `json:"prometheus,omitempty"`

	// Tracing configuration
	Tracing TracingConfig `json:"tracing,omitempty"`

	// Sharding configuration
	Sharding ShardingConfig `json:"sharding,omitempty"`
}

// MemoryConfig represents configuration for in-memory cache backend.
type MemoryConfig struct {
	// MaxSize is the maximum memory size (e.g., "100MB", "1GB")
	MaxSize string `json:"max_size"`

	// MaxKeys is the maximum number of keys
	MaxKeys int64 `json:"max_keys"`

	// EvictionPolicy specifies the eviction policy: "lru", "lfu", "random"
	EvictionPolicy string `json:"eviction_policy"`

	// DefaultTTL is the default TTL for keys
	DefaultTTL time.Duration `json:"default_ttl"`

	// CleanupInterval is the interval for cleanup operations
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// RedisConfig represents configuration for Redis backend.
type RedisConfig struct {
	// Addresses is a list of Redis server addresses
	Addresses []string `json:"addresses"`

	// Password for Redis authentication
	Password string `json:"password"`

	// DB is the Redis database number
	DB int `json:"db"`

	// PoolSize is the connection pool size
	PoolSize int `json:"pool_size"`

	// DialTimeout is the connection timeout
	DialTimeout time.Duration `json:"dial_timeout"`

	// ReadTimeout is the read timeout
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout is the write timeout
	WriteTimeout time.Duration `json:"write_timeout"`

	// PoolTimeout is the pool timeout
	PoolTimeout time.Duration `json:"pool_timeout"`

	// IdleTimeout is the idle connection timeout
	IdleTimeout time.Duration `json:"idle_timeout"`

	// IdleCheckFrequency is the frequency of idle connection checks
	IdleCheckFrequency time.Duration `json:"idle_check_frequency"`

	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries"`

	// MinRetryBackoff is the minimum retry backoff
	MinRetryBackoff time.Duration `json:"min_retry_backoff"`

	// MaxRetryBackoff is the maximum retry backoff
	MaxRetryBackoff time.Duration `json:"max_retry_backoff"`

	// TLSConfig enables TLS
	TLS bool `json:"tls"`

	// TLSSkipVerify skips TLS certificate verification
	TLSSkipVerify bool `json:"tls_skip_verify"`

	// Cluster mode configuration
	Cluster RedisClusterConfig `json:"cluster,omitempty"`
}

// RedisClusterConfig represents Redis cluster configuration.
type RedisClusterConfig struct {
	// Enabled indicates if cluster mode is enabled
	Enabled bool `json:"enabled"`

	// MaxRedirects is the maximum number of redirects
	MaxRedirects int `json:"max_redirects"`

	// ReadOnly enables read-only mode
	ReadOnly bool `json:"read_only"`

	// RouteByLatency enables routing by latency
	RouteByLatency bool `json:"route_by_latency"`

	// RouteRandomly enables random routing
	RouteRandomly bool `json:"route_randomly"`
}

// MemcachedConfig represents configuration for Memcached backend.
type MemcachedConfig struct {
	// Servers is a list of Memcached server addresses
	Servers []string `json:"servers"`

	// Timeout is the connection timeout
	Timeout time.Duration `json:"timeout"`

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int `json:"max_idle_conns"`
}

// GRPCConfig represents gRPC configuration for distributed mode.
type GRPCConfig struct {
	// Port is the gRPC server port
	Port int `json:"port"`

	// Peers is a list of peer addresses
	Peers []string `json:"peers"`

	// TLS configuration
	TLS bool `json:"tls"`

	// CertFile is the path to the TLS certificate file
	CertFile string `json:"cert_file"`

	// KeyFile is the path to the TLS key file
	KeyFile string `json:"key_file"`

	// CAFile is the path to the CA certificate file
	CAFile string `json:"ca_file"`

	// ServerName for TLS verification
	ServerName string `json:"server_name"`
}

// CacheConfig represents cache configuration for hierarchical mode.
type CacheConfig struct {
	// Backend specifies the cache backend
	Backend string `json:"backend"`

	// Size is the cache size
	Size string `json:"size"`

	// TTL is the default TTL
	TTL time.Duration `json:"ttl"`

	// Backend-specific configurations
	Memory    MemoryConfig    `json:"memory,omitempty"`
	Redis     RedisConfig     `json:"redis,omitempty"`
	Memcached MemcachedConfig `json:"memcached,omitempty"`
}

// PrometheusConfig represents Prometheus metrics configuration.
type PrometheusConfig struct {
	// Enabled indicates if Prometheus metrics are enabled
	Enabled bool `json:"enabled"`

	// Namespace is the metrics namespace
	Namespace string `json:"namespace"`

	// Subsystem is the metrics subsystem
	Subsystem string `json:"subsystem"`

	// Port is the metrics server port
	Port int `json:"port"`

	// Path is the metrics endpoint path
	Path string `json:"path"`

	// Labels are additional labels for metrics
	Labels map[string]string `json:"labels"`
}

// TracingConfig represents tracing configuration.
type TracingConfig struct {
	// Enabled indicates if tracing is enabled
	Enabled bool `json:"enabled"`

	// Provider specifies the tracing provider: "jaeger", "zipkin", "otlp"
	Provider string `json:"provider"`

	// Endpoint is the tracing endpoint
	Endpoint string `json:"endpoint"`

	// ServiceName is the service name for tracing
	ServiceName string `json:"service_name"`

	// SampleRate is the sampling rate (0.0 to 1.0)
	SampleRate float64 `json:"sample_rate"`

	// Headers are additional headers for tracing
	Headers map[string]string `json:"headers"`
}

// ShardingConfig represents sharding configuration.
type ShardingConfig struct {
	// Enabled indicates if sharding is enabled
	Enabled bool `json:"enabled"`

	// Algorithm specifies the sharding algorithm: "consistent", "hash", "range"
	Algorithm string `json:"algorithm"`

	// Replicas is the number of replicas for consistent hashing
	Replicas int `json:"replicas"`

	// Shards is the number of shards
	Shards int `json:"shards"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate backend
	validBackends := []string{"memory", "redis", "memcached"}
	if !contains(validBackends, c.Backend) {
		return fmt.Errorf("invalid backend: %s, must be one of %v", c.Backend, validBackends)
	}

	// Validate serializer
	if c.Serializer == "" {
		c.Serializer = "json"
	}
	validSerializers := []string{"json", "gob", "msgpack"}
	if !contains(validSerializers, c.Serializer) {
		return fmt.Errorf("invalid serializer: %s, must be one of %v", c.Serializer, validSerializers)
	}

	// Validate compression algorithm
	if c.Compression && c.CompressionAlgorithm == "" {
		c.CompressionAlgorithm = "gzip"
	}
	if c.Compression {
		validAlgorithms := []string{"gzip", "lz4", "snappy"}
		if !contains(validAlgorithms, c.CompressionAlgorithm) {
			return fmt.Errorf("invalid compression algorithm: %s, must be one of %v", c.CompressionAlgorithm, validAlgorithms)
		}
	}

	// Validate hierarchical configuration
	if c.Hierarchical {
		if c.L1.Backend == "" || c.L2.Backend == "" {
			return fmt.Errorf("hierarchical mode requires both L1 and L2 backend configurations")
		}
	}

	// Validate distributed configuration
	if c.Distributed {
		if c.GRPC.Port == 0 {
			c.GRPC.Port = 50051
		}
		if len(c.GRPC.Peers) == 0 {
			return fmt.Errorf("distributed mode requires at least one peer")
		}
	}

	// Validate backend-specific configurations
	switch c.Backend {
	case "redis":
		if err := c.validateRedisConfig(); err != nil {
			return err
		}
	case "memcached":
		if err := c.validateMemcachedConfig(); err != nil {
			return err
		}
	case "memory":
		if err := c.validateMemoryConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateRedisConfig() error {
	if len(c.Redis.Addresses) == 0 {
		c.Redis.Addresses = []string{"localhost:6379"}
	}

	// Set defaults
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
	if c.Redis.DialTimeout == 0 {
		c.Redis.DialTimeout = 5 * time.Second
	}
	if c.Redis.ReadTimeout == 0 {
		c.Redis.ReadTimeout = 3 * time.Second
	}
	if c.Redis.WriteTimeout == 0 {
		c.Redis.WriteTimeout = 3 * time.Second
	}

	return nil
}

func (c *Config) validateMemcachedConfig() error {
	if len(c.Memcached.Servers) == 0 {
		c.Memcached.Servers = []string{"localhost:11211"}
	}

	// Set defaults
	if c.Memcached.Timeout == 0 {
		c.Memcached.Timeout = 100 * time.Millisecond
	}
	if c.Memcached.MaxIdleConns == 0 {
		c.Memcached.MaxIdleConns = 2
	}

	return nil
}

func (c *Config) validateMemoryConfig() error {
	// Set defaults
	if c.Memory.MaxSize == "" {
		c.Memory.MaxSize = "100MB"
	}
	if c.Memory.EvictionPolicy == "" {
		c.Memory.EvictionPolicy = "lru"
	}
	if c.Memory.CleanupInterval == 0 {
		c.Memory.CleanupInterval = 10 * time.Minute
	}

	validPolicies := []string{"lru", "lfu", "random"}
	if !contains(validPolicies, c.Memory.EvictionPolicy) {
		return fmt.Errorf("invalid eviction policy: %s, must be one of %v", c.Memory.EvictionPolicy, validPolicies)
	}

	return nil
}

// LoadFromFile loads configuration from a JSON file.
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() (*Config, error) {
	config := &Config{
		Backend:              getEnv("GOCACHEX_BACKEND", "memory"),
		Compression:          getEnvBool("GOCACHEX_COMPRESSION", false),
		CompressionAlgorithm: getEnv("GOCACHEX_COMPRESSION_ALGORITHM", "gzip"),
		Serializer:           getEnv("GOCACHEX_SERIALIZER", "json"),
		Distributed:          getEnvBool("GOCACHEX_DISTRIBUTED", false),
		Hierarchical:         getEnvBool("GOCACHEX_HIERARCHICAL", false),
	}

	// Redis configuration
	redisAddresses := getEnv("GOCACHEX_REDIS_ADDRESSES", "localhost:6379")
	config.Redis.Addresses = strings.Split(redisAddresses, ",")
	config.Redis.Password = getEnv("GOCACHEX_REDIS_PASSWORD", "")
	config.Redis.DB = getEnvInt("GOCACHEX_REDIS_DB", 0)

	// Memcached configuration
	memcachedServers := getEnv("GOCACHEX_MEMCACHED_SERVERS", "localhost:11211")
	config.Memcached.Servers = strings.Split(memcachedServers, ",")

	// gRPC configuration
	config.GRPC.Port = getEnvInt("GOCACHEX_GRPC_PORT", 50051)
	grpcPeers := getEnv("GOCACHEX_GRPC_PEERS", "")
	if grpcPeers != "" {
		config.GRPC.Peers = strings.Split(grpcPeers, ",")
	}

	// Prometheus configuration
	config.Prometheus.Enabled = getEnvBool("GOCACHEX_PROMETHEUS_ENABLED", false)
	config.Prometheus.Port = getEnvInt("GOCACHEX_PROMETHEUS_PORT", 8080)
	config.Prometheus.Path = getEnv("GOCACHEX_PROMETHEUS_PATH", "/metrics")

	// Tracing configuration
	config.Tracing.Enabled = getEnvBool("GOCACHEX_TRACING_ENABLED", false)
	config.Tracing.Provider = getEnv("GOCACHEX_TRACING_PROVIDER", "jaeger")
	config.Tracing.Endpoint = getEnv("GOCACHEX_TRACING_ENDPOINT", "")
	config.Tracing.ServiceName = getEnv("GOCACHEX_TRACING_SERVICE_NAME", "gocachex")

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveToFile saves configuration to a JSON file.
func (c *Config) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue := parseInt(value); intValue != 0 {
			return intValue
		}
	}
	return defaultValue
}

func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}
