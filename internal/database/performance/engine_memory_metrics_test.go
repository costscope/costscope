package performance

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// global sink to retain allocations and prevent compiler optimization.
var memorySink [][]byte

// TestOptimizeMemory_ExceedsLimit triggers the MaxMemoryMB limit by allocating a large slice.
func TestOptimizeMemory_ExceedsLimit(t *testing.T) {
	cfg := &PerformanceConfig{Enabled: true, Memory: &MemoryConfig{MaxMemoryMB: 1024, GCThresholdMB: 512}, Parallel: &ParallelConfig{WorkerCount: 2, WorkerTimeout: time.Second}, Cache: &CacheConfig{MaxSize: 10}}
	eng := NewPerformanceEngine(cfg)
	if err := eng.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Allocate several large slices to grow live heap usage.
	for i := 0; i < 4; i++ { // ~32MB
		b := make([]byte, 8<<20)
		for j := 0; j < len(b); j += 4096 { // touch pages to ensure physical allocation
			b[j] = byte(i)
		}
		memorySink = append(memorySink, b)
	}
	// Measure current allocation and set an artificial MaxMemoryMB just below it to force failure.
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	allocMB := bytesToMBInt64(ms.Alloc)
	if allocMB < 5 { // safety: environment too small? skip deterministically
		t.Skipf("allocation unexpectedly small (%d MB); skipping limit test", allocMB)
	}
	eng.config.Memory.MaxMemoryMB = allocMB - 1 // set limit below observed usage
	if err := eng.OptimizeMemory(context.Background()); err == nil {
		t.Fatalf("expected memory limit error with alloc=%dMB limit=%dMB, got nil", allocMB, eng.config.Memory.MaxMemoryMB)
	}
}

// TestOptimizeMemory_GCThreshold ensures no error when below max but above GC threshold.
func TestOptimizeMemory_GCThreshold(t *testing.T) {
	cfg := &PerformanceConfig{Enabled: true, Memory: &MemoryConfig{MaxMemoryMB: 512, GCThresholdMB: 1}, Parallel: &ParallelConfig{WorkerCount: 1, WorkerTimeout: time.Second}, Cache: &CacheConfig{MaxSize: 10}}
	eng := NewPerformanceEngine(cfg)
	if err := eng.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Allocate ~2MB so it exceeds GCThresholdMB but well below MaxMemoryMB.
	buf := make([]byte, 2<<20)
	_ = buf
	if err := eng.OptimizeMemory(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetMetrics_EnabledAndDisabled validates metrics population based on Enabled flag.
func TestGetMetrics_EnabledAndDisabled(t *testing.T) {
	// Enabled engine
	engEnabled := NewPerformanceEngine(&PerformanceConfig{Enabled: true, Memory: &MemoryConfig{MaxMemoryMB: 64, GCThresholdMB: 32}, Parallel: &ParallelConfig{WorkerCount: 7, WorkerTimeout: time.Second}, Cache: &CacheConfig{MaxSize: 5}})
	if err := engEnabled.Start(context.Background()); err != nil {
		t.Fatalf("start enabled: %v", err)
	}
	mEnabled := engEnabled.GetMetrics()
	if mEnabled.MemoryStats == nil || mEnabled.ProcessorStats == nil {
		t.Fatalf("expected memory & processor stats populated")
	}
	if mEnabled.ProcessorStats.WorkerCount != 7 {
		t.Fatalf("expected worker count 7, got %d", mEnabled.ProcessorStats.WorkerCount)
	}
	if mEnabled.CacheStats == nil {
		t.Fatalf("expected cache stats populated")
	}

	// Disabled engine
	engDisabled := NewPerformanceEngine(&PerformanceConfig{Enabled: false})
	if err := engDisabled.Start(context.Background()); err != nil {
		t.Fatalf("start disabled: %v", err)
	}
	mDisabled := engDisabled.GetMetrics()
	if mDisabled.MemoryStats != nil || mDisabled.CacheStats != nil || mDisabled.ProcessorStats != nil {
		t.Fatalf("expected nil stats when disabled")
	}
}
