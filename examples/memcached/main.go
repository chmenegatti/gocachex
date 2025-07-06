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
	fmt.Println("GoCacheX Memcached Example")
	fmt.Println("==========================")

	// Configure Memcached cache
	cfg := config.Config{
		Backend: "memcached",
		Memcached: config.MemcachedConfig{
			Servers:      []string{"localhost:11211"},
			Timeout:      5 * time.Second,
			MaxIdleConns: 10,
		},
		Serializer:           "json",
		Compression:          true,
		CompressionAlgorithm: "gzip",
	}

	// Create cache client
	cache, err := gocachex.New(cfg)
	if err != nil {
		log.Printf("Failed to create Memcached cache (server might not be running): %v", err)
		fmt.Println("Note: This example requires Memcached to be running on localhost:11211")
		fmt.Println("Install and start Memcached:")
		fmt.Println("  sudo apt-get install memcached  # Ubuntu/Debian")
		fmt.Println("  brew install memcached           # macOS")
		fmt.Println("  memcached -d -m 64 -p 11211")
		return
	}
	defer cache.Close()

	ctx := context.Background()

	fmt.Println("\n1. Basic Memcached Operations:")

	// Test health
	if err := cache.Health(ctx); err != nil {
		log.Printf("Memcached health check failed: %v", err)
		return
	}
	fmt.Println("✓ Memcached connection healthy")

	// Set data
	userData := map[string]interface{}{
		"id":       1,
		"username": "johndoe",
		"email":    "john@example.com",
		"profile": map[string]interface{}{
			"bio":      "Software Engineer",
			"location": "San Francisco",
			"skills":   []string{"Go", "Python", "JavaScript"},
		},
		"preferences": map[string]bool{
			"notifications": true,
			"dark_mode":     false,
			"newsletter":    true,
		},
	}

	if err := cache.Set(ctx, "user:memcached:1", userData, 10*time.Minute); err != nil {
		log.Printf("Error setting user data: %v", err)
		return
	}
	fmt.Println("✓ Set user data in Memcached")

	// Get data
	retrievedData, err := cache.Get(ctx, "user:memcached:1")
	if err != nil {
		log.Printf("Error getting user data: %v", err)
		return
	}
	fmt.Printf("✓ Retrieved user data: %+v\n", retrievedData)

	fmt.Println("\n2. Batch Operations:")

	// Set multiple products
	products := map[string]interface{}{
		"product:laptop": map[string]interface{}{
			"name":     "Gaming Laptop",
			"price":    1299.99,
			"category": "Electronics",
			"specs":    map[string]string{"cpu": "Intel i7", "ram": "16GB", "storage": "512GB SSD"},
			"in_stock": true,
		},
		"product:mouse": map[string]interface{}{
			"name":     "Wireless Mouse",
			"price":    49.99,
			"category": "Accessories",
			"specs":    map[string]string{"type": "wireless", "dpi": "1200", "buttons": "3"},
			"in_stock": true,
		},
		"product:keyboard": map[string]interface{}{
			"name":     "Mechanical Keyboard",
			"price":    129.99,
			"category": "Accessories",
			"specs":    map[string]string{"type": "mechanical", "layout": "qwerty", "backlight": "rgb"},
			"in_stock": false,
		},
	}

	if err := cache.SetMulti(ctx, products, 15*time.Minute); err != nil {
		log.Printf("Error setting product data: %v", err)
		return
	}
	fmt.Println("✓ Set multiple products")

	// Get multiple products
	productKeys := []string{"product:laptop", "product:mouse", "product:keyboard"}
	retrievedProducts, err := cache.GetMulti(ctx, productKeys)
	if err != nil {
		log.Printf("Error getting products: %v", err)
		return
	}

	fmt.Println("✓ Retrieved products:")
	for key, product := range retrievedProducts {
		fmt.Printf("  %s: %+v\n", key, product)
	}

	fmt.Println("\n3. Atomic Operations:")

	// Initialize counters
	if err := cache.Set(ctx, "counter:views", 100, time.Hour); err != nil {
		log.Printf("Error setting counter: %v", err)
		return
	}
	fmt.Println("✓ Initialized page view counter")

	// Increment counter
	newValue, err := cache.Increment(ctx, "counter:views", 5)
	if err != nil {
		log.Printf("Error incrementing counter: %v", err)
		return
	}
	fmt.Printf("✓ Incremented counter by 5, new value: %d\n", newValue)

	// Decrement counter
	newValue, err = cache.Decrement(ctx, "counter:views", 2)
	if err != nil {
		log.Printf("Error decrementing counter: %v", err)
		return
	}
	fmt.Printf("✓ Decremented counter by 2, new value: %d\n", newValue)

	fmt.Println("\n4. Key Management:")

	// Check existence
	exists, err := cache.Exists(ctx, "user:memcached:1")
	if err != nil {
		log.Printf("Error checking existence: %v", err)
		return
	}
	fmt.Printf("✓ User key exists: %t\n", exists)

	// Set with expiration
	tempData := map[string]string{"temp": "data", "expires": "in 3 seconds"}
	if err := cache.Set(ctx, "temp:data", tempData, 3*time.Second); err != nil {
		log.Printf("Error setting temp data: %v", err)
		return
	}
	fmt.Println("✓ Set temporary data with 3-second TTL")

	// Check TTL
	ttl, err := cache.TTL(ctx, "temp:data")
	if err != nil {
		log.Printf("Error getting TTL: %v", err)
		return
	}
	fmt.Printf("✓ Temporary data TTL: %v\n", ttl)

	// Wait for expiration
	fmt.Println("⏳ Waiting 4 seconds for expiration...")
	time.Sleep(4 * time.Second)

	// Check if expired
	exists, err = cache.Exists(ctx, "temp:data")
	if err != nil {
		log.Printf("Error checking existence after expiration: %v", err)
		return
	}
	fmt.Printf("✓ Temporary data exists after expiration: %t\n", exists)

	fmt.Println("\n5. Cache Statistics:")

	// Get cache statistics
	stats, err := cache.Stats(ctx)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		return
	}

	fmt.Printf("Cache Statistics:\n")
	fmt.Printf("  Hits: %d\n", stats.Hits)
	fmt.Printf("  Misses: %d\n", stats.Misses)
	fmt.Printf("  Sets: %d\n", stats.Sets)
	fmt.Printf("  Deletes: %d\n", stats.Deletes)
	fmt.Printf("  Key Count: %d\n", stats.KeyCount)
	fmt.Printf("  Memory Usage: %d bytes\n", stats.MemoryUsage)
	if stats.Hits+stats.Misses > 0 {
		hitRatio := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
		fmt.Printf("  Hit Ratio: %.2f%%\n", hitRatio)
	}
	fmt.Printf("  Uptime: %d seconds\n", stats.Uptime)

	fmt.Println("\n6. Cleanup:")

	// Delete specific keys
	if err := cache.Delete(ctx, "user:memcached:1"); err != nil {
		log.Printf("Error deleting user: %v", err)
	} else {
		fmt.Println("✓ Deleted user data")
	}

	// Delete multiple keys
	deleteKeys := []string{"product:laptop", "product:mouse", "product:keyboard"}
	if err := cache.DeleteMulti(ctx, deleteKeys); err != nil {
		log.Printf("Error deleting products: %v", err)
	} else {
		fmt.Println("✓ Deleted product data")
	}

	fmt.Println("\n✓ Memcached example completed successfully!")
	fmt.Println("Note: In production, configure connection pooling, timeouts, and error handling")
	fmt.Println("      according to your application's requirements.")
}
