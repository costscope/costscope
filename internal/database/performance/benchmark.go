package performance

// Unified benchmark harness exposed always (formerly behind perf build tag). This keeps the
// external perf-bench tool working without reintroducing build tag divergence. Lightweight
// and only used when explicitly invoked.

import (
	"context"
	"fmt"
	"time"
)

// PerformanceBenchmark runs performance benchmarks over the unified engine.
type PerformanceBenchmark struct {
	engine *PerformanceEngine
}

// NewPerformanceBenchmark creates a new performance benchmark harness.
func NewPerformanceBenchmark(engine *PerformanceEngine) *PerformanceBenchmark {
	return &PerformanceBenchmark{engine: engine}
}

// BenchmarkResult holds benchmark results.
type BenchmarkResult struct {
	TestName         string        `json:"test_name"`
	Duration         time.Duration `json:"duration"`
	OperationsCount  int64         `json:"operations_count"`
	OperationsPerSec float64       `json:"operations_per_sec"`
	MemoryUsedMB     int64         `json:"memory_used_mb"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
	Timestamp        time.Time     `json:"timestamp"`
}

// RunMemoryBenchmark benchmarks memory optimization + GC path.
func (pb *PerformanceBenchmark) RunMemoryBenchmark(ctx context.Context, operations int) *BenchmarkResult {
	start := time.Now()
	startStats := pb.engine.GetMetrics()
	for i := 0; i < operations; i++ {
		if i%100 == 0 {
			_ = pb.engine.OptimizeMemory(ctx)
		}
	}
	duration := time.Since(start)
	endStats := pb.engine.GetMetrics()
	memoryUsed := int64(0)
	if endStats.MemoryStats != nil && startStats.MemoryStats != nil {
		memoryUsed = endStats.MemoryStats.AllocMB - startStats.MemoryStats.AllocMB
	}
	return &BenchmarkResult{
		TestName:         "Memory Management",
		Duration:         duration,
		OperationsCount:  int64(operations),
		OperationsPerSec: float64(operations) / maxFloat(duration.Seconds(), 1e-9),
		MemoryUsedMB:     memoryUsed,
		Timestamp:        time.Now(),
	}
}

// RunCacheBenchmark benchmarks cache hit path.
func (pb *PerformanceBenchmark) RunCacheBenchmark(operations int) *BenchmarkResult {
	start := time.Now()
	hits := 0
	for i := 0; i < operations; i++ {
		key := fmt.Sprintf("bench_key_%d", i%100)
		if _, found := pb.engine.CacheGet(key); found {
			hits++
		} else {
			_ = pb.engine.CacheSet(key, fmt.Sprintf("value_%d", i))
		}
	}
	duration := time.Since(start)
	return &BenchmarkResult{
		TestName:         "Cache Performance",
		Duration:         duration,
		OperationsCount:  int64(operations),
		OperationsPerSec: float64(operations) / maxFloat(duration.Seconds(), 1e-9),
		CacheHitRate:     float64(hits) / float64(operations),
		Timestamp:        time.Now(),
	}
}

// RunParallelBenchmark benchmarks parallel processing.
func (pb *PerformanceBenchmark) RunParallelBenchmark(ctx context.Context, jobCount int) *BenchmarkResult {
	start := time.Now()
	jobs := make([]Job, jobCount)
	for i := 0; i < jobCount; i++ {
		jobs[i] = Job{ID: fmt.Sprintf("bench_job_%d", i), Data: i, Processor: func(data interface{}) (interface{}, error) {
			num := data.(int)
			total := 0
			for j := 0; j < 500; j++ { // simulate cpu work
				total += num * j
			}
			return total, nil
		}}
	}
	results, err := pb.engine.ExecuteParallel(ctx, jobs)
	duration := time.Since(start)
	success := int64(len(results))
	if err != nil {
		success = 0
		for _, r := range results {
			if r.Error == nil {
				success++
			}
		}
	}
	return &BenchmarkResult{
		TestName:         "Parallel Processing",
		Duration:         duration,
		OperationsCount:  success,
		OperationsPerSec: float64(success) / maxFloat(duration.Seconds(), 1e-9),
		Timestamp:        time.Now(),
	}
}

// maxFloat helper to avoid division by zero.
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
