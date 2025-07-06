package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/pkg/config"
)

const (
	defaultConfigPath = "cache_config.json"
	defaultKey        = "example:key"
	defaultValue      = "example value"
	defaultTTL        = "5m"
)

func main() {
	var (
		configPath = flag.String("config", defaultConfigPath, "Path to cache configuration file")
		backend    = flag.String("backend", "memory", "Cache backend: memory, redis, memcached")
		operation  = flag.String("op", "demo", "Operation: demo, get, set, delete, exists, stats, clear")
		key        = flag.String("key", defaultKey, "Cache key")
		value      = flag.String("value", defaultValue, "Cache value (for set operation)")
		ttlStr     = flag.String("ttl", defaultTTL, "TTL for set operation (e.g., 5m, 1h, 30s)")
		jsonValue  = flag.Bool("json", false, "Treat value as JSON")
		help       = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	fmt.Println("GoCacheX CLI Example")
	fmt.Println("===================")

	// Parse TTL
	ttl, err := time.ParseDuration(*ttlStr)
	if err != nil {
		log.Fatalf("Invalid TTL format: %v", err)
	}

	// Create configuration
	cfg := createConfig(*configPath, *backend)

	// Create cache client
	cache, err := gocachex.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Perform operation
	switch *operation {
	case "demo":
		runDemo(ctx, cache)
	case "get":
		getValue(ctx, cache, *key)
	case "set":
		setValue(ctx, cache, *key, *value, ttl, *jsonValue)
	case "delete":
		deleteValue(ctx, cache, *key)
	case "exists":
		checkExists(ctx, cache, *key)
	case "stats":
		showStats(ctx, cache)
	case "clear":
		clearCache(ctx, cache)
	default:
		fmt.Printf("Unknown operation: %s\n", *operation)
		showHelp()
		os.Exit(1)
	}
}

func createConfig(configPath, backend string) config.Config {
	// Try to load from config file first
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Loading configuration from: %s\n", configPath)
		if cfg, err := loadConfigFromFile(configPath); err == nil {
			return cfg
		} else {
			fmt.Printf("Warning: Failed to load config file: %v\n", err)
		}
	}

	// Fallback to command line backend
	fmt.Printf("Using backend: %s\n", backend)
	cfg := config.Config{
		Backend:              backend,
		Serializer:           "json",
		Compression:          false,
		CompressionAlgorithm: "gzip",
	}

	switch backend {
	case "memory":
		cfg.Memory = config.MemoryConfig{
			MaxSize:         "100MB",
			MaxKeys:         10000,
			EvictionPolicy:  "lru",
			DefaultTTL:      time.Hour,
			CleanupInterval: 5 * time.Minute,
		}
	case "redis":
		cfg.Redis = config.RedisConfig{
			Addresses:       []string{"localhost:6379"},
			PoolSize:        10,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolTimeout:     4 * time.Second,
			IdleTimeout:     5 * time.Minute,
			MaxRetries:      3,
			MinRetryBackoff: 100 * time.Millisecond,
			MaxRetryBackoff: 1 * time.Second,
		}
	case "memcached":
		cfg.Memcached = config.MemcachedConfig{
			Servers:      []string{"localhost:11211"},
			Timeout:      5 * time.Second,
			MaxIdleConns: 10,
		}
	}

	return cfg
}

func loadConfigFromFile(path string) (config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config.Config{}, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

func runDemo(ctx context.Context, cache gocachex.Cache) {
	fmt.Println("\nRunning demo operations...")

	// 1. Set some demo data
	demoData := map[string]interface{}{
		"demo:user:1": map[string]interface{}{
			"id":     1,
			"name":   "Alice Johnson",
			"email":  "alice@example.com",
			"role":   "developer",
			"joined": "2023-01-15",
			"active": true,
		},
		"demo:product:laptop": map[string]interface{}{
			"id":       "laptop-001",
			"name":     "Gaming Laptop",
			"price":    1299.99,
			"category": "Electronics",
			"rating":   4.5,
		},
		"demo:config:app": map[string]interface{}{
			"app_name":  "MyApp",
			"version":   "1.2.3",
			"debug":     false,
			"max_users": 1000,
			"features":  []string{"auth", "analytics", "notifications"},
		},
	}

	fmt.Println("1. Setting demo data...")
	for key, value := range demoData {
		if err := cache.Set(ctx, key, value, 10*time.Minute); err != nil {
			log.Printf("Error setting %s: %v", key, err)
		} else {
			fmt.Printf("  ✓ Set %s\n", key)
		}
	}

	// 2. Get demo data
	fmt.Println("\n2. Getting demo data...")
	for key := range demoData {
		value, err := cache.Get(ctx, key)
		if err != nil {
			log.Printf("Error getting %s: %v", key, err)
		} else {
			fmt.Printf("  ✓ %s: %+v\n", key, value)
		}
	}

	// 3. Batch operations
	fmt.Println("\n3. Batch operations...")
	keys := make([]string, 0, len(demoData))
	for key := range demoData {
		keys = append(keys, key)
	}

	values, err := cache.GetMulti(ctx, keys)
	if err != nil {
		log.Printf("Error getting multiple values: %v", err)
	} else {
		fmt.Printf("  ✓ Retrieved %d items in batch\n", len(values))
	}

	// 4. Atomic operations
	fmt.Println("\n4. Atomic operations...")
	if err := cache.Set(ctx, "demo:counter", 100, time.Hour); err != nil {
		log.Printf("Error setting counter: %v", err)
	} else {
		fmt.Println("  ✓ Set counter to 100")

		if newVal, err := cache.Increment(ctx, "demo:counter", 25); err != nil {
			log.Printf("Error incrementing counter: %v", err)
		} else {
			fmt.Printf("  ✓ Incremented by 25, new value: %d\n", newVal)
		}

		if newVal, err := cache.Decrement(ctx, "demo:counter", 10); err != nil {
			log.Printf("Error decrementing counter: %v", err)
		} else {
			fmt.Printf("  ✓ Decremented by 10, new value: %d\n", newVal)
		}
	}

	// 5. Existence checks
	fmt.Println("\n5. Existence checks...")
	for _, key := range []string{"demo:user:1", "demo:nonexistent"} {
		if exists, err := cache.Exists(ctx, key); err != nil {
			log.Printf("Error checking existence of %s: %v", key, err)
		} else {
			fmt.Printf("  %s exists: %t\n", key, exists)
		}
	}

	// 6. Stats
	fmt.Println("\n6. Cache statistics...")
	if stats, err := cache.Stats(ctx); err != nil {
		log.Printf("Error getting stats: %v", err)
	} else {
		fmt.Printf("  Hits: %d, Misses: %d, Sets: %d\n", stats.Hits, stats.Misses, stats.Sets)
		fmt.Printf("  Key Count: %d, Memory Usage: %d bytes\n", stats.KeyCount, stats.MemoryUsage)
		if stats.Hits+stats.Misses > 0 {
			hitRatio := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
			fmt.Printf("  Hit Ratio: %.2f%%\n", hitRatio)
		}
	}

	fmt.Println("\n✓ Demo completed successfully!")
}

func getValue(ctx context.Context, cache gocachex.Cache, key string) {
	value, err := cache.Get(ctx, key)
	if err != nil {
		fmt.Printf("Error getting key '%s': %v\n", key, err)
		return
	}
	fmt.Printf("Key: %s\nValue: %+v\n", key, value)
}

func setValue(ctx context.Context, cache gocachex.Cache, key, value string, ttl time.Duration, isJSON bool) {
	var val interface{}
	if isJSON {
		if err := json.Unmarshal([]byte(value), &val); err != nil {
			fmt.Printf("Error parsing JSON value: %v\n", err)
			return
		}
	} else {
		// Try to parse as number or boolean
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			val = v
		} else if v, err := strconv.ParseFloat(value, 64); err == nil {
			val = v
		} else if v, err := strconv.ParseBool(value); err == nil {
			val = v
		} else {
			val = value
		}
	}

	if err := cache.Set(ctx, key, val, ttl); err != nil {
		fmt.Printf("Error setting key '%s': %v\n", key, err)
		return
	}
	fmt.Printf("✓ Set key '%s' with TTL %v\n", key, ttl)
}

func deleteValue(ctx context.Context, cache gocachex.Cache, key string) {
	if err := cache.Delete(ctx, key); err != nil {
		fmt.Printf("Error deleting key '%s': %v\n", key, err)
		return
	}
	fmt.Printf("✓ Deleted key '%s'\n", key)
}

func checkExists(ctx context.Context, cache gocachex.Cache, key string) {
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		fmt.Printf("Error checking existence of key '%s': %v\n", key, err)
		return
	}
	fmt.Printf("Key '%s' exists: %t\n", key, exists)
}

func showStats(ctx context.Context, cache gocachex.Cache) {
	stats, err := cache.Stats(ctx)
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
		return
	}

	fmt.Println("Cache Statistics:")
	fmt.Printf("  Hits: %d\n", stats.Hits)
	fmt.Printf("  Misses: %d\n", stats.Misses)
	fmt.Printf("  Sets: %d\n", stats.Sets)
	fmt.Printf("  Deletes: %d\n", stats.Deletes)
	fmt.Printf("  Evictions: %d\n", stats.Evictions)
	fmt.Printf("  Key Count: %d\n", stats.KeyCount)
	fmt.Printf("  Memory Usage: %d bytes\n", stats.MemoryUsage)
	if stats.Hits+stats.Misses > 0 {
		hitRatio := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
		fmt.Printf("  Hit Ratio: %.2f%%\n", hitRatio)
	}
	fmt.Printf("  Uptime: %d seconds\n", stats.Uptime)
}

func clearCache(ctx context.Context, cache gocachex.Cache) {
	if err := cache.Clear(ctx); err != nil {
		fmt.Printf("Error clearing cache: %v\n", err)
		return
	}
	fmt.Println("✓ Cache cleared")
}

func showHelp() {
	fmt.Println("GoCacheX CLI - Cache Operations Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -config string    Path to config file (default \"cache_config.json\")")
	fmt.Println("  -backend string   Cache backend: memory, redis, memcached (default \"memory\")")
	fmt.Println("  -op string        Operation: demo, get, set, delete, exists, stats, clear (default \"demo\")")
	fmt.Println("  -key string       Cache key (default \"example:key\")")
	fmt.Println("  -value string     Cache value for set operation (default \"example value\")")
	fmt.Println("  -ttl string       TTL for set operation (default \"5m\")")
	fmt.Println("  -json            Treat value as JSON")
	fmt.Println("  -help            Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Run demo with memory backend")
	fmt.Println("  go run main.go -backend memory -op demo")
	fmt.Println("")
	fmt.Println("  # Set a value")
	fmt.Println("  go run main.go -op set -key \"user:1\" -value \"John Doe\" -ttl 1h")
	fmt.Println("")
	fmt.Println("  # Set JSON value")
	fmt.Println("  go run main.go -op set -key \"user:1\" -value '{\"name\":\"John\",\"age\":30}' -json")
	fmt.Println("")
	fmt.Println("  # Get a value")
	fmt.Println("  go run main.go -op get -key \"user:1\"")
	fmt.Println("")
	fmt.Println("  # Check if key exists")
	fmt.Println("  go run main.go -op exists -key \"user:1\"")
	fmt.Println("")
	fmt.Println("  # Show cache statistics")
	fmt.Println("  go run main.go -op stats")
	fmt.Println("")
	fmt.Println("  # Clear cache")
	fmt.Println("  go run main.go -op clear")
	fmt.Println("")
	fmt.Println("Configuration File Example (cache_config.json):")
	exampleConfig := config.Config{
		Backend:              "redis",
		Serializer:           "json",
		Compression:          true,
		CompressionAlgorithm: "gzip",
		Redis: config.RedisConfig{
			Addresses:   []string{"localhost:6379"},
			PoolSize:    10,
			DialTimeout: 5 * time.Second,
		},
	}

	configJSON, _ := json.MarshalIndent(exampleConfig, "  ", "  ")
	fmt.Printf("  %s\n", string(configJSON))
}
