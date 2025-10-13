//go:build duckdb

package duckdb

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/database"
)

// QueryCache provides intelligent caching for DuckDB query results
type QueryCache struct {
	cache     map[string]*CacheEntry
	mutex     sync.RWMutex
	maxSize   int
	hitCount  int64
	missCount int64
	evictions int64
}

// CacheEntry represents a cached query result with metadata
type CacheEntry struct {
	Result      *database.QueryResult `json:"result"`
	Timestamp   time.Time             `json:"timestamp"`
	TTL         time.Duration         `json:"ttl"`
	AccessCount int64                 `json:"access_count"`
	Size        int64                 `json:"size_bytes"`
}

// NewQueryCache creates a new query cache
func NewQueryCache(maxSize int) *QueryCache {
	return &QueryCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves a cached query result
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()

	entry, exists := qc.cache[key]
	if !exists {
		qc.missCount++
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Remove expired entry (will be cleaned up later)
		delete(qc.cache, key)
		qc.missCount++
		return nil, false
	}

	// Update access statistics
	entry.AccessCount++
	qc.hitCount++

	return entry.Result, true
}

// Set stores a query result in the cache
func (qc *QueryCache) Set(key string, result *database.QueryResult, ttl time.Duration) {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	// Calculate result size (approximate)
	size := qc.calculateResultSize(result)

	entry := &CacheEntry{
		Result:      result,
		Timestamp:   time.Now(),
		TTL:         ttl,
		AccessCount: 0,
		Size:        size,
	}

	// Check if we need to evict entries
	if len(qc.cache) >= qc.maxSize {
		qc.evictLRU()
	}

	qc.cache[key] = entry
}

// generateQueryHash generates a hash for query and parameters
func (qc *QueryCache) generateQueryHash(query string, params map[string]interface{}) string {
	h := sha256.New()
	h.Write([]byte(query))

	if params != nil {
		paramBytes, _ := json.Marshal(params)
		h.Write(paramBytes)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// calculateResultSize estimates the memory size of a query result
func (qc *QueryCache) calculateResultSize(result *database.QueryResult) int64 {
	if result == nil {
		return 0
	}

	size := int64(0)

	// Column names
	for _, col := range result.Columns {
		size += int64(len(col))
	}

	// Data rows (rough estimation)
	size += int64(len(result.Data)) * 100 // Estimate 100 bytes per row

	return size
}

// evictLRU evicts the least recently used entry
func (qc *QueryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	var lowestAccess int64 = -1

	// Find LRU entry (least accessed and oldest)
	for key, entry := range qc.cache {
		if lowestAccess == -1 || entry.AccessCount < lowestAccess {
			lowestAccess = entry.AccessCount
			oldestKey = key
			oldestTime = entry.Timestamp
		} else if entry.AccessCount == lowestAccess && entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(qc.cache, oldestKey)
		qc.evictions++
	}
}

// Clear clears all cached entries
func (qc *QueryCache) Clear() {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	qc.cache = make(map[string]*CacheEntry)
	qc.hitCount = 0
	qc.missCount = 0
	qc.evictions = 0
}

// GetHitRate returns the cache hit rate
func (qc *QueryCache) GetHitRate() float64 {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()

	total := qc.hitCount + qc.missCount
	if total == 0 {
		return 0.0
	}

	return float64(qc.hitCount) / float64(total)
}

// GetStats returns cache statistics
func (qc *QueryCache) GetStats() *database.CacheStats {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()

	totalSize := int64(0)
	entryCount := len(qc.cache)

	for _, entry := range qc.cache {
		totalSize += entry.Size
	}

	return &database.CacheStats{
		HitCount:   qc.hitCount,
		MissCount:  qc.missCount,
		HitRate:    qc.GetHitRate(),
		Size:       totalSize,
		EntryCount: entryCount,
		MaxSize:    int64(qc.maxSize),
		Evictions:  qc.evictions,
	}
}

// CleanupExpired removes expired entries from the cache
func (qc *QueryCache) CleanupExpired() {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	now := time.Now()
	for key, entry := range qc.cache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(qc.cache, key)
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up expired entries
func (qc *QueryCache) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			qc.CleanupExpired()
		}
	}()
}
