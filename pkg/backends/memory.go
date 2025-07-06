package backends

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chmenegatti/gocachex/pkg/config"
)

// MemoryBackend implements an in-memory cache backend.
type MemoryBackend struct {
	mu          sync.RWMutex
	data        map[string]*memoryItem
	stats       *memoryStats
	config      config.MemoryConfig
	stopCleanup chan bool
	maxSize     int64
	currentSize int64
}

type memoryItem struct {
	value       []byte
	expireTime  time.Time
	accessTime  time.Time
	accessCount int64
}

type memoryStats struct {
	hits      int64
	misses    int64
	sets      int64
	deletes   int64
	evictions int64
	startTime time.Time
}

// NewMemoryBackend creates a new in-memory backend.
func NewMemoryBackend(cfg config.MemoryConfig) (*MemoryBackend, error) {
	maxSize, err := parseSize(cfg.MaxSize)
	if err != nil {
		return nil, fmt.Errorf("invalid max size: %w", err)
	}

	backend := &MemoryBackend{
		data:        make(map[string]*memoryItem),
		config:      cfg,
		maxSize:     maxSize,
		stopCleanup: make(chan bool),
		stats: &memoryStats{
			startTime: time.Now(),
		},
	}

	// Start cleanup goroutine
	go backend.cleanup()

	return backend, nil
}

// Get retrieves a value from the cache.
func (m *MemoryBackend) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&m.stats.misses, 1)
		return nil, fmt.Errorf("key not found")
	}

	// Check expiration
	if !item.expireTime.IsZero() && time.Now().After(item.expireTime) {
		m.mu.Lock()
		delete(m.data, key)
		m.mu.Unlock()
		atomic.AddInt64(&m.stats.misses, 1)
		return nil, fmt.Errorf("key expired")
	}

	// Update access statistics
	item.accessTime = time.Now()
	atomic.AddInt64(&item.accessCount, 1)
	atomic.AddInt64(&m.stats.hits, 1)

	return item.value, nil
}

// Set stores a value in the cache.
func (m *MemoryBackend) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var expireTime time.Time
	if ttl > 0 {
		expireTime = time.Now().Add(ttl)
	} else if m.config.DefaultTTL > 0 {
		expireTime = time.Now().Add(m.config.DefaultTTL)
	}

	item := &memoryItem{
		value:      value,
		expireTime: expireTime,
		accessTime: time.Now(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we need to evict items
	newSize := m.currentSize + int64(len(value))
	if m.maxSize > 0 && newSize > m.maxSize {
		m.evictItems(newSize - m.maxSize)
	}

	// Check max keys limit
	if m.config.MaxKeys > 0 && int64(len(m.data)) >= m.config.MaxKeys {
		m.evictLRU()
	}

	// Update current size
	if oldItem, exists := m.data[key]; exists {
		m.currentSize -= int64(len(oldItem.value))
	}
	m.currentSize += int64(len(value))

	m.data[key] = item
	atomic.AddInt64(&m.stats.sets, 1)

	return nil
}

// Delete removes a value from the cache.
func (m *MemoryBackend) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item, exists := m.data[key]; exists {
		m.currentSize -= int64(len(item.value))
		delete(m.data, key)
		atomic.AddInt64(&m.stats.deletes, 1)
	}

	return nil
}

// Exists checks if a key exists in the cache.
func (m *MemoryBackend) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check expiration
	if !item.expireTime.IsZero() && time.Now().After(item.expireTime) {
		m.mu.Lock()
		delete(m.data, key)
		m.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// GetMulti retrieves multiple values from the cache.
func (m *MemoryBackend) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	for _, key := range keys {
		if value, err := m.Get(ctx, key); err == nil {
			result[key] = value
		}
	}

	return result, nil
}

// SetMulti stores multiple values in the cache.
func (m *MemoryBackend) SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	for key, value := range items {
		if err := m.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMulti removes multiple values from the cache.
func (m *MemoryBackend) DeleteMulti(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Increment atomically increments a numeric value.
func (m *MemoryBackend) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists {
		// Create new item with delta value
		value := fmt.Sprintf("%d", delta)
		m.data[key] = &memoryItem{
			value:      []byte(value),
			accessTime: time.Now(),
		}
		return delta, nil
	}

	// Parse current value
	var current int64
	if err := json.Unmarshal(item.value, &current); err != nil {
		return 0, fmt.Errorf("value is not a number")
	}

	// Increment and store
	newValue := current + delta
	value := fmt.Sprintf("%d", newValue)
	item.value = []byte(value)
	item.accessTime = time.Now()

	return newValue, nil
}

// Decrement atomically decrements a numeric value.
func (m *MemoryBackend) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return m.Increment(ctx, key, -delta)
}

// Expire sets a timeout on a key.
func (m *MemoryBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	if ttl > 0 {
		item.expireTime = time.Now().Add(ttl)
	} else {
		item.expireTime = time.Time{}
	}

	return nil
}

// TTL returns the remaining time to live of a key.
func (m *MemoryBackend) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	if item.expireTime.IsZero() {
		return -1, nil // No expiration
	}

	remaining := time.Until(item.expireTime)
	if remaining < 0 {
		return 0, nil // Expired
	}

	return remaining, nil
}

// Clear removes all keys from the cache.
func (m *MemoryBackend) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]*memoryItem)
	m.currentSize = 0

	return nil
}

// Stats returns cache statistics.
func (m *MemoryBackend) Stats(ctx context.Context) (*Stats, error) {
	m.mu.RLock()
	keyCount := int64(len(m.data))
	memoryUsage := m.currentSize
	m.mu.RUnlock()

	return &Stats{
		Hits:        atomic.LoadInt64(&m.stats.hits),
		Misses:      atomic.LoadInt64(&m.stats.misses),
		Sets:        atomic.LoadInt64(&m.stats.sets),
		Deletes:     atomic.LoadInt64(&m.stats.deletes),
		Evictions:   atomic.LoadInt64(&m.stats.evictions),
		KeyCount:    keyCount,
		MemoryUsage: memoryUsage,
		Uptime:      int64(time.Since(m.stats.startTime).Seconds()),
	}, nil
}

// Health checks the health of the backend.
func (m *MemoryBackend) Health(ctx context.Context) error {
	// Memory backend is always healthy if it's running
	return nil
}

// Close closes the backend and releases resources.
func (m *MemoryBackend) Close() error {
	close(m.stopCleanup)
	return nil
}

// cleanup runs periodic cleanup of expired items.
func (m *MemoryBackend) cleanup() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpired()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes expired items from the cache.
func (m *MemoryBackend) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, item := range m.data {
		if !item.expireTime.IsZero() && now.After(item.expireTime) {
			m.currentSize -= int64(len(item.value))
			delete(m.data, key)
		}
	}
}

// evictItems evicts items to free up the specified amount of memory.
func (m *MemoryBackend) evictItems(sizeToFree int64) {
	switch m.config.EvictionPolicy {
	case "lru":
		m.evictLRU()
	case "lfu":
		m.evictLFU()
	case "random":
		m.evictRandom()
	default:
		m.evictLRU()
	}
}

// evictLRU evicts the least recently used item.
func (m *MemoryBackend) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range m.data {
		if oldestKey == "" || item.accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.accessTime
		}
	}

	if oldestKey != "" {
		item := m.data[oldestKey]
		m.currentSize -= int64(len(item.value))
		delete(m.data, oldestKey)
		atomic.AddInt64(&m.stats.evictions, 1)
	}
}

// evictLFU evicts the least frequently used item.
func (m *MemoryBackend) evictLFU() {
	var targetKey string
	var minAccess int64 = -1

	for key, item := range m.data {
		if minAccess == -1 || item.accessCount < minAccess {
			targetKey = key
			minAccess = item.accessCount
		}
	}

	if targetKey != "" {
		item := m.data[targetKey]
		m.currentSize -= int64(len(item.value))
		delete(m.data, targetKey)
		atomic.AddInt64(&m.stats.evictions, 1)
	}
}

// evictRandom evicts a random item.
func (m *MemoryBackend) evictRandom() {
	for key, item := range m.data {
		m.currentSize -= int64(len(item.value))
		delete(m.data, key)
		atomic.AddInt64(&m.stats.evictions, 1)
		break // Remove only one item
	}
}

// parseSize parses a size string like "100MB" into bytes.
func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	// Simple implementation - can be enhanced
	switch {
	case len(sizeStr) > 2 && sizeStr[len(sizeStr)-2:] == "KB":
		size := parseInt(sizeStr[:len(sizeStr)-2])
		return int64(size) * 1024, nil
	case len(sizeStr) > 2 && sizeStr[len(sizeStr)-2:] == "MB":
		size := parseInt(sizeStr[:len(sizeStr)-2])
		return int64(size) * 1024 * 1024, nil
	case len(sizeStr) > 2 && sizeStr[len(sizeStr)-2:] == "GB":
		size := parseInt(sizeStr[:len(sizeStr)-2])
		return int64(size) * 1024 * 1024 * 1024, nil
	default:
		return int64(parseInt(sizeStr)), nil
	}
}

func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			break
		}
	}
	return result
}
