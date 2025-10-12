#!/bin/bash

# CostScope DevContainer Setup Script
# This script sets up the development environment with all necessary tools

set -euo pipefail

# Pin exact Go toolchain (override base image minor) to avoid mixed stdlib objects.
# Default matches go.mod (go 1.24.6). Can be overridden via COSTSCOPE_GO_VERSION env (set in devcontainer.json).
GO_VERSION="${COSTSCOPE_GO_VERSION:-1.24.6}"
export GOTOOLCHAIN="local"
echo "[setup] Ensuring Go ${GO_VERSION} toolchain is installed (GOTOOLCHAIN=${GOTOOLCHAIN})"
if ! go version 2>/dev/null | grep -q "go${GO_VERSION}"; then
    echo "[setup] Replacing existing Go with ${GO_VERSION}"
    sudo rm -rf /usr/local/go && \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz" -o /tmp/go.tgz && \
    sudo tar -C /usr/local -xzf /tmp/go.tgz && rm /tmp/go.tgz
fi
go version

echo " Setting up CostScope development environment..."

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

# Update system packages
print_status "Updating system packages..."
sudo apt-get update -qq

# Install additional system dependencies
print_status "Installing system dependencies..."
sudo apt-get install -y \
    curl \
    wget \
    git \
    make \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    jq \
    unzip \
    yamllint \
    tree \
    htop \
    vim

# Install Go tools for development
print_status "Installing Go development tools..."

# Ensure GOPATH/bin on PATH for current session and future shells
GOBIN="$(go env GOPATH)/bin"
export PATH="$GOBIN:$PATH"
if ! grep -q "\$GOPATH/bin" ~/.bashrc 2>/dev/null; then
    echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
fi

# golangci-lint for static analysis — require v2 to match .golangci.yml (version: "2")
print_status "Installing golangci-lint (v2)..."
# Detect current installed major version if any
current_major=""
if command -v golangci-lint >/dev/null 2>&1; then
    # Typical output contains: "golangci-lint has version 1.54.2" or "golangci-lint has version 2.0.0"
    raw_ver=$(golangci-lint version 2>/dev/null || true)
    current_major=$(echo "$raw_ver" | grep -Eo 'version [0-9]+' | awk '{print $2}' || true)
fi
if [ "${current_major:-}" != "2" ]; then
    print_status "Installing/Upgrading golangci-lint to v2.x (current major: ${current_major:-none})"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -o /tmp/golangci-install.sh
    # Prefer an explicit v2 tag first, fall back to latest if the tag string ever changes upstream
    if ! sudo sh /tmp/golangci-install.sh -b /usr/local/bin v2.0.0 >/dev/null 2>&1; then
        sudo sh /tmp/golangci-install.sh -b /usr/local/bin latest
    fi
else
    print_status "golangci-lint v2 already present"
fi
golangci-lint version || true

# govulncheck for vulnerability scanning
print_status "Installing govulncheck..."
go install golang.org/x/vuln/cmd/govulncheck@latest

# gosec for security analysis (included in golangci-lint)
print_status "gosec is included in golangci-lint - no separate installation needed"

# gocyclo for cyclomatic complexity
print_status "Installing gocyclo..."
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# staticcheck for additional static analysis
print_status "Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@latest

# ineffassign for finding ineffective assignments
print_status "Installing ineffassign..."
go install github.com/gordonklaus/ineffassign@latest

# misspell for spell checking
print_status "Installing misspell..."
go install github.com/client9/misspell/cmd/misspell@latest

# goimports for import management
print_status "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest

# go-mod-outdated for dependency checking
print_status "Installing go-mod-outdated..."
go install github.com/psampaz/go-mod-outdated@latest

# deadcode detector
print_status "Installing deadcode..."
go install golang.org/x/tools/cmd/deadcode@latest

# Air for live reloading during development (requires Go >= 1.25). Optional under older Go.
print_status "Installing Air..."
if ! go install github.com/air-verse/air@latest >/dev/null 2>&1; then
    print_warning "Air install failed (likely requires Go >= 1.25); skipping under Go ${GO_VERSION}"
fi

# Delve debugger
print_status "Installing Delve debugger..."
go install github.com/go-delve/delve/cmd/dlv@latest

# Install testing tools
print_status "Installing testing tools..."

# testify for better testing (library - included in go.mod, not CLI tool)

# go-test-coverage for coverage reporting
go install github.com/vladopajic/go-test-coverage/v2@latest

# Install additional utilities
print_status "Installing additional utilities..."

# httpie for API testing
sudo apt-get install -y httpie

# Setup git hooks directory
print_status "Configuring git hooks..."
# Prefer versioned hooks via core.hooksPath -> .githooks (contains delegating pre-commit)
if ! git config core.hooksPath >/dev/null 2>&1; then
    git config core.hooksPath .githooks || true
fi

configured_path=$(git config core.hooksPath 2>/dev/null || echo "")
if [ "$configured_path" != ".githooks" ]; then
    print_status "core.hooksPath currently '$configured_path' (expected .githooks); setting now"
    git config core.hooksPath .githooks || true
fi

# Ensure the versioned hook is executable
chmod +x .githooks/pre-commit 2>/dev/null || true

# Fallback: if .githooks/pre-commit missing (should not happen), install a strict minimal hook that fails on lint errors
if [ ! -f .githooks/pre-commit ]; then
    print_status "Versioned .githooks/pre-commit missing; installing strict fallback into .git/hooks"
    mkdir -p .git/hooks
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -euo pipefail
echo "(fallback) Running pre-commit checks..."
golangci-lint run || { echo " golangci-lint failed" >&2; exit 1; }
go test ./...   || { echo " tests failed" >&2; exit 1; }
echo " fallback pre-commit checks passed"
EOF
    chmod +x .git/hooks/pre-commit || true
fi

# Create quality check script
print_status "Creating quality check script..."
cat > scripts/quality-check.sh << 'EOF'
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
EOF

mkdir -p scripts
chmod +x scripts/quality-check.sh

# Setup workspace
print_status "Setting up workspace..."

# Install Go dependencies
print_status "Installing Go module dependencies..."
go mod download
go mod tidy

# Build the project to ensure everything works
print_status "Building project (initial attempt)..."
if ! go build .; then
    print_warning "Initial build failed – attempting with auto toolchain (go version mismatch likely)."
    if ! go build .; then
        print_error "Build still failing after auto toolchain attempt. Check go.mod version vs installed toolchain."
        exit 1
    fi
fi

# Create useful aliases
print_status "Setting up useful aliases..."
cat >> ~/.bashrc << 'EOF'

# CostScope development aliases
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
alias ..='cd ..'
alias ...='cd ../..'
alias grep='grep --color=auto'

# Go development aliases
alias gotest='go test -v -race -cover'
alias gobuild='go build -race'
alias golint='golangci-lint run'
alias gomod='go mod tidy && go mod download'
alias gofmt='goimports -w .'

# CostScope specific aliases
alias quality='./scripts/quality-check.sh'
alias duplicates='make duplicates'
alias costscope-build='go build -o costscope .'
alias costscope-test='go test ./... -v -race -cover'

EOF

print_success "Development environment setup completed!"
print_status "Available tools:"
echo "   golangci-lint - Static analysis (includes gosec)"
echo "   govulncheck - Vulnerability scanning"
echo "   gocyclo - Cyclomatic complexity"
echo "   staticcheck - Additional static analysis"
echo "   Air - Live reloading"
echo "   Delve - Debugger"
echo "   Pre-commit hooks"
echo "   yamllint - YAML linter for workflow validation"

# Install quality tools for DevContainer
print_status " Installing quality tools for DevContainer..."
bash /workspaces/costscope/scripts/install-quality-tools-devcontainer.sh

print_status "Useful commands:"
echo "   make quality - Run all quality checks"
echo "   make duplicates - Check for duplicates"
echo "   make test - Run tests"
echo "   make build - Build the project"
echo "   make dev - Start development server with live reload"

print_success " Ready to develop CostScope!"

# Ensure the container hostname 'mic' resolves to avoid noisy "sudo: unable to resolve host mic" warnings
print_status "Ensuring /etc/hosts contains mapping for 'mic'..."
if sudo grep -qw "mic" /etc/hosts; then
    print_status "/etc/hosts already contains 'mic' mapping"
else
    echo "127.0.0.1 mic" | sudo tee -a /etc/hosts >/dev/null
    print_success "Added '127.0.0.1 mic' to /etc/hosts"
fi
