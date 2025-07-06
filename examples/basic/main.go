// Package main demonstrates basic usage of GoCacheX.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/pkg/config"
)

func main() {
	fmt.Println("GoCacheX Basic Example")
	fmt.Println("======================")

	// Example 1: Basic in-memory cache
	fmt.Println("\n1. Basic In-Memory Cache:")
	basicMemoryExample()

	// Example 2: Cache with TTL
	fmt.Println("\n2. Cache with TTL:")
	ttlExample()

	// Example 3: Batch operations
	fmt.Println("\n3. Batch Operations:")
	batchExample()

	// Example 4: Atomic operations
	fmt.Println("\n4. Atomic Operations:")
	atomicExample()

	// Example 5: Cache statistics
	fmt.Println("\n5. Cache Statistics:")
	statsExample()
}

// basicMemoryExample demonstrates basic cache operations
func basicMemoryExample() {
	// Create a simple in-memory cache
	cache, err := gocachex.New(config.Config{
		Backend: "memory",
		Memory: config.MemoryConfig{
			MaxSize:         "10MB",
			EvictionPolicy:  "lru",
			CleanupInterval: 30 * time.Second,
		},
		Serializer: "json",
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Set some values
	fmt.Println("Setting values...")
	cache.Set(ctx, "user:1", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}, 5*time.Minute)

	cache.Set(ctx, "user:2", map[string]interface{}{
		"name":  "Jane Smith",
		"email": "jane@example.com",
		"age":   25,
	}, 5*time.Minute)

	cache.Set(ctx, "config:timeout", 30, 10*time.Minute)

	// Get values
	fmt.Println("Getting values...")
	if user1, err := cache.Get(ctx, "user:1"); err == nil {
		fmt.Printf("user:1 = %+v\n", user1)
	}

	if timeout, err := cache.Get(ctx, "config:timeout"); err == nil {
		fmt.Printf("config:timeout = %v\n", timeout)
	}

	// Check existence
	if exists, _ := cache.Exists(ctx, "user:1"); exists {
		fmt.Println("user:1 exists in cache")
	}

	// Delete a value
	cache.Delete(ctx, "user:2")
	if exists, _ := cache.Exists(ctx, "user:2"); !exists {
		fmt.Println("user:2 was deleted from cache")
	}
}

// ttlExample demonstrates TTL functionality
func ttlExample() {
	cache, err := gocachex.New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Set value with short TTL
	fmt.Println("Setting value with 2 second TTL...")
	cache.Set(ctx, "temp:data", "This will expire soon", 2*time.Second)

	// Check immediately
	if value, err := cache.Get(ctx, "temp:data"); err == nil {
		fmt.Printf("Immediately after set: %v\n", value)
	}

	// Wait and check again
	fmt.Println("Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	if _, err := cache.Get(ctx, "temp:data"); err != nil {
		fmt.Println("Value expired as expected")
	}
}

// batchExample demonstrates batch operations
func batchExample() {
	cache, err := gocachex.New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values at once
	fmt.Println("Setting multiple values...")
	items := map[string]interface{}{
		"product:1": map[string]interface{}{"name": "Laptop", "price": 999},
		"product:2": map[string]interface{}{"name": "Mouse", "price": 25},
		"product:3": map[string]interface{}{"name": "Keyboard", "price": 75},
	}
	cache.SetMulti(ctx, items, 5*time.Minute)

	// Get multiple values at once
	fmt.Println("Getting multiple values...")
	keys := []string{"product:1", "product:2", "product:3", "product:4"}
	results, err := cache.GetMulti(ctx, keys)
	if err == nil {
		for key, value := range results {
			fmt.Printf("%s = %+v\n", key, value)
		}
	}

	// Delete multiple values
	fmt.Println("Deleting multiple values...")
	cache.DeleteMulti(ctx, []string{"product:1", "product:2"})

	// Verify deletion
	if exists, _ := cache.Exists(ctx, "product:1"); !exists {
		fmt.Println("product:1 was deleted")
	}
	if exists, _ := cache.Exists(ctx, "product:3"); exists {
		fmt.Println("product:3 still exists")
	}
}

// atomicExample demonstrates atomic operations
func atomicExample() {
	cache, err := gocachex.New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Initialize counter
	fmt.Println("Initializing counter...")
	cache.Set(ctx, "counter", 10, 10*time.Minute)

	// Increment counter
	fmt.Println("Incrementing counter...")
	newValue, err := cache.Increment(ctx, "counter", 5)
	if err == nil {
		fmt.Printf("Counter after increment: %d\n", newValue)
	}

	// Decrement counter
	fmt.Println("Decrementing counter...")
	newValue, err = cache.Decrement(ctx, "counter", 3)
	if err == nil {
		fmt.Printf("Counter after decrement: %d\n", newValue)
	}

	// SetNX (set if not exists)
	fmt.Println("Testing SetNX...")
	success, err := cache.SetNX(ctx, "counter", 100, 5*time.Minute)
	if err == nil {
		fmt.Printf("SetNX on existing key succeeded: %t\n", success)
	}

	success, err = cache.SetNX(ctx, "new_key", "new_value", 5*time.Minute)
	if err == nil {
		fmt.Printf("SetNX on new key succeeded: %t\n", success)
	}
}

// statsExample demonstrates cache statistics
func statsExample() {
	cache, err := gocachex.New(config.Config{
		Backend:    "memory",
		Serializer: "json",
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Perform some operations to generate stats
	cache.Set(ctx, "key1", "value1", 5*time.Minute)
	cache.Set(ctx, "key2", "value2", 5*time.Minute)
	cache.Get(ctx, "key1")    // Hit
	cache.Get(ctx, "key3")    // Miss
	cache.Delete(ctx, "key2") // Delete

	// Get statistics
	stats, err := cache.Stats(ctx)
	if err == nil {
		fmt.Printf("Cache Statistics:\n")
		fmt.Printf("  Hits: %d\n", stats.Hits)
		fmt.Printf("  Misses: %d\n", stats.Misses)
		fmt.Printf("  Sets: %d\n", stats.Sets)
		fmt.Printf("  Deletes: %d\n", stats.Deletes)
		fmt.Printf("  Key Count: %d\n", stats.KeyCount)
		fmt.Printf("  Memory Usage: %d bytes\n", stats.MemoryUsage)
		fmt.Printf("  Hit Ratio: %.2f%%\n", stats.HitRatio()*100)
		fmt.Printf("  Uptime: %d seconds\n", stats.Uptime)
	}

	// Test health check
	if err := cache.Health(ctx); err == nil {
		fmt.Println("Cache health check: OK")
	}
}
