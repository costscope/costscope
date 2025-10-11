#!/bin/bash

#  CostScope Production Deployment Script
# Final deployment with optimizations

set -euo pipefail

echo " CostScope Production Deployment"
echo "=================================="
echo

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEPLOYMENT_ENV=${1:-production}
BINARY_NAME="costscope"
OPTIMIZED_BINARY="costscope-optimized"
BACKUP_DIR="backup_$(date +%Y%m%d_%H%M%S)"

# Functions
log_step() {
    echo -e "${BLUE} $1${NC}"
}

log_success() {
    echo -e "${GREEN} $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}️  $1${NC}"
}

log_error() {
    echo -e "${RED} $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."
    
    # Check if optimized binary exists
    if [ ! -f "bin/$OPTIMIZED_BINARY" ]; then
        log_error "Optimized binary not found: bin/$OPTIMIZED_BINARY"
        echo "Run: CGO_ENABLED=1 go build -ldflags='-w -s' -o bin/$OPTIMIZED_BINARY ./main.go"
        exit 1
    fi
    
    # Check if optimization config exists
    if [ ! -f "configs/optimization.yaml" ]; then
        log_error "Optimization config not found: configs/optimization.yaml"
        exit 1
    fi
    
    # Check if environment config exists
    if [ ! -f "configs/$DEPLOYMENT_ENV.yaml" ]; then
        log_error "Environment config not found: configs/$DEPLOYMENT_ENV.yaml"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Create backup
create_backup() {
    log_step "Creating backup..."
    
    mkdir -p "$BACKUP_DIR"
    
    # Backup current binary if exists
    if [ -f "bin/$BINARY_NAME" ]; then
        cp "bin/$BINARY_NAME" "$BACKUP_DIR/"
        log_success "Current binary backed up"
    fi
    
    # Backup configs
    cp -r configs/ "$BACKUP_DIR/"
    log_success "Configuration backed up to $BACKUP_DIR"
}

# Deploy optimized binary
deploy_binary() {
    log_step "Deploying optimized binary..."
    
    # Copy optimized binary
    cp "bin/$OPTIMIZED_BINARY" "bin/$BINARY_NAME"
    chmod +x "bin/$BINARY_NAME"
    
    # Verify binary
    if ./bin/$BINARY_NAME version >/dev/null 2>&1; then
        log_success "Optimized binary deployed successfully"
    else
        log_error "Binary deployment failed - rolling back"
        if [ -f "$BACKUP_DIR/$BINARY_NAME" ]; then
            cp "$BACKUP_DIR/$BINARY_NAME" "bin/"
        fi
        exit 1
    fi
    
    # Show binary size
    local size=$(ls -lh "bin/$BINARY_NAME" | awk '{print $5}')
    log_success "Binary size: $size"
}

# Configure runtime environment
configure_environment() {
    log_step "Configuring runtime environment..."
    
    # Set environment variables for optimization
    cat > .env.optimization << EOF
# CostScope Production Optimization Environment
export COSTSCOPE_ENV=$DEPLOYMENT_ENV
export COSTSCOPE_CONFIG=configs/$DEPLOYMENT_ENV.yaml

# Memory optimization
export GOGC=200
export GOMEMLIMIT=1GB
export GOMAXPROCS=0

# Performance monitoring
export COSTSCOPE_ENABLE_MONITORING=true
export COSTSCOPE_MONITORING_INTERVAL=30s

# Security
export COSTSCOPE_ENABLE_SECURITY=true
export COSTSCOPE_LOG_LEVEL=info

# Database optimization
export COSTSCOPE_DB_MAX_OPEN_CONNS=25
export COSTSCOPE_DB_MAX_IDLE_CONNS=5
export COSTSCOPE_DB_CONN_MAX_LIFETIME=5m

# Build info
export COSTSCOPE_BUILD_OPTIMIZED=true
export COSTSCOPE_BUILD_VERSION=$(date +%Y%m%d-%H%M%S)
EOF

    log_success "Environment configuration created: .env.optimization"
    log_warning "Source with: source .env.optimization"
}

# Start performance monitoring
start_monitoring() {
    log_step "Starting performance monitoring..."
    
    # Create monitoring script
    cat > scripts/start_monitoring.sh << 'EOF'
#!/bin/bash

# Start CostScope with performance monitoring
echo " Starting CostScope with performance monitoring..."

# Source optimization environment
if [ -f ".env.optimization" ]; then
    source .env.optimization
fi

# Create monitoring directory
mkdir -p monitoring/$(date +%Y%m%d)

# Start with profiling endpoints enabled
./bin/costscope \
    --config configs/production.yaml \
    --enable-monitoring \
    --pprof-enabled \
    --pprof-addr localhost:6060 \
    > logs/costscope_$(date +%Y%m%d_%H%M%S).log 2>&1 &

COSTSCOPE_PID=$!
echo $COSTSCOPE_PID > costscope.pid

echo " CostScope started with PID: $COSTSCOPE_PID"
echo " Profiling available at: http://localhost:6060/debug/pprof/"
echo " Log file: logs/costscope_$(date +%Y%m%d_%H%M%S).log"
echo
echo "To stop: kill $COSTSCOPE_PID"
echo "To monitor: tail -f logs/costscope_$(date +%Y%m%d_%H%M%S).log"
EOF

    chmod +x scripts/start_monitoring.sh
    log_success "Monitoring script created: scripts/start_monitoring.sh"
}

# Create health check script
create_health_check() {
    log_step "Creating health check script..."
    
    cat > scripts/health_check.sh << 'EOF'
#!/bin/bash

# CostScope Health Check Script
echo " CostScope Health Check"
echo "========================"

# Check if process is running
if [ -f "costscope.pid" ]; then
    PID=$(cat costscope.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo " Process running (PID: $PID)"
    else
        echo " Process not running"
        rm -f costscope.pid
        exit 1
    fi
else
    echo " PID file not found"
    exit 1
fi

# Check memory usage
echo " Memory Usage:"
ps -p $PID -o pid,ppid,pcpu,pmem,rss,vsz,comm --no-headers

# Check if profiling endpoint is accessible
if curl -s http://localhost:6060/debug/pprof/ > /dev/null; then
    echo " Profiling endpoint accessible"
else
    echo "️  Profiling endpoint not accessible"
fi

# Check log file
LOG_FILE=$(ls -t logs/costscope_*.log 2>/dev/null | head -1)
if [ -n "$LOG_FILE" ]; then
    echo " Latest log: $LOG_FILE"
    echo " Last 5 lines:"
    tail -5 "$LOG_FILE"
else
    echo "️  No log files found"
fi

echo
echo " Health check completed"
EOF

    chmod +x scripts/health_check.sh
    log_success "Health check script created: scripts/health_check.sh"
}

# Create performance benchmarking script
create_benchmark_script() {
    log_step "Creating benchmark script..."
    
    cat > scripts/benchmark_production.sh << 'EOF'
#!/bin/bash

# CostScope Production Benchmarking
echo " CostScope Production Benchmarking"
echo "====================================="

# Create benchmark directory
mkdir -p benchmarks/$(date +%Y%m%d)
BENCH_DIR="benchmarks/$(date +%Y%m%d)/bench_$(date +%H%M%S)"
mkdir -p "$BENCH_DIR"

echo " Running performance benchmarks..."

# CPU profiling
echo " CPU Profiling (30s)..."
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30 > /dev/null 2>&1 &
PPROF_PID=$!

# Memory profiling
echo " Memory Profiling..."
curl -s http://localhost:6060/debug/pprof/heap > "$BENCH_DIR/heap.prof"

# Goroutine profiling
echo " Goroutine Profiling..."
curl -s http://localhost:6060/debug/pprof/goroutine > "$BENCH_DIR/goroutine.prof"

# System stats
echo " System Statistics..."
cat > "$BENCH_DIR/system_stats.txt" << EOL
Timestamp: $(date)
System Load: $(uptime)
Memory Info: $(free -h | grep Mem)
CPU Info: $(grep 'cpu cores' /proc/cpuinfo | head -1)
Process Info: $(ps aux | grep costscope | grep -v grep)
EOL

# Wait for CPU profiling to complete
wait $PPROF_PID

echo " Benchmarking completed"
echo " Results saved to: $BENCH_DIR"
echo " View CPU profile: go tool pprof $BENCH_DIR/profile"
echo " View memory profile: go tool pprof $BENCH_DIR/heap.prof"
EOF

    chmod +x scripts/benchmark_production.sh
    log_success "Benchmark script created: scripts/benchmark_production.sh"
}

# Validate deployment
validate_deployment() {
    log_step "Validating deployment..."
    
    # Check binary version
    VERSION=$(./bin/$BINARY_NAME version 2>/dev/null | head -1 || echo "Unknown")
    log_success "Binary version: $VERSION"
    
    # Check binary size optimization
    ORIGINAL_SIZE=$(ls -l "bin/$OPTIMIZED_BINARY" | awk '{print $5}')
    OPTIMIZED_SIZE=$(echo "scale=1; $ORIGINAL_SIZE / 1024 / 1024" | bc -l)
    log_success "Binary size: ${OPTIMIZED_SIZE}MB"
    
    # Validate configuration
    if ./bin/$BINARY_NAME validate-config configs/$DEPLOYMENT_ENV.yaml >/dev/null 2>&1; then
        log_success "Configuration validation passed"
    else
        log_warning "Configuration validation failed or not supported"
    fi
    
    # Check unified performance engine modules (replaces internal/optimization)
    if [ -d "internal/database/performance" ]; then
        local modules=$(ls internal/database/performance/*.go 2>/dev/null | wc -l)
        log_success "Unified performance engine modules: $modules files"
    else
        log_warning "Unified performance engine not found: internal/database/performance"
    fi
    
    log_success "Deployment validation completed"
}

# Create rollback script
create_rollback_script() {
    log_step "Creating rollback script..."
    
    cat > scripts/rollback.sh << EOF
#!/bin/bash

# CostScope Rollback Script
echo " CostScope Rollback"
echo "==================="

# Stop current process
if [ -f "costscope.pid" ]; then
    PID=\$(cat costscope.pid)
    if ps -p \$PID > /dev/null 2>&1; then
        echo " Stopping current process (PID: \$PID)..."
        kill \$PID
        sleep 5
        if ps -p \$PID > /dev/null 2>&1; then
            echo " Force killing process..."
            kill -9 \$PID
        fi
    fi
    rm -f costscope.pid
fi

# Restore backup
if [ -d "$BACKUP_DIR" ]; then
    echo " Restoring from backup: $BACKUP_DIR"
    
    if [ -f "$BACKUP_DIR/$BINARY_NAME" ]; then
        cp "$BACKUP_DIR/$BINARY_NAME" "bin/"
        echo " Binary restored"
    fi
    
    # Restore configs if needed
    # cp -r "$BACKUP_DIR/configs/" ./
    
    echo " Rollback completed"
else
    echo " Backup directory not found: $BACKUP_DIR"
    exit 1
fi
EOF

    chmod +x scripts/rollback.sh
    log_success "Rollback script created: scripts/rollback.sh"
}

# Main deployment process
main() {
    echo -e "${BLUE}Starting deployment for environment: $DEPLOYMENT_ENV${NC}"
    echo
    
    # Execute deployment steps
    check_prerequisites
    create_backup
    deploy_binary
    configure_environment
    start_monitoring
    create_health_check
    create_benchmark_script
    create_rollback_script
    validate_deployment
    
    echo
    echo -e "${GREEN} DEPLOYMENT COMPLETED SUCCESSFULLY!${NC}"
    echo "=================================="
    echo
    echo -e "${BLUE} Next Steps:${NC}"
    echo "1. Source environment: source .env.optimization"
    echo "2. Start service: ./scripts/start_monitoring.sh"
    echo "3. Health check: ./scripts/health_check.sh"
    echo "4. Run benchmarks: ./scripts/benchmark_production.sh"
    echo "5. Monitor logs: tail -f logs/costscope_*.log"
    echo
    echo -e "${BLUE} Useful URLs:${NC}"
    echo "- Profiling: http://localhost:6060/debug/pprof/"
    echo "- CPU Profile: http://localhost:6060/debug/pprof/profile"
    echo "- Memory Profile: http://localhost:6060/debug/pprof/heap"
    echo
    echo -e "${YELLOW}️  Emergency Rollback:${NC}"
    echo "./scripts/rollback.sh"
    echo
    echo -e "${GREEN} CostScope $DEPLOYMENT_ENV deployment ready!${NC}"
}

# Run main function
main "$@"
