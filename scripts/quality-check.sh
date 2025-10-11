#!/bin/bash

# Quality check script for CostScope
set -euo pipefail

echo " Running comprehensive quality checks..."

# 1. Check for duplicates
echo " Step 1: Checking for duplicates (dupl)..."
make duplicates

# 2. Static analysis
echo " Step 2: Running static analysis..."
golangci-lint run --timeout=10m

# 3. Security scan (gosec is included in golangci-lint above)
echo " Step 3: Security scan completed via golangci-lint"

# 4. Vulnerability check
echo " Step 4: Checking for vulnerabilities..."
govulncheck ./...

# 5. Tests
echo " Step 5: Running tests..."
go test -race -cover ./...

# 6. Build check
echo " Step 6: Checking build..."
go build ./...

echo " All quality checks passed!"
