#!/bin/bash

#  CostScope Architectural Optimization - Comprehensive Analysis
# Phase: Performance Analysis & Code Quality Assessment

set -e

echo " COSTSCOPE ARCHITECTURAL OPTIMIZATION SUITE"
echo "=============================================="
echo " Phase: Performance Profiling & Code Analysis"
echo " Started: $(date)"
echo ""

# Default pinned golangci-lint version (can be overridden via env WANT_VERSION)
WANT_VERSION="${WANT_VERSION:-v1.54.2}"

# Detect CI environments where we should enforce the pinned binary for determinism
CI_ENFORCE_PINNED=false
if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ]; then
    CI_ENFORCE_PINNED=true
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create results directory
RESULTS_DIR="optimization_results_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULTS_DIR"

echo " Results will be saved to: $RESULTS_DIR"
echo ""

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

# Function to print error
print_error() {
    echo -e "${RED} $1${NC}"
}

# 1. PROJECT ANALYSIS
print_section " 1. PROJECT STRUCTURE ANALYSIS"

echo " Analyzing project structure..."
find . -name "*.go" -not -path "./_archive/*" | wc -l > "$RESULTS_DIR/go_files_count.txt"
GO_FILES=$(cat "$RESULTS_DIR/go_files_count.txt")
echo "   Go files: $GO_FILES"

# Lines of code analysis
echo " Calculating lines of code..."
find . -name "*.go" -not -path "./_archive/*" -exec wc -l {} + | tail -1 > "$RESULTS_DIR/total_loc.txt"
TOTAL_LOC=$(cat "$RESULTS_DIR/total_loc.txt" | awk '{print $1}')
echo "   Total LOC: $TOTAL_LOC"

# Package analysis
echo " Analyzing packages..."
find . -name "*.go" -not -path "./_archive/*" -exec dirname {} \; | sort | uniq > "$RESULTS_DIR/packages.txt"
PACKAGE_COUNT=$(cat "$RESULTS_DIR/packages.txt" | wc -l)
echo "   Packages: $PACKAGE_COUNT"

print_success "Project structure analysis completed"
echo ""

# 2. DEPENDENCY ANALYSIS
print_section " 2. DEPENDENCY ANALYSIS"

echo " Analyzing Go modules..."
go mod graph > "$RESULTS_DIR/dependency_graph.txt" 2>/dev/null || echo "Warning: Could not generate dependency graph"
go list -m all > "$RESULTS_DIR/all_modules.txt"
DIRECT_DEPS=$(grep -v "^$(go list -m)$" "$RESULTS_DIR/all_modules.txt" | wc -l)
echo "   Direct dependencies: $DIRECT_DEPS"

echo " Checking for unused dependencies..."
go mod tidy
go mod why -m $(go list -m all | grep -v "^$(go list -m)$" | awk '{print $1}') > "$RESULTS_DIR/dependency_usage.txt" 2>/dev/null || echo "Some dependencies could not be analyzed"

print_success "Dependency analysis completed"
echo ""

# 3. PERFORMANCE PROFILING
print_section " 3. PERFORMANCE PROFILING"

echo " Running performance profiler..."
cd scripts
if go run performance_profiler.go; then
    print_success "Performance profiling completed"
    mv cpu.prof "../$RESULTS_DIR/"
    mv mem.prof "../$RESULTS_DIR/"
else
    print_error "Performance profiling failed"
fi
cd ..

echo ""

# 4. CODE QUALITY ANALYSIS
print_section " 4. CODE QUALITY ANALYSIS"

# Install/check linting tools
echo " Checking linting tools..."
if ! command -v golangci-lint &> /dev/null; then
    print_warning "golangci-lint not found, attempting installation of pinned $WANT_VERSION..."
    if command -v brew &> /dev/null; then
        brew install golangci-lint || true
    else
        # try go install using the pinned version
        if ! go install github.com/golangci/golangci-lint/cmd/golangci-lint@${WANT_VERSION} 2>/dev/null; then
            # fallback to upstream installer that installs to GOPATH/bin
            if command -v curl >/dev/null 2>&1; then
                curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -o /tmp/golangci-install.sh && sh /tmp/golangci-install.sh -b $(go env GOPATH)/bin ${WANT_VERSION} || true
            fi
        fi
    fi
fi

echo " Running golangci-lint analysis... (preferring pinned ${WANT_VERSION})"

# Prefer a pinned/downloaded golangci-lint release under /tmp for deterministic behavior
REL_VER="${WANT_VERSION#v}"
TMPDIR="/tmp/golangci-${REL_VER}"
mkdir -p "$TMPDIR"
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64) ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *) ARCH="$ARCH_RAW" ;;
esac
TAR_NAME="golangci-lint-${REL_VER}-${OS}-${ARCH}.tar.gz"
TAR_URL="https://github.com/golangci/golangci-lint/releases/download/${WANT_VERSION}/${TAR_NAME}"
TAR_PATH="$TMPDIR/$TAR_NAME"

# Reuse any previously extracted binary
EX_BIN="$(find "$TMPDIR" -type f -name golangci-lint -print -quit 2>/dev/null || true)"
if [ -z "$EX_BIN" ]; then
    if command -v curl >/dev/null 2>&1; then
        echo "Attempting to download pinned golangci-lint from $TAR_URL"
        if curl -fsSL "$TAR_URL" -o "$TAR_PATH"; then
            tar -xzf "$TAR_PATH" -C "$TMPDIR" || true
            EX_BIN="$(find "$TMPDIR" -type f -name golangci-lint -print -quit 2>/dev/null || true)"
        else
            echo "Could not download $TAR_URL; will fall back to system binary if present"
        fi
    else
        echo "curl not available; will fall back to system binary if present"
    fi
fi

if [ -n "$EX_BIN" ]; then
    GOLANGCI_BIN="$EX_BIN"
    echo "Using pinned golangci-lint at $GOLANGCI_BIN"
else
    if command -v golangci-lint >/dev/null 2>&1; then
        GOLANGCI_BIN="$(command -v golangci-lint)"
        echo "Using system golangci-lint at $GOLANGCI_BIN"
    else
        if [ "$CI_ENFORCE_PINNED" = true ]; then
            echo "ERROR: golangci-lint pinned binary ${WANT_VERSION} not available in CI; failing for determinism" >&2
            exit 1
        fi
        echo "golangci-lint not found; skipping lint step"
        GOLANGCI_BIN=""
    fi
fi

if [ -n "$GOLANGCI_BIN" ]; then
    LINT_ERR_FILE="$RESULTS_DIR/lint_report.err"
    rm -f "$LINT_ERR_FILE"
    # Run with --no-config to avoid malformed project configs causing failures; this produces machine-readable JSON
    "$GOLANGCI_BIN" run --no-config --out-format=json ./... > "$RESULTS_DIR/lint_report.json" 2> "$LINT_ERR_FILE" || echo "Lint completed with issues (see $LINT_ERR_FILE)"
    # Also produce a human-readable text output
    "$GOLANGCI_BIN" run --no-config > "$RESULTS_DIR/lint_report.txt" 2>> "$LINT_ERR_FILE" || true
fi

# Count issues
if [ -f "$RESULTS_DIR/lint_report.json" ]; then
    LINT_ISSUES=$(jq '.Issues | length' "$RESULTS_DIR/lint_report.json" 2>/dev/null || echo "0")
    echo "   Lint issues found: $LINT_ISSUES"
else
    echo "   Lint report not generated"
fi

print_success "Code quality analysis completed"
echo ""

# 5. SECURITY ANALYSIS
print_section " 5. SECURITY ANALYSIS"

echo "️  Installing gosec..."
if ! command -v gosec &> /dev/null; then
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
fi

echo " Running security analysis..."
gosec -fmt json -out "$RESULTS_DIR/security_report.json" ./... 2>/dev/null || echo "Security scan completed with findings"
gosec -fmt text -out "$RESULTS_DIR/security_report.txt" ./... 2>/dev/null || echo "Security scan completed with findings"

# Count security issues
if [ -f "$RESULTS_DIR/security_report.json" ]; then
    SECURITY_ISSUES=$(jq '.Issues | length' "$RESULTS_DIR/security_report.json" 2>/dev/null || echo "0")
    echo "   Security issues found: $SECURITY_ISSUES"
else
    echo "   Security report not generated"
fi

print_success "Security analysis completed"
echo ""

# 6. DEAD CODE DETECTION
print_section " 6. DEAD CODE DETECTION"

echo " Analyzing unused code..."
# Find potentially unused functions and variables
echo "Finding exported but potentially unused symbols..."
# Initialize diagnostics file for exported-symbols analysis
DIAG_FILE="$RESULTS_DIR/exported_symbols.err"
rm -f "$DIAG_FILE" || true
touch "$DIAG_FILE" || true
# Use find -type f so directories are not passed to grep and avoid warnings.
# Redirect grep stderr to diagnostics file (append) so console stays clean but
# diagnostics are preserved for later inspection.
go list -f '{{.Dir}}' ./... | xargs -I {} find {} -type f -name "*.go" -exec grep -l -E '^func [A-Z]' {} + > "$RESULTS_DIR/exported_functions.txt" 2>>"$DIAG_FILE" || true
go list -f '{{.Dir}}' ./... | xargs -I {} find {} -type f -name "*.go" -exec grep -l -E '^var [A-Z]' {} + > "$RESULTS_DIR/exported_vars.txt" 2>>"$DIAG_FILE" || true

# Check for unused imports
echo "Checking for unused imports..."
find . -name "*.go" -not -path "./_archive/*" -exec goimports -l {} \; > "$RESULTS_DIR/unused_imports.txt" 2>/dev/null || echo "goimports check completed"

print_success "Dead code detection completed"
echo ""

# 7. COMPLEXITY ANALYSIS
print_section " 7. COMPLEXITY ANALYSIS"

echo " Analyzing cyclomatic complexity..."
if ! command -v gocyclo &> /dev/null; then
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
fi

find . -name "*.go" -not -path "./_archive/*" | xargs gocyclo -over 10 > "$RESULTS_DIR/complexity_report.txt" || echo "No high complexity functions found"
COMPLEX_FUNCTIONS=$(cat "$RESULTS_DIR/complexity_report.txt" | wc -l)
echo "   High complexity functions (>10): $COMPLEX_FUNCTIONS"

print_success "Complexity analysis completed"
echo ""

# 8. TEST COVERAGE ANALYSIS
print_section " 8. TEST COVERAGE ANALYSIS"

echo " Analyzing test coverage..."
go test -coverprofile="$RESULTS_DIR/coverage.out" ./... > "$RESULTS_DIR/test_output.txt" 2>&1 || echo "Some tests failed"
if [ -f "$RESULTS_DIR/coverage.out" ]; then
    go tool cover -html="$RESULTS_DIR/coverage.out" -o "$RESULTS_DIR/coverage.html"
    COVERAGE=$(go tool cover -func="$RESULTS_DIR/coverage.out" | tail -1 | awk '{print $3}')
    echo "   Overall coverage: $COVERAGE"
else
    echo "   Coverage report not generated"
fi

print_success "Test coverage analysis completed"
echo ""

# 9. MEMORY USAGE ANALYSIS
print_section " 9. MEMORY USAGE ANALYSIS"

echo " Building application for memory analysis..."
go build -o "$RESULTS_DIR/costscope_optimized" ./main.go

echo " Analyzing binary size..."
ls -lh "$RESULTS_DIR/costscope_optimized" | awk '{print $5}' > "$RESULTS_DIR/binary_size.txt"
BINARY_SIZE=$(cat "$RESULTS_DIR/binary_size.txt")
echo "   Binary size: $BINARY_SIZE"

# Strip binary for size comparison
echo "️  Creating stripped binary..."
go build -ldflags="-w -s" -o "$RESULTS_DIR/costscope_stripped" ./main.go
ls -lh "$RESULTS_DIR/costscope_stripped" | awk '{print $5}' > "$RESULTS_DIR/stripped_size.txt"
STRIPPED_SIZE=$(cat "$RESULTS_DIR/stripped_size.txt")
echo "   Stripped binary size: $STRIPPED_SIZE"

print_success "Memory usage analysis completed"
echo ""

# 10. GENERATE OPTIMIZATION REPORT
print_section " 10. GENERATING OPTIMIZATION REPORT"

cat > "$RESULTS_DIR/optimization_summary.md" << EOF
#  CostScope Architectural Optimization Report

**Generated:** $(date)
**Phase:** Performance Analysis & Code Quality Assessment

##  Project Metrics

- **Go Files:** $GO_FILES
- **Total Lines of Code:** $TOTAL_LOC
- **Packages:** $PACKAGE_COUNT
- **Direct Dependencies:** $DIRECT_DEPS
- **Binary Size:** $BINARY_SIZE
- **Stripped Binary Size:** $STRIPPED_SIZE

##  Code Quality

- **Lint Issues:** $LINT_ISSUES
- **Security Issues:** $SECURITY_ISSUES
- **High Complexity Functions:** $COMPLEX_FUNCTIONS
- **Test Coverage:** ${COVERAGE:-"N/A"}

##  Optimization Opportunities

### 1. Performance Optimizations
- Review CPU and memory profiles in \`cpu.prof\` and \`mem.prof\`
- Analyze memory allocation patterns
- Optimize hot paths identified in profiling

### 2. Code Quality Improvements
- Address lint issues in \`lint_report.txt\`
- Reduce cyclomatic complexity in high-complexity functions
- Improve test coverage to >90%

### 3. Security Enhancements
- Review security findings in \`security_report.txt\`
- Implement secure coding practices
- Add security-focused tests

### 4. Architectural Improvements
- Remove dead code and unused imports
- Consolidate duplicate functionality
- Optimize dependency usage

##  Next Steps

1. **Immediate Actions:**
   - Fix critical lint and security issues
   - Optimize memory-intensive operations
   - Reduce binary size through dependency cleanup

2. **Medium-term Goals:**
   - Implement caching strategies
   - Optimize database queries
   - Improve error handling patterns

3. **Long-term Objectives:**
   - Achieve <15% memory usage reduction
   - Improve performance by >20%
   - Maintain >95% test coverage

##  Generated Files

- \`cpu.prof\` - CPU profiling data
- \`mem.prof\` - Memory profiling data
- \`lint_report.txt\` - Code quality issues
- \`security_report.txt\` - Security vulnerabilities
- \`coverage.html\` - Test coverage report
- \`complexity_report.txt\` - High complexity functions

EOF

print_success "Optimization report generated"
echo ""

# If diagnostics were produced for exported symbols, add a short note to the summary
if [ -n "${DIAG_FILE:-}" ] && [ -s "$DIAG_FILE" ]; then
    echo "Note: exported symbol analysis produced diagnostics at: $(basename "$DIAG_FILE")" >> "$RESULTS_DIR/optimization_summary.md"
    echo " (see $RESULTS_DIR/$(basename "$DIAG_FILE") for details)"
fi

# FINAL SUMMARY
print_section " ANALYSIS COMPLETE"

echo " Analysis Summary:"
echo "   ️  Files analyzed: $GO_FILES Go files"
echo "    Lines of code: $TOTAL_LOC"
echo "    Lint issues: $LINT_ISSUES"
echo "   ️  Security issues: $SECURITY_ISSUES"
echo "    Complex functions: $COMPLEX_FUNCTIONS"
echo "    Binary size: $BINARY_SIZE"
echo ""
echo " All results saved to: $RESULTS_DIR"
echo " Review optimization_summary.md for detailed recommendations"
echo ""

print_success "CostScope architectural optimization analysis completed!"
echo ""
echo " Next commands to run:"
echo "   go tool pprof $RESULTS_DIR/cpu.prof"
echo "   go tool pprof $RESULTS_DIR/mem.prof"
echo "   open $RESULTS_DIR/coverage.html"
echo "   cat $RESULTS_DIR/optimization_summary.md"
