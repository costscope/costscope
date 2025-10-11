package gcshared

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// Run executes a simple GC benchmark/test across the provided GOGC values.
// If gogcValues is nil or empty, a default set is used. Name is optional and
// is only used to prefix the banner line for easier tool identification.
func Run(name string, gogcValues []int) {
	if len(gogcValues) == 0 {
		gogcValues = []int{50, 100, 200, 300}
	}

	if name != "" {
		fmt.Printf("[%s] Testing different GOGC settings...\n", name)
	} else {
		fmt.Println("Testing different GOGC settings...")
	}

	for _, gogc := range gogcValues {
		testGCSettings(gogc)
	}

	// Reset to default
	debug.SetGCPercent(100)
}

func testGCSettings(gogc int) {
	debug.SetGCPercent(gogc)

	start := time.Now()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Simulate memory-intensive workload
	for i := 0; i < 10000; i++ {
		data := make([]map[string]interface{}, 100)
		for j := range data {
			data[j] = map[string]interface{}{
				"id":   i*100 + j,
				"data": make([]byte, 1024),
			}
		}

		if i%1000 == 0 {
			runtime.GC()
		}
	}

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	duration := time.Since(start)

	fmt.Printf("GOGC=%d: Duration=%v, GC_cycles=%d, Peak_memory=%dKB\n",
		gogc, duration, m2.NumGC-m1.NumGC, m2.Sys/1024)
}
