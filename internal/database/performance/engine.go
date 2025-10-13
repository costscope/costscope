package performance

// Unified (no build tags) performance engine. Former perf-only behaviors (memory threshold callbacks,
// cache monitoring, persistence, tag-aware eviction) are now runtime gated by config flags so we can
// retire the `perf` build tag without fragmenting the API surface.
// Backward compatibility: existing code that depended on the minimal build works because advanced
// features are disabled by default. Former perf tests can be re-run simply by enabling flags in config.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// PerformanceEngine integrates memory management, parallel processing, and caching (runtime‑gated extras)
type PerformanceEngine struct {
	cache         *EnhancedCache
	config        *PerformanceConfig
	isInitialized bool
	mu            sync.RWMutex
}

// PerformanceConfig holds configuration for the performance engine
type PerformanceConfig struct {
	// Deprecated (kept for backward compatibility – values inform soft limits only)
	Memory *MemoryConfig `json:"memory"`
	// Deprecated: parallel worker settings retained to size concurrency in ExecuteParallel
	Parallel *ParallelConfig `json:"parallel"`
	Cache    *CacheConfig    `json:"cache"`
	Enabled  bool            `json:"enabled"`
}

// MemoryConfig (deprecated) – retained so existing construction sites compile; only MaxMemoryMB & GCThresholdMB used.
type MemoryConfig struct {
	MaxMemoryMB     int64         `json:"max_memory_mb"`
	GCThresholdMB   int64         `json:"gc_threshold_mb"`
	MonitorEnabled  bool          `json:"-"` // ignored
	MonitorInterval time.Duration `json:"-"` // ignored
}

// ParallelConfig (deprecated) – concurrency sizing only.
type ParallelConfig struct {
	WorkerCount      int           `json:"worker_count"`
	QueueSize        int           `json:"-"` // ignored
	MemoryLimitMB    int64         `json:"-"`
	EnableMonitoring bool          `json:"-"`
	WorkerTimeout    time.Duration `json:"worker_timeout"`
}

// Job represents a unit of work to be processed (API preserved from previous implementation).
type Job struct {
	ID        string
	Data      interface{}
	Processor func(interface{}) (interface{}, error)
	Priority  JobPriority // retained for compatibility; unused in simplified scheduler
	Timeout   time.Duration
}

// JobPriority retained for API compatibility.
type JobPriority int

const (
	PriorityLow JobPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// JobResult represents the result of a processed job.
type JobResult struct {
	JobID    string        `json:"job_id"`
	Result   interface{}   `json:"result"`
	Error    error         `json:"error"`
	Duration time.Duration `json:"duration"`
	WorkerID int           `json:"worker_id"`
}

// MemoryStats (simplified) gathered directly from runtime.MemStats.
type MemoryStats struct {
	AllocMB      int64     `json:"alloc_mb"`
	TotalAllocMB int64     `json:"total_alloc_mb"`
	SysMB        int64     `json:"sys_mb"`
	NumGC        uint32    `json:"num_gc"`
	LastGC       time.Time `json:"last_gc"`
	Timestamp    time.Time `json:"timestamp"`
}

// ProcessorStats (minimal) retained for metrics compatibility; values derived from execution snapshot.
type ProcessorStats struct {
	WorkerCount int       `json:"worker_count"`
	Timestamp   time.Time `json:"timestamp"`
}

// bytesToMBInt64 safely converts a byte count (uint64) to megabytes (int64) with saturation
// to avoid potential overflow flagged by gosec (G115). Extremely large values are capped at MaxInt64.
func bytesToMBInt64(b uint64) int64 { // divisor constants are well-known powers of two
	const mb = 1024 * 1024
	v := b / mb
	if v > uint64(math.MaxInt64) { // saturate defensively
		return math.MaxInt64
	}
	return int64(v)
}

// safeUnixNano converts a runtime-provided uint64 nanosecond timestamp to time.Unix argument safely.
// It clamps to math.MaxInt64 if the value ever exceeded signed range (theoretically unreachable here).
func safeUnixNano(ns uint64) int64 {
	if ns > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(ns)
}

// PerformanceMetrics holds performance metrics
type PerformanceMetrics struct {
	MemoryStats    *MemoryStats    `json:"memory_stats"`
	ProcessorStats *ProcessorStats `json:"processor_stats"`
	CacheStats     *CacheStats     `json:"cache_stats"`
	Timestamp      time.Time       `json:"timestamp"`
}

// DefaultPerformanceConfig returns default performance configuration with advanced (ex-perf) features OFF
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		Enabled: true,
		Memory: &MemoryConfig{
			MaxMemoryMB:   2048,
			GCThresholdMB: 1024,
		},
		Parallel: &ParallelConfig{
			WorkerCount:   4,
			WorkerTimeout: 30 * time.Second,
		},
		Cache: &CacheConfig{
			MaxSize:         10000,
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
	}
}

// NewPerformanceEngine creates a new performance engine
func NewPerformanceEngine(config *PerformanceConfig) *PerformanceEngine {
	if config == nil {
		config = DefaultPerformanceConfig()
	}
	// Sane fallbacks
	if config.Memory == nil {
		config.Memory = &MemoryConfig{MaxMemoryMB: 2048, GCThresholdMB: 1024}
	}
	if config.Parallel == nil {
		config.Parallel = &ParallelConfig{WorkerCount: 4, WorkerTimeout: 30 * time.Second}
	}
	pe := &PerformanceEngine{config: config}
	if config.Enabled {
		pe.cache = NewEnhancedCache(config.Cache)
		if config.Cache != nil && config.Cache.Monitor != nil && config.Cache.Monitor.Enabled {
			pe.cache.AddMonitorCallback(pe.handleCachePerformance)
		}
	}
	return pe
}

// Start initializes and starts the performance engine
func (pe *PerformanceEngine) Start(ctx context.Context) error { //nolint:revive // ctx kept for future hooks
	pe.mu.Lock()
	defer pe.mu.Unlock()
	if pe.isInitialized {
		return fmt.Errorf("performance engine already started")
	}
	if !pe.config.Enabled {
		pe.isInitialized = true
		return nil
	}
	pe.isInitialized = true
	return nil
}

// Stop shuts down the performance engine
func (pe *PerformanceEngine) Stop() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	if !pe.isInitialized {
		return nil
	}
	if pe.cache != nil {
		if err := pe.cache.Close(); err != nil {
			return fmt.Errorf("cache close error: %w", err)
		}
	}
	pe.isInitialized = false
	return nil
}

// OptimizeMemory performs memory optimization
func (pe *PerformanceEngine) OptimizeMemory(ctx context.Context) error { //nolint:revive // ctx reserved for future extensions
	if !pe.config.Enabled {
		return nil
	}
	// Force GC and check simple threshold if provided
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := bytesToMBInt64(m.Alloc)
	if pe.config.Memory != nil && pe.config.Memory.MaxMemoryMB > 0 && allocMB > pe.config.Memory.MaxMemoryMB {
		return fmt.Errorf("memory usage %d MB exceeds limit %d MB", allocMB, pe.config.Memory.MaxMemoryMB)
	}
	if pe.config.Memory != nil && pe.config.Memory.GCThresholdMB > 0 && allocMB > pe.config.Memory.GCThresholdMB {
		runtime.GC()
	}
	return nil
}

// ExecuteParallel executes jobs in parallel
func (pe *PerformanceEngine) ExecuteParallel(ctx context.Context, jobs []Job) ([]JobResult, error) {
	if !pe.config.Enabled {
		return nil, fmt.Errorf("parallel processing not enabled")
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	workerCount := 4
	if pe.config.Parallel != nil && pe.config.Parallel.WorkerCount > 0 {
		workerCount = pe.config.Parallel.WorkerCount
	}
	if workerCount > len(jobs) {
		workerCount = len(jobs)
	}
	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup
	results := make([]JobResult, len(jobs))
	// Track first context error
	var ctxErr error
	var ctxErrOnce sync.Once
	start := time.Now()
	for i, job := range jobs {
		i, job := i, job // capture
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if job.Processor == nil {
				results[i] = JobResult{JobID: job.ID, Error: errors.New("nil processor"), Duration: time.Since(start)}
				return
			}
			jStart := time.Now()
			// Per-job timeout
			jCtx := ctx
			var cancel context.CancelFunc
			if job.Timeout > 0 {
				jCtx, cancel = context.WithTimeout(ctx, job.Timeout)
			}
			if cancel != nil {
				defer cancel()
			}
			// Run processor in separate goroutine but avoid data race by communicating via channel
			type procResult struct {
				res interface{}
				err error
			}
			prCh := make(chan procResult, 1)
			go func() {
				// Always non-blocking send due to buffer size 1
				r, e := job.Processor(job.Data)
				prCh <- procResult{res: r, err: e}
			}()
			var finalRes interface{}
			var finalErr error
			select {
			case <-jCtx.Done():
				finalErr = fmt.Errorf("job timeout: %v", jCtx.Err())
			case pr := <-prCh:
				finalRes = pr.res
				finalErr = pr.err
			}
			if finalErr != nil && errors.Is(finalErr, context.Canceled) {
				ctxErrOnce.Do(func() { ctxErr = finalErr })
			}
			results[i] = JobResult{JobID: job.ID, Result: finalRes, Error: finalErr, Duration: time.Since(jStart), WorkerID: i % workerCount}
		}()
	}
	wg.Wait()
	if ctxErr != nil {
		return results, ctxErr
	}
	return results, nil
}

// CacheGet retrieves a value from cache
func (pe *PerformanceEngine) CacheGet(key string) (interface{}, bool) {
	if !pe.config.Enabled || pe.cache == nil {
		return nil, false
	}
	return pe.cache.Get(key)
}

func (pe *PerformanceEngine) CacheSet(key string, value interface{}) error {
	if !pe.config.Enabled || pe.cache == nil {
		return nil
	}
	return pe.cache.Set(key, value)
}
func (pe *PerformanceEngine) CacheDelete(key string) bool {
	if !pe.config.Enabled || pe.cache == nil {
		return false
	}
	return pe.cache.Delete(key)
}

// GetMetrics returns comprehensive performance metrics
func (pe *PerformanceEngine) GetMetrics() *PerformanceMetrics {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	metrics := &PerformanceMetrics{Timestamp: time.Now()}
	if pe.config.Enabled {
		// Memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		metrics.MemoryStats = &MemoryStats{
			AllocMB:      bytesToMBInt64(m.Alloc),
			TotalAllocMB: bytesToMBInt64(m.TotalAlloc),
			SysMB:        bytesToMBInt64(m.Sys),
			NumGC:        m.NumGC,
			LastGC:       time.Unix(0, safeUnixNano(m.LastGC)),
			Timestamp:    time.Now(),
		}
		metrics.ProcessorStats = &ProcessorStats{WorkerCount: pe.config.Parallel.WorkerCount, Timestamp: time.Now()}
		if pe.cache != nil {
			metrics.CacheStats = pe.cache.GetStats()
		}
	}
	return metrics
}

// --- Former perf-only callbacks now runtime gated ---
func (pe *PerformanceEngine) handleCachePerformance(stats *CacheStats) { // unchanged behavior
	if stats != nil && stats.HitRate < 0.5 {
		logging.GetLogger().WarnWithFields("cache hit rate below target", map[string]interface{}{"hit_rate_pct": stats.HitRate * 100})
	}
}

// IsEnabled returns whether performance engine is enabled
func (pe *PerformanceEngine) IsEnabled() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.config.Enabled
}

// IsInitialized returns whether performance engine is initialized
func (pe *PerformanceEngine) IsInitialized() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.isInitialized
}

// Minimal build omits memory/caching performance callback handlers (perf build implements them).

// NOTE: Benchmark-related types & methods moved under build tag 'perf' in performance_benchmarks_perf.go
