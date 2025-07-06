// Package main demonstrates Redis backend usage with GoCacheX.
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
	fmt.Println("GoCacheX Redis Example")
	fmt.Println("======================")

	// Note: This example requires Redis to be running on localhost:6379
	// Start Redis with: docker run -d --name redis -p 6379:6379 redis:alpine

	cache, err := gocachex.New(config.Config{
		Backend: "redis",
		Redis: config.RedisConfig{
			Addresses: []string{"localhost:6379"},
			Password:  "",
			DB:        0,
			PoolSize:  10,
		},
		Serializer:  "json",
		Compression: true,
	})
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Test basic operations
	fmt.Println("\n1. Basic Operations:")
	testBasicOperations(ctx, cache)

	// Test compression
	fmt.Println("\n2. Compression Test:")
	testCompression(ctx, cache)

	// Test persistence
	fmt.Println("\n3. Persistence Test:")
	testPersistence(ctx, cache)
}

func testBasicOperations(ctx context.Context, cache gocachex.Cache) {
	// Set values
	user := map[string]interface{}{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
		"settings": map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
		},
	}

	err := cache.Set(ctx, "user:1", user, 10*time.Minute)
	if err != nil {
		log.Printf("Error setting user: %v", err)
		return
	}
	fmt.Println("✓ Set complex user object")

	// Get value
	if retrievedUser, err := cache.Get(ctx, "user:1"); err == nil {
		fmt.Printf("✓ Retrieved user: %+v\n", retrievedUser)
	} else {
		log.Printf("Error getting user: %v", err)
	}

	// Atomic operations
	if newVal, err := cache.Increment(ctx, "page_views", 1); err == nil {
		fmt.Printf("✓ Page views: %d\n", newVal)
	}

	// Batch operations
	products := map[string]interface{}{
		"product:laptop":  map[string]interface{}{"name": "Gaming Laptop", "price": 1299.99},
		"product:mouse":   map[string]interface{}{"name": "Wireless Mouse", "price": 59.99},
		"product:monitor": map[string]interface{}{"name": "4K Monitor", "price": 399.99},
	}

	if err := cache.SetMulti(ctx, products, 15*time.Minute); err == nil {
		fmt.Println("✓ Set multiple products")
	}

	keys := []string{"product:laptop", "product:mouse", "product:monitor"}
	if results, err := cache.GetMulti(ctx, keys); err == nil {
		fmt.Printf("✓ Retrieved %d products\n", len(results))
	}
}

func testCompression(ctx context.Context, cache gocachex.Cache) {
	// Large JSON data that benefits from compression
	largeData := map[string]interface{}{
		"description": "This is a very long description that repeats many times. " +
			"This is a very long description that repeats many times. " +
			"This is a very long description that repeats many times. " +
			"This is a very long description that repeats many times.",
		"items": make([]map[string]interface{}, 100),
	}

	// Fill with repetitive data
	for i := 0; i < 100; i++ {
		largeData["items"].([]map[string]interface{})[i] = map[string]interface{}{
			"id":          i,
			"name":        fmt.Sprintf("Item %d", i),
			"description": "Standard item description that repeats for all items",
			"category":    "electronics",
			"tags":        []string{"popular", "electronics", "gadget"},
		}
	}

	err := cache.Set(ctx, "large_data", largeData, 5*time.Minute)
	if err != nil {
		log.Printf("Error setting large data: %v", err)
		return
	}
	fmt.Println("✓ Stored large data with compression")

	if _, err := cache.Get(ctx, "large_data"); err == nil {
		fmt.Println("✓ Retrieved compressed large data")
	} else {
		log.Printf("Error getting large data: %v", err)
	}
}

func testPersistence(ctx context.Context, cache gocachex.Cache) {
	// Set a value that should persist
	sessionData := map[string]interface{}{
		"user_id":     123,
		"login_time":  time.Now().Unix(),
		"permissions": []string{"read", "write", "admin"},
	}

	err := cache.Set(ctx, "session:abc123", sessionData, 30*time.Minute)
	if err != nil {
		log.Printf("Error setting session: %v", err)
		return
	}
	fmt.Println("✓ Set session data")

	// Simulate application restart by creating a new cache instance
	newCache, err := gocachex.New(config.Config{
		Backend: "redis",
		Redis: config.RedisConfig{
			Addresses: []string{"localhost:6379"},
			Password:  "",
			DB:        0,
		},
		Serializer: "json",
	})
	if err != nil {
		log.Printf("Error creating new cache instance: %v", err)
		return
	}
	defer newCache.Close()

	// Check if data persisted
	if retrievedSession, err := newCache.Get(ctx, "session:abc123"); err == nil {
		fmt.Printf("✓ Session data persisted: %+v\n", retrievedSession)
	} else {
		log.Printf("Error: Session data not found after restart: %v", err)
	}

	// Clean up
	newCache.Delete(ctx, "session:abc123")
	fmt.Println("✓ Cleaned up session data")
}
