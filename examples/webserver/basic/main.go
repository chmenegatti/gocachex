// Package main demonstrates how to use GoCacheX in a web server application.
// This example shows various caching strategies for different types of web requests.
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

// User represents a simple user model
type User struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
}

// WebServer represents our web server with cache
type WebServer struct {
	cache gocachex.Cache
	users map[int]*User // Simulated database
	ctx   context.Context
}

// NewWebServer creates a new web server instance with cache
func NewWebServer() *WebServer {
	// Configure cache - you can switch between different backends
	cache, err := gocachex.New(config.Config{
		Backend: "memory",
		Memory: config.MemoryConfig{
			MaxSize:         "100MB",
			MaxKeys:         1000,
			EvictionPolicy:  "lru",
			DefaultTTL:      5 * time.Minute,
			CleanupInterval: time.Minute,
		},
		// Uncomment to use Redis instead
		// Backend: "redis",
		// Redis: config.RedisConfig{
		//     Addresses: []string{"localhost:6379"},
		//     Password:  "",
		//     DB:        0,
		//     PoolSize:  10,
		// },
		Compression:          true,
		CompressionAlgorithm: "gzip",
		Serializer:           "json",
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// Initialize with some sample data
	users := map[int]*User{
		1: {ID: 1, Name: "Alice", Email: "alice@example.com", Created: time.Now()},
		2: {ID: 2, Name: "Bob", Email: "bob@example.com", Created: time.Now()},
		3: {ID: 3, Name: "Charlie", Email: "charlie@example.com", Created: time.Now()},
	}

	return &WebServer{
		cache: cache,
		users: users,
		ctx:   context.Background(),
	}
}

// extractUserID extracts user ID from URL path
func extractUserID(urlPath string) (int, error) {
	// Extract ID from path like /api/v1/users/123
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	if len(parts) < 4 || parts[3] == "" {
		return 0, fmt.Errorf("no user ID provided")
	}
	return strconv.Atoi(parts[3])
}

// getUserHandler handles GET /users/{id} requests with caching
func (ws *WebServer) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("user:%d", userID)

	// Try to get from cache first
	cachedUser, err := ws.cache.Get(ws.ctx, cacheKey)
	if err == nil {
		log.Printf("Cache HIT for user %d", userID)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cachedUser)
		return
	}

	log.Printf("Cache MISS for user %d", userID)

	// Get from "database" (our map)
	user, exists := ws.users[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Cache the user for 5 minutes
	err = ws.cache.Set(ws.ctx, cacheKey, user, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to cache user %d: %v", userID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(user)
}

// createUserHandler handles POST /users requests
func (ws *WebServer) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate new ID
	newID := len(ws.users) + 1
	user.ID = newID
	user.Created = time.Now()

	// Save to "database"
	ws.users[newID] = &user

	// Cache the new user
	cacheKey := fmt.Sprintf("user:%d", newID)
	err := ws.cache.Set(ws.ctx, cacheKey, &user, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to cache new user %d: %v", newID, err)
	}

	// Invalidate users list cache
	ws.cache.Delete(ws.ctx, "users:all")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&user)
}

// updateUserHandler handles PUT /users/{id} requests
func (ws *WebServer) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updatedUser User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if user exists
	if _, exists := ws.users[userID]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update user
	updatedUser.ID = userID
	ws.users[userID] = &updatedUser

	// Update cache
	cacheKey := fmt.Sprintf("user:%d", userID)
	err = ws.cache.Set(ws.ctx, cacheKey, &updatedUser, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to update cache for user %d: %v", userID, err)
	}

	// Invalidate users list cache
	ws.cache.Delete(ws.ctx, "users:all")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&updatedUser)
}

// deleteUserHandler handles DELETE /users/{id} requests
func (ws *WebServer) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if user exists
	if _, exists := ws.users[userID]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Delete from "database"
	delete(ws.users, userID)

	// Remove from cache
	cacheKey := fmt.Sprintf("user:%d", userID)
	ws.cache.Delete(ws.ctx, cacheKey)

	// Invalidate users list cache
	ws.cache.Delete(ws.ctx, "users:all")

	w.WriteHeader(http.StatusNoContent)
}

// getAllUsersHandler handles GET /users requests with caching
func (ws *WebServer) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	cacheKey := "users:all"

	// Try to get from cache first
	cachedUsers, err := ws.cache.Get(ws.ctx, cacheKey)
	if err == nil {
		log.Println("Cache HIT for all users")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cachedUsers)
		return
	}

	log.Println("Cache MISS for all users")

	// Get all users from "database"
	users := make([]*User, 0, len(ws.users))
	for _, user := range ws.users {
		users = append(users, user)
	}

	// Cache the users list for 2 minutes (shorter TTL for lists)
	err = ws.cache.Set(ws.ctx, cacheKey, users, 2*time.Minute)
	if err != nil {
		log.Printf("Failed to cache users list: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(users)
}

// cacheStatsHandler provides cache statistics
func (ws *WebServer) cacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := ws.cache.Stats(ws.ctx)
	if err != nil {
		http.Error(w, "Failed to get cache stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// clearCacheHandler clears all cache
func (ws *WebServer) clearCacheHandler(w http.ResponseWriter, r *http.Request) {
	err := ws.cache.Clear(ws.ctx)
	if err != nil {
		http.Error(w, "Failed to clear cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cache cleared successfully"))
}

// healthHandler provides health check
func (ws *WebServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check cache health
	exists, err := ws.cache.Exists(ws.ctx, "health-check")
	if err != nil {
		http.Error(w, "Cache unhealthy", http.StatusServiceUnavailable)
		return
	}

	// Set a test key if it doesn't exist
	if !exists {
		ws.cache.Set(ws.ctx, "health-check", "ok", time.Minute)
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"cache":     "connected",
		"users":     len(ws.users),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// usersRouter handles all /users routes
func (ws *WebServer) usersRouter(w http.ResponseWriter, r *http.Request) {
	// Remove /api/v1 prefix
	urlPath := strings.TrimPrefix(r.URL.Path, "/api/v1")

	switch {
	case urlPath == "/users" && r.Method == "GET":
		ws.getAllUsersHandler(w, r)
	case urlPath == "/users" && r.Method == "POST":
		ws.createUserHandler(w, r)
	case strings.HasPrefix(urlPath, "/users/") && r.Method == "GET":
		ws.getUserHandler(w, r)
	case strings.HasPrefix(urlPath, "/users/") && r.Method == "PUT":
		ws.updateUserHandler(w, r)
	case strings.HasPrefix(urlPath, "/users/") && r.Method == "DELETE":
		ws.deleteUserHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// cacheRouter handles all /cache routes
func (ws *WebServer) cacheRouter(w http.ResponseWriter, r *http.Request) {
	// Remove /api/v1 prefix
	urlPath := strings.TrimPrefix(r.URL.Path, "/api/v1")

	switch {
	case urlPath == "/cache/stats" && r.Method == "GET":
		ws.cacheStatsHandler(w, r)
	case urlPath == "/cache/clear" && r.Method == "POST":
		ws.clearCacheHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// Create web server instance
	ws := NewWebServer()

	// Set up routes
	http.HandleFunc("/api/v1/users", ws.usersRouter)
	http.HandleFunc("/api/v1/users/", ws.usersRouter)
	http.HandleFunc("/api/v1/cache/", ws.cacheRouter)
	http.HandleFunc("/health", ws.healthHandler)

	// Add logging middleware
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Check if route exists
		if strings.HasPrefix(r.URL.Path, "/api/v1/users") {
			ws.usersRouter(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/v1/cache/") {
			ws.cacheRouter(w, r)
		} else if r.URL.Path == "/health" {
			ws.healthHandler(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}

		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})

	// Start server
	port := ":8080"
	log.Printf("Starting web server on port %s", port)
	log.Printf("API endpoints:")
	log.Printf("  GET    /api/v1/users           - Get all users")
	log.Printf("  GET    /api/v1/users/{id}      - Get user by ID")
	log.Printf("  POST   /api/v1/users           - Create new user")
	log.Printf("  PUT    /api/v1/users/{id}      - Update user")
	log.Printf("  DELETE /api/v1/users/{id}      - Delete user")
	log.Printf("  GET    /api/v1/cache/stats     - Get cache statistics")
	log.Printf("  POST   /api/v1/cache/clear     - Clear cache")
	log.Printf("  GET    /health                 - Health check")
	log.Printf("")
	log.Printf("Try these commands:")
	log.Printf("  curl http://localhost:8080/health")
	log.Printf("  curl http://localhost:8080/api/v1/users")
	log.Printf("  curl http://localhost:8080/api/v1/users/1")
	log.Printf("  curl http://localhost:8080/api/v1/cache/stats")

	log.Fatal(http.ListenAndServe(port, nil))
}
