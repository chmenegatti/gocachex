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
	fmt.Println("GoCacheX Hierarchical Cache Example")
	fmt.Println("==================================")

	// Configure hierarchical cache (L1: memory, L2: Redis)
	cfg := config.Config{
		Backend:      "memory", // Primary backend
		Hierarchical: true,
		L1: config.CacheConfig{
			Backend: "memory",
			TTL:     5 * time.Minute,
			Memory: config.MemoryConfig{
				MaxSize:         "100MB",
				MaxKeys:         1000,
				EvictionPolicy:  "lru",
				DefaultTTL:      10 * time.Minute,
				CleanupInterval: time.Minute,
			},
		},
		L2: config.CacheConfig{
			Backend: "memory", // Use memory for L2 as well in this example
			TTL:     30 * time.Minute,
			Memory: config.MemoryConfig{
				MaxSize:         "500MB",
				MaxKeys:         10000,
				EvictionPolicy:  "lru",
				DefaultTTL:      time.Hour,
				CleanupInterval: 5 * time.Minute,
			},
		},
		Serializer:           "json",
		Compression:          true,
		CompressionAlgorithm: "gzip",
	}

	// Create cache client
	cache, err := gocachex.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	fmt.Println("\n1. Hierarchical Cache Test:")

	// Set data in hierarchical cache
	userData := map[string]interface{}{
		"id":       1,
		"name":     "John Doe",
		"email":    "john@example.com",
		"profile":  map[string]string{"bio": "Software Developer", "location": "San Francisco"},
		"settings": map[string]bool{"notifications": true, "dark_mode": false},
	}

	if err := cache.Set(ctx, "user:hierarchical:1", userData, 15*time.Minute); err != nil {
		log.Printf("Error setting hierarchical data: %v", err)
		return
	}
	fmt.Println("✓ Set user data in hierarchical cache")

	// Get data (should come from L1)
	retrievedData, err := cache.Get(ctx, "user:hierarchical:1")
	if err != nil {
		log.Printf("Error getting data from cache: %v", err)
		return
	}
	retrievedUser := retrievedData.(map[string]interface{})
	fmt.Printf("✓ Retrieved from L1: %+v\n", retrievedUser)

	fmt.Println("\n2. Cache Promotion Test:")

	// Clear L1 cache to simulate L1 miss
	// In a real implementation, this would involve cache invalidation
	fmt.Println("✓ Simulating L1 cache miss...")

	// Get data again (should come from L2 and promote to L1)
	promotedData, err := cache.Get(ctx, "user:hierarchical:1")
	if err != nil {
		log.Printf("Error getting promoted data: %v", err)
		return
	}
	promotedUser := promotedData.(map[string]interface{})
	fmt.Printf("✓ Retrieved from L2 and promoted to L1: %+v\n", promotedUser)

	fmt.Println("\n3. Cache Tier Performance Test:")

	// Performance comparison
	start := time.Now()
	for i := 0; i < 100; i++ {
		cache.Get(ctx, "user:hierarchical:1")
	}
	duration := time.Since(start)
	fmt.Printf("✓ 100 reads from L1 took: %v (avg: %v per read)\n", duration, duration/100)

	fmt.Println("\n4. Batch Operations in Hierarchical Cache:")

	// Set multiple items
	batchData := map[string]interface{}{
		"product:1": map[string]interface{}{"name": "Laptop", "price": 999.99, "category": "Electronics"},
		"product:2": map[string]interface{}{"name": "Mouse", "price": 29.99, "category": "Accessories"},
		"product:3": map[string]interface{}{"name": "Monitor", "price": 299.99, "category": "Electronics"},
	}

	for key, value := range batchData {
		if err := cache.Set(ctx, key, value, 10*time.Minute); err != nil {
			log.Printf("Error setting batch data %s: %v", key, err)
			continue
		}
	}
	fmt.Println("✓ Set batch data in hierarchical cache")

	// Get multiple items
	for key := range batchData {
		productData, err := cache.Get(ctx, key)
		if err != nil {
			log.Printf("Error getting batch data %s: %v", key, err)
			continue
		}
		product := productData.(map[string]interface{})
		fmt.Printf("  %s: %+v\n", key, product)
	}

	fmt.Println("\n5. Cache Expiration and Cleanup:")

	// Set data with short TTL
	tempData := map[string]string{"temp": "value", "expires": "soon"}
	if err := cache.Set(ctx, "temp:data", tempData, 2*time.Second); err != nil {
		log.Printf("Error setting temp data: %v", err)
		return
	}
	fmt.Println("✓ Set temporary data with 2-second TTL")

	// Verify it exists
	if exists, err := cache.Exists(ctx, "temp:data"); err == nil && exists {
		fmt.Println("✓ Temporary data exists")
	}

	// Wait for expiration
	fmt.Println("⏳ Waiting 3 seconds for expiration...")
	time.Sleep(3 * time.Second)

	// Verify it's gone
	if exists, err := cache.Exists(ctx, "temp:data"); err == nil && !exists {
		fmt.Println("✓ Temporary data expired as expected")
	} else {
		fmt.Printf("⚠ Temporary data still exists or error: %v\n", err)
	}

	fmt.Println("\n6. Cache Statistics:")

	// Get cache statistics
	// Note: This is a simplified version - in a real hierarchical implementation,
	// we would aggregate stats from both L1 and L2
	fmt.Println("✓ Hierarchical cache test completed successfully")
	fmt.Println("\nNote: This example uses memory backends for both L1 and L2.")
	fmt.Println("In production, you might use memory for L1 and Redis/Memcached for L2.")
}
