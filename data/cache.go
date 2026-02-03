// Package data provides data fetching and caching functionality.
package data

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"sector-analyzer/config"
)

// CacheEntry represents a cached item with expiration.
type CacheEntry struct {
	Data      interface{} `json:"data"`
	CachedAt  time.Time   `json:"cached_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// Cache provides thread-safe in-memory caching with TTL.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

// NewCache creates a new cache instance.
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
	}
}

// GenerateKey creates a unique cache key from source and params.
func GenerateKey(source string, params map[string]interface{}) string {
	data, _ := json.Marshal(params)
	keyString := source + ":" + string(data)
	hash := md5.Sum([]byte(keyString))
	return hex.EncodeToString(hash[:])
}

// Get retrieves data from cache if valid. Expired entries return false.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Auto-expire: if past expiry time, treat as cache miss
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

// Set stores data in cache with default TTL.
func (c *Cache) Set(key string, data interface{}) {
	c.SetWithTTL(key, data, config.CacheDuration)
}

// SetWithTTL stores data in cache with custom TTL.
func (c *Cache) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.entries[key] = CacheEntry{
		Data:      data,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}
}

// Clear removes all entries from cache.
func (c *Cache) Clear() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := len(c.entries)
	c.entries = make(map[string]CacheEntry)
	return count
}

// Info returns cache statistics.
func (c *Cache) Info() CacheInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var valid, expired int
	now := time.Now()

	for _, entry := range c.entries {
		if now.Before(entry.ExpiresAt) {
			valid++
		} else {
			expired++
		}
	}

	return CacheInfo{
		TotalEntries:   len(c.entries),
		ValidEntries:   valid,
		ExpiredEntries: expired,
	}
}

// CacheInfo contains cache statistics.
type CacheInfo struct {
	TotalEntries   int `json:"total_entries"`
	ValidEntries   int `json:"valid_entries"`
	ExpiredEntries int `json:"expired_entries"`
}

// GlobalCache is the shared cache instance.
var GlobalCache = NewCache()
