package cache

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

// TokenCache defines the interface for token caching
type TokenCache interface {
	Set(key string, token *Token)
	Get(key string) (*Token, bool)
	Delete(key string)
	Clear()
	Size() int
}

// Cache defines a generic cache interface
type Cache interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string)
	Clear()
	Size() int
}

// InMemoryTokenCache implements TokenCache using an in-memory cache
type InMemoryTokenCache struct {
	cache *cache.Cache
	mutex sync.RWMutex
}

// NewInMemoryTokenCache creates a new in-memory token cache
func NewInMemoryTokenCache(defaultExpiration, cleanupInterval time.Duration) *InMemoryTokenCache {
	return &InMemoryTokenCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set stores a token in the cache
func (c *InMemoryTokenCache) Set(key string, token *Token) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if token == nil {
		log.Debug().Str("key", key).Msg("Attempted to cache nil token")
		return
	}

	// Calculate expiration based on token expiry
	var expiration time.Duration = cache.DefaultExpiration
	if !token.ExpiresAt.IsZero() {
		ttl := time.Until(token.ExpiresAt)
		if ttl > 0 {
			// Cache for 90% of the token's remaining lifetime to account for clock skew
			expiration = time.Duration(float64(ttl) * 0.9)
		}
	}

	c.cache.Set(key, token, expiration)
	
	log.Debug().
		Str("key", key).
		Time("expires_at", token.ExpiresAt).
		Dur("cache_ttl", expiration).
		Msg("Token cached successfully")
}

// Get retrieves a token from the cache
func (c *InMemoryTokenCache) Get(key string) (*Token, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	value, found := c.cache.Get(key)
	if !found {
		log.Debug().Str("key", key).Msg("Token not found in cache")
		return nil, false
	}

	token, ok := value.(*Token)
	if !ok {
		log.Error().Str("key", key).Msg("Invalid token type in cache")
		c.cache.Delete(key) // Clean up invalid entry
		return nil, false
	}

	// Double-check if token is still valid
	if token.IsExpired() {
		log.Debug().Str("key", key).Msg("Cached token has expired")
		c.cache.Delete(key) // Clean up expired token
		return nil, false
	}

	log.Debug().Str("key", key).Msg("Token retrieved from cache")
	return token, true
}

// Delete removes a token from the cache
func (c *InMemoryTokenCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache.Delete(key)
	log.Debug().Str("key", key).Msg("Token deleted from cache")
}

// Clear removes all tokens from the cache
func (c *InMemoryTokenCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache.Flush()
	log.Debug().Msg("Token cache cleared")
}

// Size returns the number of items in the cache
func (c *InMemoryTokenCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return c.cache.ItemCount()
}

// GetExpiredTokens returns a list of expired token keys
func (c *InMemoryTokenCache) GetExpiredTokens() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	var expiredKeys []string
	for key, item := range c.cache.Items() {
		if token, ok := item.Object.(*Token); ok {
			if token.IsExpired() {
				expiredKeys = append(expiredKeys, key)
			}
		}
	}
	
	return expiredKeys
}

// CleanupExpiredTokens manually removes expired tokens
func (c *InMemoryTokenCache) CleanupExpiredTokens() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	expiredKeys := []string{}
	for key, item := range c.cache.Items() {
		if token, ok := item.Object.(*Token); ok {
			if token.IsExpired() {
				expiredKeys = append(expiredKeys, key)
			}
		}
	}

	for _, key := range expiredKeys {
		c.cache.Delete(key)
	}

	if len(expiredKeys) > 0 {
		log.Debug().Int("count", len(expiredKeys)).Msg("Cleaned up expired tokens")
	}

	return len(expiredKeys)
}

// InMemoryCache implements Cache using an in-memory cache
type InMemoryCache struct {
	cache *cache.Cache
	mutex sync.RWMutex
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(defaultExpiration, cleanupInterval time.Duration) *InMemoryCache {
	return &InMemoryCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache.Set(key, value, cache.DefaultExpiration)
	log.Debug().Str("key", key).Msg("Value cached successfully")
}

// SetWithTTL stores a value in the cache with a specific TTL
func (c *InMemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache.Set(key, value, ttl)
	log.Debug().Str("key", key).Dur("ttl", ttl).Msg("Value cached with TTL")
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	value, found := c.cache.Get(key)
	if found {
		log.Debug().Str("key", key).Msg("Value retrieved from cache")
	} else {
		log.Debug().Str("key", key).Msg("Value not found in cache")
	}
	return value, found
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache.Delete(key)
	log.Debug().Str("key", key).Msg("Value deleted from cache")
}

// Clear removes all values from the cache
func (c *InMemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache.Flush()
	log.Debug().Msg("Cache cleared")
}

// Size returns the number of items in the cache
func (c *InMemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return c.cache.ItemCount()
}

// Keys returns all keys in the cache
func (c *InMemoryCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	keys := make([]string, 0, c.cache.ItemCount())
	for key := range c.cache.Items() {
		keys = append(keys, key)
	}
	
	return keys
}

// GetWithExpiration retrieves a value with its expiration time
func (c *InMemoryCache) GetWithExpiration(key string) (interface{}, time.Time, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return c.cache.GetWithExpiration(key)
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size      int   `json:"size"`
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	HitRatio  float64 `json:"hit_ratio"`
}

// StatsCache wraps a cache to collect statistics
type StatsCache struct {
	cache  Cache
	hits   int64
	misses int64
	mutex  sync.RWMutex
}

// NewStatsCache creates a cache that tracks statistics
func NewStatsCache(cache Cache) *StatsCache {
	return &StatsCache{
		cache: cache,
	}
}

// Set stores a value in the cache
func (sc *StatsCache) Set(key string, value interface{}) {
	sc.cache.Set(key, value)
}

// Get retrieves a value from the cache and tracks stats
func (sc *StatsCache) Get(key string) (interface{}, bool) {
	value, found := sc.cache.Get(key)
	
	sc.mutex.Lock()
	if found {
		sc.hits++
	} else {
		sc.misses++
	}
	sc.mutex.Unlock()
	
	return value, found
}

// Delete removes a value from the cache
func (sc *StatsCache) Delete(key string) {
	sc.cache.Delete(key)
}

// Clear removes all values from the cache and resets stats
func (sc *StatsCache) Clear() {
	sc.cache.Clear()
	sc.mutex.Lock()
	sc.hits = 0
	sc.misses = 0
	sc.mutex.Unlock()
}

// Size returns the number of items in the cache
func (sc *StatsCache) Size() int {
	return sc.cache.Size()
}

// GetStats returns cache statistics
func (sc *StatsCache) GetStats() CacheStats {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	total := sc.hits + sc.misses
	var hitRatio float64
	if total > 0 {
		hitRatio = float64(sc.hits) / float64(total)
	}
	
	return CacheStats{
		Size:     sc.cache.Size(),
		Hits:     sc.hits,
		Misses:   sc.misses,
		HitRatio: hitRatio,
	}
}

// ResetStats resets the cache statistics
func (sc *StatsCache) ResetStats() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	
	sc.hits = 0
	sc.misses = 0
}