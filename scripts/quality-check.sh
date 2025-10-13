#!/usr/bin/env bash

# Quality check script for CostScope
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::log " Running comprehensive quality checks..."

# 1. Check for duplicates
ci::log " Step 1: Checking for duplicates (dupl)..."
make duplicates

# 2. Static analysis
ci::log " Step 2: Running static analysis..."
golangci-lint run --timeout=10m

# 3. Security scan (gosec is included in golangci-lint above)
ci::log " Step 3: Security scan completed via golangci-lint"

# 4. Vulnerability check
ci::log " Step 4: Checking for vulnerabilities..."
govulncheck ./...

# 5. Tests
ci::log " Step 5: Running tests..."
go test -race -cover ./...

# 6. Build check
ci::log " Step 6: Checking build..."
go build ./...

ci::log " All quality checks passed!"
