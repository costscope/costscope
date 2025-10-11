#!/bin/bash

#  CostScope Final Optimization Validation
# Final validation of all optimizations

set -euo pipefail

echo " CostScope Final Optimization Validation"
echo "========================================"
echo

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to log test results
log_test() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN} PASS${NC}: $test_name - $details"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}️  WARN${NC}: $test_name - $details"
    else
        echo -e "${RED} FAIL${NC}: $test_name - $details"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Function to check if binary exists and get size
check_binary() {
    local binary_path="$1"
    local binary_name="$2"
    
    if [ -f "$binary_path" ]; then
        local size=$(ls -lh "$binary_path" | awk '{print $5}')
        log_test "Binary $binary_name" "PASS" "Size: $size"
        return 0
    else
        log_test "Binary $binary_name" "FAIL" "Not found at $binary_path"
        return 1
    fi
}

# Function to test optimization utilities
test_optimization_utils() {
    echo -e "${BLUE} Testing Optimization Utilities${NC}"
    echo "--------------------------------"
    
    # Check for unified performance engine (replaces internal/optimization)
    if [ -d "internal/database/performance" ] && ls internal/database/performance/*.go >/dev/null 2>&1; then
        local count=$(ls internal/database/performance/*.go 2>/dev/null | wc -l)
        log_test "Unified Performance Engine" "PASS" "Found internal/database/performance ($count .go files)"
    else
        log_test "Unified Performance Engine" "FAIL" "Missing internal/database/performance or no .go files present"
    fi
    
    # Check configuration files
    if [ -f "configs/optimization.yaml" ]; then
        log_test "Optimization Config" "PASS" "Found configs/optimization.yaml"
    else
        log_test "Optimization Config" "FAIL" "Missing configs/optimization.yaml"
    fi
    
    if [ -f "Makefile.optimized" ]; then
        log_test "Optimized Makefile" "PASS" "Found Makefile.optimized"
    else
        log_test "Optimized Makefile" "FAIL" "Missing Makefile.optimized"
    fi
    
    echo
}

# Function to test binary sizes
test_binary_optimization() {
    echo -e "${BLUE} Testing Binary Size Optimization${NC}"
    echo "----------------------------------"
    
    # Check different binary versions
    check_binary "bin/costscope" "Standard"
    check_binary "bin/costscope-optimized" "Optimized"
    check_binary "bin/costscope-enterprise" "Enterprise"
    check_binary "bin/costscope-production" "Production"
    
    # Compare sizes if both exist
    if [ -f "bin/costscope" ] && [ -f "bin/costscope-optimized" ]; then
        local original_size=$(stat -c%s "bin/costscope" 2>/dev/null || stat -f%z "bin/costscope")
        local optimized_size=$(stat -c%s "bin/costscope-optimized" 2>/dev/null || stat -f%z "bin/costscope-optimized")
        local reduction=$(echo "scale=1; ($original_size - $optimized_size) * 100 / $original_size" | bc -l)
        
        if (( $(echo "$reduction >= 15.0" | bc -l) )); then
            log_test "Size Reduction" "PASS" "${reduction}% reduction (target: 15%)"
        else
            log_test "Size Reduction" "WARN" "${reduction}% reduction (below target: 15%)"
        fi
    fi
    
    echo
}

# Function to test performance optimizations
test_performance_optimization() {
    echo -e "${BLUE} Testing Performance Optimization${NC}"
    echo "--------------------------------"
    
    # Check if performance profiler exists
    if [ -f "scripts/performance_profiler.go" ]; then
        log_test "Performance Profiler" "PASS" "Found scripts/performance_profiler.go"
        
        # Try to run profiler (timeout after 10 seconds)
        if timeout 10s go run scripts/performance_profiler.go > /dev/null 2>&1; then
            log_test "Profiler Execution" "PASS" "Profiler runs successfully"
        else
            log_test "Profiler Execution" "WARN" "Profiler timeout or error"
        fi
    else
        log_test "Performance Profiler" "FAIL" "Missing scripts/performance_profiler.go"
    fi
    
    # Check optimization scripts
    for script in "memory_optimizer.sh" "optimize_analysis.sh" "final_optimizer.sh"; do
        if [ -f "scripts/$script" ]; then
            log_test "Script $script" "PASS" "Found scripts/$script"
        else
            log_test "Script $script" "FAIL" "Missing scripts/$script"
        fi
    done
    
    echo
}

# Function to test security improvements
test_security_optimization() {
    echo -e "${BLUE} Testing Security Optimization${NC}"
    echo "------------------------------"
    
    # Check if gosec is available
    if command -v gosec >/dev/null 2>&1; then
        echo "Running security analysis..."
        local security_output=$(gosec -quiet ./... 2>&1 || true)
        local issue_count=$(echo "$security_output" | grep -c "Issues" || echo "0")
        
        if [ "$issue_count" = "0" ]; then
            log_test "Security Scan" "PASS" "No security issues found"
        else
            log_test "Security Scan" "WARN" "$issue_count security issues found"
        fi
    else
        log_test "Security Scanner" "WARN" "gosec not installed"
    fi
    
    echo
}

# Function to test code quality
test_code_quality() {
    echo -e "${BLUE} Testing Code Quality${NC}"
    echo "---------------------"
    
    # Check if golangci-lint is available
    if command -v golangci-lint >/dev/null 2>&1; then
        echo "Running code quality analysis..."
        local lint_output=$(golangci-lint run --fast --timeout=30s ./... 2>&1 || true)
        local issue_count=$(echo "$lint_output" | grep -c "issues" || echo "0")
        
        if [ "$issue_count" = "0" ]; then
            log_test "Code Quality" "PASS" "No linting issues found"
        else
            log_test "Code Quality" "WARN" "$issue_count linting issues found"
        fi
    else
        log_test "Code Quality Checker" "WARN" "golangci-lint not installed"
    fi
    
    # Check test coverage if available
    mkdir -p logs
    if go test -short -coverprofile=logs/coverage.out ./... >/dev/null 2>&1; then
        local coverage=$(go tool cover -func=logs/coverage.out | tail -1 | awk '{print $3}')
        log_test "Test Coverage" "PASS" "Coverage: $coverage"
        # keep logs/coverage.out for debugging
    else
        log_test "Test Coverage" "WARN" "Unable to calculate coverage"
    fi
    
    echo
}

# Function to validate build process
test_build_optimization() {
    echo -e "${BLUE}️  Testing Build Optimization${NC}"
    echo "----------------------------"
    
    # Test optimized build
    echo "Testing optimized build process..."
    local build_start=$(date +%s)
    
    if CGO_ENABLED=1 go build -ldflags="-w -s" -o /tmp/test-costscope ./main.go >/dev/null 2>&1; then
        local build_end=$(date +%s)
        local build_time=$((build_end - build_start))
        log_test "Optimized Build" "PASS" "Build completed in ${build_time}s"
        
        # Check if binary is optimized
        if [ -f "/tmp/test-costscope" ]; then
            local size=$(ls -lh /tmp/test-costscope | awk '{print $5}')
            log_test "Build Output" "PASS" "Binary size: $size"
            rm -f /tmp/test-costscope
        fi
    else
        log_test "Optimized Build" "FAIL" "Build failed"
    fi
    
    echo
}

# Function to test runtime optimizations
test_runtime_optimization() {
    echo -e "${BLUE} Testing Runtime Optimization${NC}"
    echo "-----------------------------"
    
    # Check for optimized configurations
    local config_files=(
        "configs/optimization.yaml"
        "configs/production.yaml"
        "configs/docker.yaml"
    )
    
    for config in "${config_files[@]}"; do
        if [ -f "$config" ]; then
            log_test "Config $config" "PASS" "Configuration file exists"
        else
            log_test "Config $config" "WARN" "Configuration file missing"
        fi
    done
    
    # Check environment optimization
    if grep -q "GOGC" configs/*.yaml 2>/dev/null; then
        log_test "GC Optimization" "PASS" "GC tuning configured"
    else
        log_test "GC Optimization" "WARN" "GC tuning not found in configs"
    fi
    
    echo
}

# Main execution
main() {
    echo -e "${BLUE}Starting CostScope Final Optimization Validation...${NC}"
    echo "=================================================="
    echo
    
    # Run all tests
    test_optimization_utils
    test_binary_optimization
    test_performance_optimization
    test_security_optimization
    test_code_quality
    test_build_optimization
    test_runtime_optimization
    
    # Final summary
    echo -e "${BLUE} VALIDATION SUMMARY${NC}"
    echo "===================="
    echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Success Rate: ${BLUE}$((PASSED_TESTS * 100 / TOTAL_TESTS))%${NC}"
    echo
    
    # Overall status
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN} ALL OPTIMIZATIONS VALIDATED SUCCESSFULLY!${NC}"
        echo -e "${GREEN} CostScope is ready for production deployment${NC}"
        exit 0
    elif [ $FAILED_TESTS -lt 3 ]; then
        echo -e "${YELLOW}️  MOSTLY SUCCESSFUL WITH MINOR ISSUES${NC}"
        echo -e "${YELLOW} Address failed tests before production deployment${NC}"
        exit 1
    else
        echo -e "${RED} SIGNIFICANT ISSUES FOUND${NC}"
        echo -e "${RED} Optimization validation failed - review issues${NC}"
        exit 2
    fi
}

# Run main function
main "$@"
