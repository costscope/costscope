package performance

import (
	"encoding/json"
	"sync"
	"time"
)

// EnhancedCache provides advanced caching capabilities with persistence and monitoring
type EnhancedCache struct {
	cache       map[string]*CacheEntry
	mu          sync.RWMutex
	maxSize     int
	defaultTTL  time.Duration
	stats       *CacheStats
	persistence CachePersistence
	monitor     *CacheMonitor
	stopCh      chan struct{}
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
	LastAccess  time.Time   `json:"last_access"`
	Size        int64       `json:"size"`
	Tags        []string    `json:"tags"`
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Evictions   int64     `json:"evictions"`
	Size        int       `json:"size"`
	TotalSize   int64     `json:"total_size_bytes"`
	HitRate     float64   `json:"hit_rate"`
	LastCleanup time.Time `json:"last_cleanup"`
	mu          sync.RWMutex
}

// CachePersistence defines interface for cache persistence
type CachePersistence interface {
	Save(entries map[string]*CacheEntry) error
	Load() (map[string]*CacheEntry, error)
	Clear() error
}

// CacheMonitor monitors cache performance and triggers optimizations
type CacheMonitor struct {
	enabled       bool
	interval      time.Duration
	hitRateTarget float64
	callbacks     []func(*CacheStats)
	mu            sync.RWMutex
}

// CacheConfig holds configuration for enhanced cache
type CacheConfig struct {
	MaxSize         int           `json:"max_size"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	Persistence     CachePersistence
	Monitor         *CacheMonitorConfig `json:"monitor"`
}

// CacheMonitorConfig holds monitor configuration
type CacheMonitorConfig struct {
	Enabled       bool          `json:"enabled"`
	Interval      time.Duration `json:"interval"`
	HitRateTarget float64       `json:"hit_rate_target"`
}

// NewEnhancedCache creates a new enhanced cache
func NewEnhancedCache(config *CacheConfig) *EnhancedCache {
	if config == nil {
		config = &CacheConfig{
			MaxSize:         10000,
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 10 * time.Minute,
			Monitor: &CacheMonitorConfig{
				Enabled:       true,
				Interval:      1 * time.Minute,
				HitRateTarget: 0.8,
			},
		}
	}

	// Apply sane defaults when provided values are zero or negative
	if config.MaxSize <= 0 {
		config.MaxSize = 10000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 1 * time.Hour
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 10 * time.Minute
	}
	if config.Monitor != nil && config.Monitor.Enabled {
		if config.Monitor.Interval <= 0 {
			config.Monitor.Interval = 1 * time.Minute
		}
		if config.Monitor.HitRateTarget <= 0 {
			config.Monitor.HitRateTarget = 0.8
		}
	}

	cache := &EnhancedCache{
		cache:       make(map[string]*CacheEntry),
		maxSize:     config.MaxSize,
		defaultTTL:  config.DefaultTTL,
		stats:       &CacheStats{},
		persistence: config.Persistence,
		stopCh:      make(chan struct{}),
	}

	// Initialize monitor
	if config.Monitor != nil && config.Monitor.Enabled {
		cache.monitor = &CacheMonitor{
			enabled:       true,
			interval:      config.Monitor.Interval,
			hitRateTarget: config.Monitor.HitRateTarget,
		}
		go cache.startMonitoring()
	}

	// Start cleanup routine
	go cache.startCleanup(config.CleanupInterval)

	// Load persisted entries if available
	if cache.persistence != nil {
		cache.loadFromPersistence()
	}

	return cache
}

// Get retrieves a value from cache
func (ec *EnhancedCache) Get(key string) (interface{}, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	entry, exists := ec.cache[key]
	if !exists {
		ec.updateStats(false, 0)
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		ec.mu.RUnlock()
		ec.mu.Lock()
		delete(ec.cache, key)
		ec.mu.Unlock()
		ec.mu.RLock()
		ec.updateStats(false, 0)
		return nil, false
	}

	// Update access metadata
	entry.AccessCount++
	entry.LastAccess = time.Now()
	ec.updateStats(true, entry.Size)

	return entry.Value, true
}

// Set stores a value in cache with default TTL
func (ec *EnhancedCache) Set(key string, value interface{}) error {
	return ec.SetWithTTL(key, value, ec.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (ec *EnhancedCache) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	return ec.SetWithOptions(key, value, &CacheOptions{
		TTL: ttl,
	})
}

// SetWithTags stores a value with tags for bulk operations
func (ec *EnhancedCache) SetWithTags(key string, value interface{}, tags []string) error {
	return ec.SetWithOptions(key, value, &CacheOptions{
		TTL:  ec.defaultTTL,
		Tags: tags,
	})
}

// CacheOptions holds options for cache operations
type CacheOptions struct {
	TTL  time.Duration
	Tags []string
}

// SetWithOptions stores a value with custom options
func (ec *EnhancedCache) SetWithOptions(key string, value interface{}, options *CacheOptions) error {
	if options == nil {
		options = &CacheOptions{TTL: ec.defaultTTL}
	}

	// Calculate size estimate
	size := ec.estimateSize(value)

	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		ExpiresAt:   time.Now().Add(options.TTL),
		CreatedAt:   time.Now(),
		AccessCount: 1,
		LastAccess:  time.Now(),
		Size:        size,
		Tags:        options.Tags,
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Check if we need to evict
	if len(ec.cache) >= ec.maxSize {
		ec.evictLRU()
	}

	ec.cache[key] = entry
	return nil
}

// Delete removes a key from cache
func (ec *EnhancedCache) Delete(key string) bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	_, exists := ec.cache[key]
	if exists {
		delete(ec.cache, key)
	}
	return exists
}

// DeleteByTags removes all entries with specified tags
func (ec *EnhancedCache) DeleteByTags(tags []string) int {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	deleted := 0
	for key, entry := range ec.cache {
		if ec.hasAnyTag(entry.Tags, tags) {
			delete(ec.cache, key)
			deleted++
		}
	}
	return deleted
}

// Clear removes all entries from cache
func (ec *EnhancedCache) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.cache = make(map[string]*CacheEntry)
}

// GetStats returns cache statistics
func (ec *EnhancedCache) GetStats() *CacheStats {
	ec.stats.mu.RLock()
	defer ec.stats.mu.RUnlock()

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	// Calculate hit rate
	total := ec.stats.Hits + ec.stats.Misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(ec.stats.Hits) / float64(total)
	}

	// Calculate total size
	var totalSize int64
	for _, entry := range ec.cache {
		totalSize += entry.Size
	}

	return &CacheStats{
		Hits:        ec.stats.Hits,
		Misses:      ec.stats.Misses,
		Evictions:   ec.stats.Evictions,
		Size:        len(ec.cache),
		TotalSize:   totalSize,
		HitRate:     hitRate,
		LastCleanup: ec.stats.LastCleanup,
	}
}

// Keys returns all cache keys
func (ec *EnhancedCache) Keys() []string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	keys := make([]string, 0, len(ec.cache))
	for key := range ec.cache {
		keys = append(keys, key)
	}
	return keys
}

// Close shuts down the cache
func (ec *EnhancedCache) Close() error {
	close(ec.stopCh)

	// Save to persistence if available
	if ec.persistence != nil {
		ec.mu.RLock()
		entries := make(map[string]*CacheEntry)
		for k, v := range ec.cache {
			entries[k] = v
		}
		ec.mu.RUnlock()

		return ec.persistence.Save(entries)
	}

	return nil
}

// Helper methods

func (ec *EnhancedCache) updateStats(hit bool, size int64) {
	ec.stats.mu.Lock()
	defer ec.stats.mu.Unlock()

	if hit {
		ec.stats.Hits++
	} else {
		ec.stats.Misses++
	}

	// Update total size tracking for cache management
	if size > 0 {
		ec.stats.TotalSize += size
	}
}

func (ec *EnhancedCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range ec.cache {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		delete(ec.cache, oldestKey)
		ec.stats.mu.Lock()
		ec.stats.Evictions++
		ec.stats.mu.Unlock()
	}
}

func (ec *EnhancedCache) estimateSize(value interface{}) int64 {
	// Simple size estimation
	data, err := json.Marshal(value)
	if err != nil {
		return 1024 // Default estimate
	}
	return int64(len(data))
}

func (ec *EnhancedCache) hasAnyTag(entryTags, searchTags []string) bool {
	for _, searchTag := range searchTags {
		for _, entryTag := range entryTags {
			if entryTag == searchTag {
				return true
			}
		}
	}
	return false
}

func (ec *EnhancedCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ec.cleanup()
		case <-ec.stopCh:
			return
		}
	}
}

func (ec *EnhancedCache) cleanup() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	now := time.Now()
	for key, entry := range ec.cache {
		if now.After(entry.ExpiresAt) {
			delete(ec.cache, key)
		}
	}

	ec.stats.mu.Lock()
	ec.stats.LastCleanup = now
	ec.stats.mu.Unlock()
}

func (ec *EnhancedCache) startMonitoring() {
	if ec.monitor == nil {
		return
	}

	ticker := time.NewTicker(ec.monitor.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ec.monitorPerformance()
		case <-ec.stopCh:
			return
		}
	}
}

func (ec *EnhancedCache) monitorPerformance() {
	stats := ec.GetStats()

	ec.monitor.mu.RLock()
	callbacks := ec.monitor.callbacks
	target := ec.monitor.hitRateTarget
	ec.monitor.mu.RUnlock()

	// Trigger callbacks if hit rate is below target
	if stats.HitRate < target {
		for _, callback := range callbacks {
			go callback(stats)
		}
	}
}

func (ec *EnhancedCache) loadFromPersistence() {
	entries, err := ec.persistence.Load()
	if err != nil {
		return // Silently fail, start with empty cache
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	now := time.Now()
	for key, entry := range entries {
		// Only load non-expired entries
		if now.Before(entry.ExpiresAt) {
			ec.cache[key] = entry
		}
	}
}

// AddMonitorCallback adds a performance monitoring callback
func (ec *EnhancedCache) AddMonitorCallback(callback func(*CacheStats)) {
	if ec.monitor == nil {
		return
	}

	ec.monitor.mu.Lock()
	defer ec.monitor.mu.Unlock()
	ec.monitor.callbacks = append(ec.monitor.callbacks, callback)
}

// NOTE: Deprecated GenerateKey helper removed; construct cache keys explicitly at call sites.
