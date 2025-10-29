package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// PerformanceProfiler handles comprehensive performance analysis
type PerformanceProfiler struct {
	cpuProfile *os.File
	memProfile *os.File
	startTime  time.Time
	results    *ProfileResults
}

// ProfileResults stores profiling metrics
type ProfileResults struct {
	CPUUsage     float64
	MemoryUsage  uint64
	AllocObjects uint64
	GCPauses     []time.Duration
	Goroutines   int
	Duration     time.Duration
	BenchmarkOps int64
}

// NewPerformanceProfiler creates a new profiler instance
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		results: &ProfileResults{},
	}
}

// StartProfiling begins CPU and memory profiling
func (p *PerformanceProfiler) StartProfiling() error {
	var err error

	// Enable block/mutex profiling at runtime
	runtime.SetBlockProfileRate(100) // 1 in 100 blocking events
	runtime.SetMutexProfileFraction(10)

	// Start CPU profiling
	p.cpuProfile, err = os.Create("cpu.prof")
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}

	if err := pprof.StartCPUProfile(p.cpuProfile); err != nil {
		return fmt.Errorf("could not start CPU profile: %v", err)
	}

	p.startTime = time.Now()
	return nil
}

// StopProfiling ends profiling and generates memory profile
func (p *PerformanceProfiler) StopProfiling() error {
	p.results.Duration = time.Since(p.startTime)

	// Stop CPU profiling
	pprof.StopCPUProfile()
	if err := p.cpuProfile.Close(); err != nil {
		return fmt.Errorf("failed to close CPU profile: %v", err)
	}

	// Create memory profile
	var err error
	p.memProfile, err = os.Create("mem.prof")
	if err != nil {
		return fmt.Errorf("could not create memory profile: %v", err)
	}
	defer func() {
		if cerr := p.memProfile.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close memory profile: %v\n", cerr)
		}
	}()

	runtime.GC() // Force GC to get accurate memory stats
	if err := pprof.WriteHeapProfile(p.memProfile); err != nil {
		return fmt.Errorf("could not write memory profile: %v", err)
	}

	// Collect runtime statistics
	p.collectRuntimeStats()

	// Emit block and mutex profiles
	if err := p.writeProfile("block.prof", "block"); err != nil {
		fmt.Printf("Warning: failed to write block profile: %v\n", err)
	}
	if err := p.writeProfile("mutex.prof", "mutex"); err != nil {
		fmt.Printf("Warning: failed to write mutex profile: %v\n", err)
	}

	return nil
}

// collectRuntimeStats gathers runtime performance metrics
func (p *PerformanceProfiler) collectRuntimeStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	p.results.MemoryUsage = m.Alloc
	p.results.AllocObjects = m.Mallocs - m.Frees
	p.results.Goroutines = runtime.NumGoroutine()

	// Collect GC pause times (last 256 pauses)
	p.results.GCPauses = make([]time.Duration, 0, len(m.PauseNs))
	for i := 0; i < len(m.PauseNs); i++ {
		if m.PauseNs[i] > 0 {
			// Safe conversion with bounds checking
			if m.PauseNs[i] <= uint64(math.MaxInt64) {
				p.results.GCPauses = append(p.results.GCPauses, time.Duration(m.PauseNs[i])) // #nosec G115 -- bounded above by MaxInt64
			} else {
				p.results.GCPauses = append(p.results.GCPauses, time.Duration(math.MaxInt64)) // clamp
			}
		}
	}
}

// PrintResults displays comprehensive profiling results
func (p *PerformanceProfiler) PrintResults() {
	fmt.Println("\n COSTSCOPE PERFORMANCE PROFILING RESULTS")
	fmt.Println("==================================================")

	fmt.Printf("⏱️  Execution Time: %v\n", p.results.Duration)
	fmt.Printf(" Memory Usage: %d bytes (%.2f MB)\n",
		p.results.MemoryUsage, float64(p.results.MemoryUsage)/1024/1024)
	fmt.Printf(" Allocated Objects: %d\n", p.results.AllocObjects)
	fmt.Printf(" Active Goroutines: %d\n", p.results.Goroutines)

	if len(p.results.GCPauses) > 0 {
		var totalGC time.Duration
		maxGC := time.Duration(0)
		for _, pause := range p.results.GCPauses {
			totalGC += pause
			if pause > maxGC {
				maxGC = pause
			}
		}
		avgGC := totalGC / time.Duration(len(p.results.GCPauses))
		fmt.Printf("️  GC Pauses: %d (avg: %v, max: %v)\n",
			len(p.results.GCPauses), avgGC, maxGC)
	}

	fmt.Println("\n PROFILING FILES GENERATED:")
	fmt.Println("   - cpu.prof (CPU profiling)")
	fmt.Println("   - mem.prof (Memory profiling)")
	fmt.Println("   - block.prof (Blocking events)")
	fmt.Println("   - mutex.prof (Mutex contention)")
	fmt.Println("\n ANALYSIS COMMANDS:")
	fmt.Println("   go tool pprof cpu.prof")
	fmt.Println("   go tool pprof mem.prof")
	fmt.Println("   go tool pprof block.prof")
	fmt.Println("   go tool pprof mutex.prof")
	fmt.Println("   go tool pprof -http=:8080 cpu.prof")
}

// writeProfile writes the specified profile by name
func (p *PerformanceProfiler) writeProfile(filename, profName string) error {
	// Security: filename is selected from fixed set (cpu.prof, mem.prof, block.prof, mutex.prof).
	//nolint:gosec // G304: controlled internal filenames
	f, err := os.Create(filename) // #nosec G304
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	prof := pprof.Lookup(profName)
	if prof == nil {
		return fmt.Errorf("profile %s not found", profName)
	}
	return prof.WriteTo(f, 0)
}

// BenchmarkFOCUSProcessing runs FOCUS processing benchmarks
func (p *PerformanceProfiler) BenchmarkFOCUSProcessing() {
	fmt.Println(" Running FOCUS Processing Benchmark...")

	startTime := time.Now()

	// Simulate FOCUS processing workload
	for i := 0; i < 10000; i++ {
		// Simulate memory allocation patterns in FOCUS processing
		data := make([]byte, 1024*10) // 10KB allocations
		_ = data

		if i%1000 == 0 {
			runtime.GC() // Periodic GC to simulate real workload
		}
	}

	p.results.BenchmarkOps = 10000
	elapsed := time.Since(startTime)
	fmt.Printf("   Completed %d operations in %v\n", p.results.BenchmarkOps, elapsed)
	fmt.Printf("   Rate: %.2f ops/sec\n", float64(p.results.BenchmarkOps)/elapsed.Seconds())
}

func main() {
	profiler := NewPerformanceProfiler()

	fmt.Println(" COSTSCOPE PERFORMANCE PROFILING SUITE")
	fmt.Println("Starting comprehensive performance analysis...")

	// Start profiling
	if err := profiler.StartProfiling(); err != nil {
		log.Fatal(err)
	}

	// Run benchmarks
	profiler.BenchmarkFOCUSProcessing()

	// Simulate additional workloads
	fmt.Println(" Simulating concurrent workloads...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start multiple goroutines to simulate concurrent processing
	for i := 0; i < 10; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate work
					data := make([]map[string]interface{}, 100)
					for j := range data {
						data[j] = map[string]interface{}{
							"id":    j,
							"value": fmt.Sprintf("worker-%d-item-%d", id, j),
						}
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for workload completion
	<-ctx.Done()

	// Stop profiling and generate results
	if err := profiler.StopProfiling(); err != nil {
		log.Fatal(err)
	}

	profiler.PrintResults()

	fmt.Println("\n Performance profiling completed!")
	fmt.Println(" Next steps: Analyze profiles and identify optimization opportunities")
}
