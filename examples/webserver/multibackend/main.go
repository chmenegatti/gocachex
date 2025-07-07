// Package main demonstrates different cache backends in a web server.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/pkg/config"
)

// Product represents a product model
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

// MultiBackendServer demonstrates using multiple cache backends
type MultiBackendServer struct {
	memoryCache      gocachex.Cache // Fast L1 cache
	distributedCache gocachex.Cache // Redis L2 cache (if available)
	products         map[int]*Product
	ctx              context.Context
}

// NewMultiBackendServer creates a server with multiple cache backends
func NewMultiBackendServer() *MultiBackendServer {
	// L1 Cache - Memory (fast, limited size)
	memoryCache, err := gocachex.New(config.Config{
		Backend: "memory",
		Memory: config.MemoryConfig{
			MaxSize:         "50MB",
			MaxKeys:         500,
			EvictionPolicy:  "lru",
			DefaultTTL:      2 * time.Minute,
			CleanupInterval: 30 * time.Second,
		},
		Compression:          true,
		CompressionAlgorithm: "gzip",
		Serializer:           "json",
	})
	if err != nil {
		log.Fatalf("Failed to create memory cache: %v", err)
	}

	// L2 Cache - Redis (larger, distributed)
	// Comment this out if Redis is not available
	var distributedCache gocachex.Cache
	redisCache, err := gocachex.New(config.Config{
		Backend: "redis",
		Redis: config.RedisConfig{
			Addresses:    []string{"localhost:6379"},
			Password:     "",
			DB:           0,
			PoolSize:     10,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Compression:          true,
		CompressionAlgorithm: "gzip",
		Serializer:           "json",
	})
	if err != nil {
		log.Printf("Redis not available, using memory cache only: %v", err)
		distributedCache = nil
	} else {
		distributedCache = redisCache
		log.Println("Using Redis as L2 cache")
	}

	// Sample products
	products := map[int]*Product{
		1: {ID: 1, Name: "Laptop", Price: 999.99, Category: "Electronics", Description: "High-performance laptop", Created: time.Now()},
		2: {ID: 2, Name: "Mouse", Price: 29.99, Category: "Electronics", Description: "Wireless mouse", Created: time.Now()},
		3: {ID: 3, Name: "Book", Price: 19.99, Category: "Books", Description: "Programming guide", Created: time.Now()},
		4: {ID: 4, Name: "Coffee", Price: 4.99, Category: "Food", Description: "Premium coffee beans", Created: time.Now()},
		5: {ID: 5, Name: "Headphones", Price: 199.99, Category: "Electronics", Description: "Noise-canceling headphones", Created: time.Now()},
	}

	return &MultiBackendServer{
		memoryCache:      memoryCache,
		distributedCache: distributedCache,
		products:         products,
		ctx:              context.Background(),
	}
}

// getProduct implements hierarchical caching: L1 (memory) -> L2 (redis) -> database
func (s *MultiBackendServer) getProduct(id int) (*Product, string, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	// Try L1 cache first (memory)
	if product, err := s.memoryCache.Get(s.ctx, cacheKey); err == nil {
		log.Printf("L1 Cache HIT for product %d", id)
		return product.(*Product), "L1-HIT", nil
	}
	log.Printf("L1 Cache MISS for product %d", id)

	// Try L2 cache (Redis) if available
	if s.distributedCache != nil {
		if product, err := s.distributedCache.Get(s.ctx, cacheKey); err == nil {
			log.Printf("L2 Cache HIT for product %d", id)
			// Store in L1 for faster access next time
			s.memoryCache.Set(s.ctx, cacheKey, product, 2*time.Minute)
			return product.(*Product), "L2-HIT", nil
		}
		log.Printf("L2 Cache MISS for product %d", id)
	}

	// Get from "database"
	product, exists := s.products[id]
	if !exists {
		return nil, "MISS", fmt.Errorf("product not found")
	}

	// Store in both L1 and L2 caches
	s.memoryCache.Set(s.ctx, cacheKey, product, 2*time.Minute)
	if s.distributedCache != nil {
		s.distributedCache.Set(s.ctx, cacheKey, product, 10*time.Minute)
	}

	log.Printf("Database access for product %d", id)
	return product, "DB", nil
}

// invalidateProduct removes from all cache levels
func (s *MultiBackendServer) invalidateProduct(id int) {
	cacheKey := fmt.Sprintf("product:%d", id)
	s.memoryCache.Delete(s.ctx, cacheKey)
	if s.distributedCache != nil {
		s.distributedCache.Delete(s.ctx, cacheKey)
	}
}

// extractProductID extracts product ID from URL
func extractProductID(urlPath string) (int, error) {
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	if len(parts) < 4 || parts[3] == "" {
		return 0, fmt.Errorf("no product ID provided")
	}
	return strconv.Atoi(parts[3])
}

// getProductHandler handles GET /products/{id}
func (s *MultiBackendServer) getProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractProductID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, cacheLevel, err := s.getProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", cacheLevel)
	json.NewEncoder(w).Encode(product)
}

// getAllProductsHandler handles GET /products
func (s *MultiBackendServer) getAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	cacheKey := "products:all"

	// Try L1 cache first
	if products, err := s.memoryCache.Get(s.ctx, cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "L1-HIT")
		json.NewEncoder(w).Encode(products)
		return
	}

	// Try L2 cache
	if s.distributedCache != nil {
		if products, err := s.distributedCache.Get(s.ctx, cacheKey); err == nil {
			// Store in L1
			s.memoryCache.Set(s.ctx, cacheKey, products, time.Minute)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "L2-HIT")
			json.NewEncoder(w).Encode(products)
			return
		}
	}

	// Get from database
	products := make([]*Product, 0, len(s.products))
	for _, product := range s.products {
		products = append(products, product)
	}

	// Cache in both levels
	s.memoryCache.Set(s.ctx, cacheKey, products, time.Minute)
	if s.distributedCache != nil {
		s.distributedCache.Set(s.ctx, cacheKey, products, 5*time.Minute)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "DB")
	json.NewEncoder(w).Encode(products)
}

// updateProductHandler handles PUT /products/{id}
func (s *MultiBackendServer) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractProductID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var updatedProduct Product
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if _, exists := s.products[id]; !exists {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	updatedProduct.ID = id
	s.products[id] = &updatedProduct

	// Invalidate caches
	s.invalidateProduct(id)
	s.memoryCache.Delete(s.ctx, "products:all")
	if s.distributedCache != nil {
		s.distributedCache.Delete(s.ctx, "products:all")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&updatedProduct)
}

// cacheStatsHandler provides detailed cache statistics
func (s *MultiBackendServer) cacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	l1Stats, _ := s.memoryCache.Stats(s.ctx)

	var l2Stats *gocachex.Stats
	if s.distributedCache != nil {
		l2Stats, _ = s.distributedCache.Stats(s.ctx)
	}

	response := map[string]interface{}{
		"l1_cache":  l1Stats,
		"l2_cache":  l2Stats,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// productsRouter handles product routes
func (s *MultiBackendServer) productsRouter(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(r.URL.Path, "/api/v1")

	switch {
	case urlPath == "/products" && r.Method == "GET":
		s.getAllProductsHandler(w, r)
	case strings.HasPrefix(urlPath, "/products/") && r.Method == "GET":
		s.getProductHandler(w, r)
	case strings.HasPrefix(urlPath, "/products/") && r.Method == "PUT":
		s.updateProductHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	server := NewMultiBackendServer()

	// Routes
	http.HandleFunc("/api/v1/products", server.productsRouter)
	http.HandleFunc("/api/v1/products/", server.productsRouter)
	http.HandleFunc("/api/v1/cache/stats", server.cacheStatsHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if strings.HasPrefix(r.URL.Path, "/api/v1/products") {
			server.productsRouter(w, r)
		} else if r.URL.Path == "/api/v1/cache/stats" {
			server.cacheStatsHandler(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}

		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})

	port := ":8081"
	log.Printf("Starting multi-backend server on port %s", port)
	log.Printf("Cache backends:")
	log.Printf("  L1: Memory (fast, limited)")
	if server.distributedCache != nil {
		log.Printf("  L2: Redis (slower, distributed)")
	} else {
		log.Printf("  L2: Not available (Redis not connected)")
	}
	log.Printf("")
	log.Printf("API endpoints:")
	log.Printf("  GET /api/v1/products          - Get all products")
	log.Printf("  GET /api/v1/products/{id}     - Get product by ID")
	log.Printf("  PUT /api/v1/products/{id}     - Update product")
	log.Printf("  GET /api/v1/cache/stats       - Cache statistics")
	log.Printf("")
	log.Printf("Test commands:")
	log.Printf("  curl http://localhost:8081/api/v1/products")
	log.Printf("  curl http://localhost:8081/api/v1/products/1")
	log.Printf("  curl http://localhost:8081/api/v1/cache/stats")

	log.Fatal(http.ListenAndServe(port, nil))
}
