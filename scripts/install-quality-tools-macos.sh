#!/bin/bash

# CostScope - Quality Tools Installation Script for macOS
# This script installs comprehensive code quality tools

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

print_status " Installing CostScope Quality Tools for macOS..."

# 1. Core Static Analysis Tools
print_status " Installing Core Static Analysis Tools..."

# golangci-lint (comprehensive linter aggregator)
if ! command_exists golangci-lint; then
    print_status "Installing golangci-lint..."
    brew install golangci-lint
else
    print_success "golangci-lint already installed"
fi

# staticcheck (advanced static analysis)
print_status "Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@latest

# gosec (security analysis)
print_status "Installing gosec..."
brew install gosec

# 2. Vulnerability and Security Tools
print_status " Installing Security Tools..."

# govulncheck (vulnerability scanner)
print_status "Installing govulncheck..."
go install golang.org/x/vuln/cmd/govulncheck@latest

# nancy (vulnerability scanner for dependencies) - repository currently unavailable
# if ! command_exists nancy; then
#     print_status "Installing nancy..."
#     go install github.com/sonatypecommunity/nancy@latest
# fi

# 3. Code Quality and Complexity Tools
print_status " Installing Code Quality Tools..."

# gocyclo (cyclomatic complexity)
print_status "Installing gocyclo..."
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# gocognit (cognitive complexity)
print_status "Installing gocognit..."
go install github.com/uudashr/gocognit/cmd/gocognit@latest

# dupl (duplicate code detection)
print_status "Installing dupl..."
go install github.com/mibk/dupl@latest

# deadcode (dead code detection)
print_status "Installing deadcode..."
go install golang.org/x/tools/cmd/deadcode@latest

# 4. Code Formatting and Import Tools
print_status " Installing Formatting Tools..."

# goimports (import management)
print_status "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest

# gofumpt (stricter gofmt)
print_status "Installing gofumpt..."
go install mvdan.cc/gofumpt@latest

# golines (line length formatter)
print_status "Installing golines..."
go install github.com/segmentio/golines@latest

# 5. Testing Tools
print_status " Installing Testing Tools..."

# gotestsum (better test output)
print_status "Installing gotestsum..."
go install gotest.tools/gotestsum@latest

# go-test-coverage (coverage analysis)
print_status "Installing go-test-coverage..."
go install github.com/vladopajic/go-test-coverage/v2@latest

# testify (testing framework) - already in go.mod
# print_status "Installing testify..."
# go install github.com/stretchr/testify@latest

# 6. Documentation Tools
print_status " Installing Documentation Tools..."

# godoc (documentation)
print_status "Installing godoc..."
go install golang.org/x/tools/cmd/godoc@latest

# gomarkdoc (markdown documentation generator)
print_status "Installing gomarkdoc..."
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest

# 7. Performance and Profiling Tools
print_status " Installing Performance Tools..."

# pprof (profiling)
print_status "Installing pprof..."
go install github.com/google/pprof@latest

# benchstat (benchmark analysis)
print_status "Installing benchstat..."
go install golang.org/x/perf/cmd/benchstat@latest

# 8. Dependency Management Tools
print_status " Installing Dependency Tools..."

# go-mod-outdated (outdated dependency detection)
print_status "Installing go-mod-outdated..."
go install github.com/psampaz/go-mod-outdated@latest

# go-licenses (license checker)
print_status "Installing go-licenses..."
go install github.com/google/go-licenses@latest

# modgraphviz (dependency graph visualization)
print_status "Installing modgraphviz..."
go install golang.org/x/exp/cmd/modgraphviz@latest

# 9. Additional Homebrew Tools
print_status " Installing Additional macOS Tools..."

# jq for JSON processing
if ! command_exists jq; then
    print_status "Installing jq..."
    brew install jq
fi

# yq for YAML processing
if ! command_exists yq; then
    print_status "Installing yq..."
    brew install yq
fi

# tree for directory visualization
if ! command_exists tree; then
    print_status "Installing tree..."
    brew install tree
fi

# htop for system monitoring
if ! command_exists htop; then
    print_status "Installing htop..."
    brew install htop
fi

# 10. Development Tools
print_status " Installing Development Tools..."

# # air (live reload for Go applications)
print_status "Installing air..."
go install github.com/air-verse/air@latest

# delve (debugger)
print_status "Installing delve..."
go install github.com/go-delve/delve/cmd/dlv@latest

# wire (dependency injection)
print_status "Installing wire..."
go install github.com/google/wire/cmd/wire@latest

print_success " All quality tools installed successfully!"

# Verify installations
print_status " Verifying installations..."

tools=(
    "golangci-lint"
    "staticcheck" 
    "gosec"
    "govulncheck"
    "gocyclo"
    "gocognit"
    "dupl"
    "deadcode"
    "goimports"
    "gofumpt"
    "gotestsum"
    "air"
    "dlv"
)

failed_tools=()

for tool in "${tools[@]}"; do
    if command_exists "$tool"; then
        print_success "$tool "
    else
        print_error "$tool "
        failed_tools+=("$tool")
    fi
done

if [ ${#failed_tools[@]} -eq 0 ]; then
    print_success " All tools installed and verified successfully!"
else
    print_warning "️ Some tools failed to install: ${failed_tools[*]}"
fi

print_status " Next steps:"
echo "  1. Run 'make quality' to check code quality"
echo "  2. Run 'make setup-hooks' to install git hooks"
echo "  3. Use 'make dev' for live development"
echo "  4. Check available commands with 'make help'"
