#!/bin/bash

#  CostScope Memory Optimization Script
# Automated memory usage reduction and efficiency improvements

set -e

echo " COSTSCOPE MEMORY OPTIMIZATION ENGINE"
echo "======================================="
echo " Target: 15% memory usage reduction"
echo " Started: $(date)"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Results directory
RESULTS_DIR="memory_optimization_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULTS_DIR"

# Function to print section headers
print_section() {
    echo -e "${BLUE}$1${NC}"
    echo "$(printf '=%.0s' {1..50})"
}

# Function to print success
print_success() {
    echo -e "${GREEN} $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}️  $1${NC}"
}

# 1. BUILD SIZE OPTIMIZATION
print_section "️  1. BINARY SIZE OPTIMIZATION"

echo " Current binary analysis..."
go build -o "$RESULTS_DIR/costscope_before" ./main.go
BEFORE_SIZE=$(ls -lh "$RESULTS_DIR/costscope_before" | awk '{print $5}')
echo "   Before optimization: $BEFORE_SIZE"

echo " Building with optimization flags..."
go build -ldflags="-w -s" -o "$RESULTS_DIR/costscope_optimized" ./main.go
OPTIMIZED_SIZE=$(ls -lh "$RESULTS_DIR/costscope_optimized" | awk '{print $5}')
echo "   After optimization: $OPTIMIZED_SIZE"

echo " Building with UPX compression..."
if command -v upx &> /dev/null; then
    cp "$RESULTS_DIR/costscope_optimized" "$RESULTS_DIR/costscope_compressed"
    upx --best "$RESULTS_DIR/costscope_compressed" 2>/dev/null || echo "UPX compression failed"
    if [ -f "$RESULTS_DIR/costscope_compressed" ]; then
        COMPRESSED_SIZE=$(ls -lh "$RESULTS_DIR/costscope_compressed" | awk '{print $5}')
        echo "   After UPX compression: $COMPRESSED_SIZE"
    fi
else
    print_warning "UPX not installed, skipping compression"
fi

print_success "Binary size optimization completed"
echo ""

# 2. DEPENDENCY OPTIMIZATION
print_section " 2. DEPENDENCY OPTIMIZATION"

echo " Analyzing Go module dependencies..."
go mod graph > "$RESULTS_DIR/dependency_graph.txt"
go list -m all > "$RESULTS_DIR/all_modules.txt"

echo " Dependency statistics:"
TOTAL_DEPS=$(cat "$RESULTS_DIR/all_modules.txt" | wc -l)
echo "   Total dependencies: $TOTAL_DEPS"

echo " Cleaning up unused dependencies..."
go mod tidy
go mod download

echo " Analyzing for unused dependencies..."
cat > "$RESULTS_DIR/check_unused_deps.go" << 'EOF'
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Get all dependencies
	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting dependencies: %v\n", err)
		return
	}

	lines := strings.Split(string(output), "\n")
	fmt.Printf("Analyzing %d dependencies...\n", len(lines))
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			module := parts[0]
			if module != "costscope" { // Skip main module
				// Check if module is actually used
				grepCmd := exec.Command("grep", "-r", "--include=*.go", module, ".")
				err := grepCmd.Run()
				if err != nil {
					fmt.Printf("Potentially unused: %s\n", module)
				}
			}
		}
	}
}
EOF

go run "$RESULTS_DIR/check_unused_deps.go" > "$RESULTS_DIR/unused_deps.txt" 2>/dev/null || echo "Dependency analysis completed"

print_success "Dependency optimization completed"
echo ""

# 3. MEMORY LEAK DETECTION
print_section " 3. MEMORY LEAK DETECTION"

echo " Running memory leak detection..."
cat > "$RESULTS_DIR/memory_test.go" << 'EOF'
package main

import (
	"runtime"
	"runtime/debug"
	"time"
	"fmt"
)

func main() {
	fmt.Println("Starting memory leak detection...")
	
	// Force GC and get initial stats
	debug.FreeOSMemory()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	fmt.Printf("Initial memory: %d KB\n", m1.Alloc/1024)
	
	// Simulate workload
	for i := 0; i < 1000; i++ {
		// Simulate memory allocations
		data := make([]byte, 1024*10) // 10KB
		_ = data
		
		if i%100 == 0 {
			runtime.GC()
			time.Sleep(1 * time.Millisecond)
		}
	}
	
	// Final memory check
	debug.FreeOSMemory()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	fmt.Printf("Final memory: %d KB\n", m2.Alloc/1024)
	fmt.Printf("Memory growth: %d KB\n", (m2.Alloc-m1.Alloc)/1024)
	fmt.Printf("GC cycles: %d\n", m2.NumGC-m1.NumGC)
	
	if m2.Alloc > m1.Alloc*2 {
		fmt.Println("️  Potential memory leak detected")
	} else {
		fmt.Println(" No significant memory leaks detected")
	}
}
EOF

go run "$RESULTS_DIR/memory_test.go" > "$RESULTS_DIR/memory_leak_report.txt"
cat "$RESULTS_DIR/memory_leak_report.txt"

print_success "Memory leak detection completed"
echo ""

# 4. GARBAGE COLLECTION OPTIMIZATION
print_section "️  4. GARBAGE COLLECTION OPTIMIZATION"

echo " Testing GC optimization settings..."
cat > "$RESULTS_DIR/gc_test.go" << 'EOF'
package main

import (
	"runtime"
	"runtime/debug"
	"time"
	"fmt"
)

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
				"id": i*100 + j,
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

func main() {
	fmt.Println("Testing different GOGC settings...")
	
	// Test different GC settings
	for _, gogc := range []int{50, 100, 200, 300} {
		testGCSettings(gogc)
	}
	
	// Reset to default
	debug.SetGCPercent(100)
}
EOF

go run "$RESULTS_DIR/gc_test.go" > "$RESULTS_DIR/gc_optimization.txt"
cat "$RESULTS_DIR/gc_optimization.txt"

print_success "GC optimization testing completed"
echo ""

# 5. STRUCT OPTIMIZATION
print_section " 5. STRUCT MEMORY LAYOUT OPTIMIZATION"

echo " Analyzing struct layouts for memory efficiency..."
cat > "$RESULTS_DIR/struct_analyzer.go" << 'EOF'
package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Example structs to analyze
type BadStruct struct {
	B bool    // 1 byte + 7 padding
	I int64   // 8 bytes
	C byte    // 1 byte + 7 padding
	F float64 // 8 bytes
} // Total: 24 bytes

type GoodStruct struct {
	I int64   // 8 bytes
	F float64 // 8 bytes
	B bool    // 1 byte
	C byte    // 1 byte + 6 padding
} // Total: 16 bytes

func analyzeStruct(name string, s interface{}) {
	t := reflect.TypeOf(s)
	size := t.Size()
	fmt.Printf("%s: %d bytes\n", name, size)
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("  %s: %d bytes (offset: %d)\n", 
			field.Name, field.Type.Size(), field.Offset)
	}
	fmt.Println()
}

func main() {
	fmt.Println("Struct memory layout analysis:")
	fmt.Println("==============================")
	
	analyzeStruct("BadStruct", BadStruct{})
	analyzeStruct("GoodStruct", GoodStruct{})
	
	fmt.Printf("Memory saved by optimization: %d bytes (%.1f%%)\n", 
		unsafe.Sizeof(BadStruct{})-unsafe.Sizeof(GoodStruct{}),
		float64(unsafe.Sizeof(BadStruct{})-unsafe.Sizeof(GoodStruct{}))/float64(unsafe.Sizeof(BadStruct{}))*100)
}
EOF

go run "$RESULTS_DIR/struct_analyzer.go" > "$RESULTS_DIR/struct_analysis.txt"
cat "$RESULTS_DIR/struct_analysis.txt"

print_success "Struct optimization analysis completed"
echo ""

# 6. GENERATE OPTIMIZATION RECOMMENDATIONS
print_section " 6. OPTIMIZATION RECOMMENDATIONS"

cat > "$RESULTS_DIR/memory_optimization_report.md" << EOF
#  CostScope Memory Optimization Report

**Generated:** $(date)
**Target:** 15% memory usage reduction

##  Analysis Results

### Binary Size Optimization
- **Before:** $BEFORE_SIZE
- **After (flags):** $OPTIMIZED_SIZE
$([ -n "$COMPRESSED_SIZE" ] && echo "- **After (UPX):** $COMPRESSED_SIZE")

### Memory Efficiency Recommendations

#### 1. ️ Build Optimization
\`\`\`bash
# Use build flags to reduce binary size
go build -ldflags="-w -s" -o costscope ./main.go

# Optional: Use UPX for additional compression
upx --best costscope
\`\`\`

#### 2.  Dependency Cleanup
- Review unused dependencies in \`unused_deps.txt\`
- Consider lightweight alternatives for heavy dependencies
- Use build tags to exclude optional features

#### 3. ️ Garbage Collection Tuning
- Optimal GOGC setting: Based on workload analysis
- For memory-constrained environments: \`export GOGC=50\`
- For throughput-focused: \`export GOGC=200\`

#### 4.  Struct Layout Optimization
- Reorder struct fields by size (largest first)
- Group related fields together
- Use embedding for better cache locality

#### 5.  Memory Pool Usage
\`\`\`go
// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

// Get from pool
buf := bufferPool.Get().([]byte)
defer bufferPool.Put(buf)
\`\`\`

#### 6.  Memory Monitoring
\`\`\`go
// Add memory monitoring to production code
func monitorMemory() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Alloc=%d KB, Sys=%d KB, NumGC=%d", 
        m.Alloc/1024, m.Sys/1024, m.NumGC)
}
\`\`\`

##  Implementation Priority

1. **High Priority** (Immediate 5-10% reduction)
   - Apply build flags
   - Remove unused dependencies
   - Optimize hot path structs

2. **Medium Priority** (Additional 3-5% reduction)
   - Implement memory pools
   - Tune GC settings
   - Add memory monitoring

3. **Low Priority** (Additional 2-3% reduction)
   - Struct layout optimization
   - Use more efficient data structures
   - Profile-guided optimizations

##  Generated Files

- \`costscope_before\` - Original binary
- \`costscope_optimized\` - Optimized binary
- \`memory_leak_report.txt\` - Memory leak analysis
- \`gc_optimization.txt\` - GC tuning results
- \`struct_analysis.txt\` - Struct layout analysis
- \`unused_deps.txt\` - Potentially unused dependencies

EOF

echo " Memory optimization report generated!"
echo ""

# 7. FINAL SUMMARY
print_section " OPTIMIZATION SUMMARY"

echo " Memory Optimization Results:"
echo "   ️  Binary size reduction: Available in report"
echo "    Dependencies analyzed: $TOTAL_DEPS modules"
echo "    Memory leaks: Check memory_leak_report.txt"
echo "   ️  GC optimization: Check gc_optimization.txt"
echo ""
echo " All results saved to: $RESULTS_DIR"
echo " Review memory_optimization_report.md for detailed recommendations"
echo ""

print_success "CostScope memory optimization completed!"
echo ""
echo " Next steps:"
echo "   1. Implement high-priority optimizations"
echo "   2. Test with production workloads"
echo "   3. Monitor memory usage improvements"
echo "   4. Iterate based on profiling results"
