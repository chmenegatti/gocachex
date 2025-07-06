// Package sharding provides consistent hashing and data distribution for GoCacheX.
package sharding

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"sort"

	"github.com/chmenegatti/gocachex/pkg/backends"
	"github.com/chmenegatti/gocachex/pkg/config"
)

// Sharder provides data sharding functionality.
type Sharder interface {
	GetShard(key string) backends.Backend
	GetShardIndex(key string) int
	AddShard(backend backends.Backend) error
	RemoveShard(index int) error
	GetShards() []backends.Backend
	GetShardCount() int
}

// ConsistentHashSharder implements consistent hashing for data distribution.
type ConsistentHashSharder struct {
	shards   []backends.Backend
	replicas int
	ring     map[uint32]int
	keys     []uint32
}

// NewConsistentHashSharder creates a new consistent hash sharder.
func NewConsistentHashSharder(replicas int) *ConsistentHashSharder {
	return &ConsistentHashSharder{
		shards:   make([]backends.Backend, 0),
		replicas: replicas,
		ring:     make(map[uint32]int),
		keys:     make([]uint32, 0),
	}
}

// AddShard adds a new shard to the consistent hash ring.
func (c *ConsistentHashSharder) AddShard(backend backends.Backend) error {
	shardIndex := len(c.shards)
	c.shards = append(c.shards, backend)

	// Add virtual nodes for this shard
	for i := 0; i < c.replicas; i++ {
		key := c.hashKey(fmt.Sprintf("shard-%d-%d", shardIndex, i))
		c.ring[key] = shardIndex
		c.keys = append(c.keys, key)
	}

	// Sort keys to maintain order
	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})

	return nil
}

// RemoveShard removes a shard from the consistent hash ring.
func (c *ConsistentHashSharder) RemoveShard(index int) error {
	if index < 0 || index >= len(c.shards) {
		return fmt.Errorf("invalid shard index: %d", index)
	}

	// Remove virtual nodes for this shard
	newKeys := make([]uint32, 0)
	for _, key := range c.keys {
		if c.ring[key] != index {
			newKeys = append(newKeys, key)
		} else {
			delete(c.ring, key)
		}
	}
	c.keys = newKeys

	// Update ring to adjust indices after removal
	newRing := make(map[uint32]int)
	for key, shardIdx := range c.ring {
		if shardIdx > index {
			newRing[key] = shardIdx - 1
		} else {
			newRing[key] = shardIdx
		}
	}
	c.ring = newRing

	// Remove shard from slice
	c.shards = append(c.shards[:index], c.shards[index+1:]...)

	return nil
}

// GetShard returns the shard backend for a given key.
func (c *ConsistentHashSharder) GetShard(key string) backends.Backend {
	if len(c.shards) == 0 {
		return nil
	}

	if len(c.shards) == 1 {
		return c.shards[0]
	}

	hash := c.hashKey(key)

	// Find the first node >= hash
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	// If no node found, wrap around to the first node
	if idx == len(c.keys) {
		idx = 0
	}

	shardIndex := c.ring[c.keys[idx]]
	return c.shards[shardIndex]
}

// GetShardIndex returns the shard index for a given key.
func (c *ConsistentHashSharder) GetShardIndex(key string) int {
	if len(c.shards) == 0 {
		return -1
	}

	if len(c.shards) == 1 {
		return 0
	}

	hash := c.hashKey(key)

	// Find the first node >= hash
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	// If no node found, wrap around to the first node
	if idx == len(c.keys) {
		idx = 0
	}

	return c.ring[c.keys[idx]]
}

// GetShards returns all shard backends.
func (c *ConsistentHashSharder) GetShards() []backends.Backend {
	return c.shards
}

// GetShardCount returns the number of shards.
func (c *ConsistentHashSharder) GetShardCount() int {
	return len(c.shards)
}

// hashKey computes a hash for a given key.
func (c *ConsistentHashSharder) hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

// HashSharder implements simple hash-based sharding.
type HashSharder struct {
	shards []backends.Backend
}

// NewHashSharder creates a new hash-based sharder.
func NewHashSharder() *HashSharder {
	return &HashSharder{
		shards: make([]backends.Backend, 0),
	}
}

// AddShard adds a new shard.
func (h *HashSharder) AddShard(backend backends.Backend) error {
	h.shards = append(h.shards, backend)
	return nil
}

// RemoveShard removes a shard.
func (h *HashSharder) RemoveShard(index int) error {
	if index < 0 || index >= len(h.shards) {
		return fmt.Errorf("invalid shard index: %d", index)
	}

	h.shards = append(h.shards[:index], h.shards[index+1:]...)
	return nil
}

// GetShard returns the shard backend for a given key.
func (h *HashSharder) GetShard(key string) backends.Backend {
	if len(h.shards) == 0 {
		return nil
	}

	hash := h.hashKey(key)
	index := int(hash) % len(h.shards)
	return h.shards[index]
}

// GetShardIndex returns the shard index for a given key.
func (h *HashSharder) GetShardIndex(key string) int {
	if len(h.shards) == 0 {
		return -1
	}

	hash := h.hashKey(key)
	return int(hash) % len(h.shards)
}

// GetShards returns all shard backends.
func (h *HashSharder) GetShards() []backends.Backend {
	return h.shards
}

// GetShardCount returns the number of shards.
func (h *HashSharder) GetShardCount() int {
	return len(h.shards)
}

// hashKey computes a hash for a given key.
func (h *HashSharder) hashKey(key string) uint32 {
	hash := md5.Sum([]byte(key))
	return uint32(hash[0])<<24 | uint32(hash[1])<<16 | uint32(hash[2])<<8 | uint32(hash[3])
}

// RangeSharder implements range-based sharding.
type RangeSharder struct {
	shards []backends.Backend
	ranges []string // Range boundaries
}

// NewRangeSharder creates a new range-based sharder.
func NewRangeSharder() *RangeSharder {
	return &RangeSharder{
		shards: make([]backends.Backend, 0),
		ranges: make([]string, 0),
	}
}

// AddShard adds a new shard with a range boundary.
func (r *RangeSharder) AddShard(backend backends.Backend) error {
	r.shards = append(r.shards, backend)
	// For simplicity, use alphabetical ranges
	r.ranges = append(r.ranges, string(rune('a'+len(r.shards)-1)))
	return nil
}

// RemoveShard removes a shard.
func (r *RangeSharder) RemoveShard(index int) error {
	if index < 0 || index >= len(r.shards) {
		return fmt.Errorf("invalid shard index: %d", index)
	}

	r.shards = append(r.shards[:index], r.shards[index+1:]...)
	r.ranges = append(r.ranges[:index], r.ranges[index+1:]...)
	return nil
}

// GetShard returns the shard backend for a given key.
func (r *RangeSharder) GetShard(key string) backends.Backend {
	if len(r.shards) == 0 {
		return nil
	}

	index := r.GetShardIndex(key)
	if index < 0 {
		return r.shards[0]
	}

	return r.shards[index]
}

// GetShardIndex returns the shard index for a given key.
func (r *RangeSharder) GetShardIndex(key string) int {
	if len(r.shards) == 0 {
		return -1
	}

	// Simple alphabetical range sharding
	if len(key) == 0 {
		return 0
	}

	firstChar := key[0]
	for i, boundary := range r.ranges {
		if firstChar <= boundary[0] {
			return i
		}
	}

	// Default to last shard
	return len(r.shards) - 1
}

// GetShards returns all shard backends.
func (r *RangeSharder) GetShards() []backends.Backend {
	return r.shards
}

// GetShardCount returns the number of shards.
func (r *RangeSharder) GetShardCount() int {
	return len(r.shards)
}

// NewSharder creates a new sharder based on configuration.
func NewSharder(cfg config.ShardingConfig) Sharder {
	switch cfg.Algorithm {
	case "consistent":
		replicas := cfg.Replicas
		if replicas <= 0 {
			replicas = 100 // Default replicas
		}
		return NewConsistentHashSharder(replicas)
	case "hash":
		return NewHashSharder()
	case "range":
		return NewRangeSharder()
	default:
		return NewConsistentHashSharder(100)
	}
}

// ShardKey is a helper function to determine which shard a key belongs to.
func ShardKey(key string, shardCount int) int {
	if shardCount <= 0 {
		return 0
	}

	hash := crc32.ChecksumIEEE([]byte(key))
	return int(hash) % shardCount
}

// GenerateShardKey creates a shard-specific key.
func GenerateShardKey(originalKey string, shardIndex int) string {
	return fmt.Sprintf("shard:%d:%s", shardIndex, originalKey)
}

// ExtractOriginalKey extracts the original key from a shard-specific key.
func ExtractOriginalKey(shardKey string) string {
	// Parse "shard:N:original_key" format
	parts := []string{}
	current := ""
	for _, char := range shardKey {
		if char == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) >= 3 && parts[0] == "shard" {
		// Join all parts after the shard index
		result := ""
		for i := 2; i < len(parts); i++ {
			if i > 2 {
				result += ":"
			}
			result += parts[i]
		}
		return result
	}

	return shardKey // Return as-is if not in expected format
}
