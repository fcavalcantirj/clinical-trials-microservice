package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache provides caching functionality for trial data
type Cache struct {
	memCache *gocache.Cache
}

// NewCache creates a new cache instance with default TTL
func NewCache(defaultTTL time.Duration) *Cache {
	if defaultTTL == 0 {
		defaultTTL = 6 * time.Hour // Default 6 hour cache
	}
	cleanupInterval := defaultTTL / 2
	if cleanupInterval < time.Minute {
		cleanupInterval = time.Minute
	}
	return &Cache{
		memCache: gocache.New(defaultTTL, cleanupInterval),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.memCache.Get(key)
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.memCache.Set(key, value, gocache.DefaultExpiration)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.memCache.Set(key, value, ttl)
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.memCache.Delete(key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.memCache.Flush()
}

// GenerateCacheKey generates a cache key from search parameters
func GenerateCacheKey(base string, params map[string]interface{}) string {
	// Simple key generation - could be improved with hashing
	key := base
	for k, v := range params {
		key += ":" + k + "=" + toString(v)
	}
	return key
}

// toString converts a value to string for cache key generation
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []string:
		result := ""
		for i, s := range val {
			if i > 0 {
				result += ","
			}
			result += s
		}
		return result
	case int:
		return string(rune(val))
	case float64:
		return string(rune(val))
	default:
		return ""
	}
}
