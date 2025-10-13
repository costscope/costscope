#!/usr/bin/env bash
set -euo pipefail

# DevContainer Quality Tools Setup Script for CostScope
# Installing code quality tools inside the container

# Source common logging helpers if available
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMMON_SH="$ROOT_DIR/scripts/ci/lib/common.sh"
if [[ -f "$COMMON_SH" ]]; then
    # shellcheck disable=SC1090
    . "$COMMON_SH"
fi

# ================= Configuration =================
# Allow overriding pinned versions via env if needed (avoids surprise breakages on 'latest').
GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION:-v1.60.3}"   # Keep in sync with CI if pinned there
STATICCHECK_VERSION="${STATICCHECK_VERSION:-latest}"        # Can pin e.g. 2024.1.1
GOSEC_VERSION="${GOSEC_VERSION:-latest}"                    # securego/gosec tag (uses @latest by default)

# List of tools for final verification (tool -> human label optional)
declare -a TOOL_LIST=(
    golangci-lint
    staticcheck
    gosec
    govulncheck
    gocyclo
    gocognit
    dupl
    deadcode
    goimports
    gofumpt
    golines
    gotestsum
    go-test-coverage
    godoc
    gomarkdoc
    pprof
    benchstat
    go-mod-outdated
    go-licenses
    modgraphviz
    air
    dlv
    wire
    trivy
)

# Helper: install Go tool only if not already present (idempotent, fast for rebuilds)
ensure_go_tool() {
    local bin_name="$1"; shift
    local install_ref="$1"; shift || true
    if command -v "$bin_name" >/dev/null 2>&1; then
        print_status "Skipping $bin_name (already installed)"
        return 0
    fi
    print_status "Installing $bin_name ($install_ref)"
    if ! go install "$install_ref" 2>/tmp/${bin_name}_install.err; then
        print_warning "$bin_name install failed (see /tmp/${bin_name}_install.err)"
        return 1
    fi
}

# Allow automatic download of newer Go toolchains if needed for some tools
export GOTOOLCHAIN="auto"

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

print_status " Installing CostScope Quality Tools for DevContainer..."

# Update system packages
print_status " Updating system packages..."
# Some devcontainer images set a non-root user (e.g. vscode). We need sudo for apt operations.
APT_ELEVATE=""
if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        APT_ELEVATE="sudo"
    else
        print_warning "No sudo available and not running as root; skipping apt update/upgrade (tools relying on system packages may be missing)."
    fi
fi

if [ -n "$APT_ELEVATE" ] || [ "$(id -u)" -eq 0 ]; then
    if ! $APT_ELEVATE apt-get update -y; then
        print_warning "apt-get update failed (permission or network); continuing with Go tool installs."
    else
        # upgrade is non-essential; ignore failures
        $APT_ELEVATE apt-get upgrade -y || print_warning "apt-get upgrade failed; continuing."
    fi
fi

# Install required packages
print_status " Installing system dependencies..."
if [ -n "$APT_ELEVATE" ] || [ "$(id -u)" -eq 0 ]; then
    $APT_ELEVATE apt-get install -y curl wget git build-essential || print_warning "Failed to install some system dependencies; proceeding."
else
    print_warning "Skipping system dependency installation (no privileges). Assuming base image has required tools."
fi

# Set GOPATH and GOBIN if not set
export GOPATH=${GOPATH:-/go}
export GOBIN=${GOBIN:-/go/bin}
export PATH=$PATH:$GOBIN

# Ensure Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# 1. Core Static Analysis Tools
print_status " Installing Core Static Analysis Tools..."

# golangci-lint (comprehensive linter)
print_status "Installing golangci-lint (pinned $GOLANGCI_LINT_VERSION)..."
if command -v golangci-lint >/dev/null 2>&1; then
    INSTALLED_GCL="$(golangci-lint version 2>/dev/null | head -n1 || true)"
    print_status "golangci-lint already present: $INSTALLED_GCL"
else
    # Prefer stable release via go install (module-aware). Fall back to upstream install script if unavailable.
    if ! go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} 2>/dev/null; then
        print_warning "go install failed; falling back to upstream installer"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -o /tmp/golangci-install.sh && sh /tmp/golangci-install.sh -b "$GOBIN" "$GOLANGCI_LINT_VERSION" || print_warning "golangci-lint install script failed"
    fi
fi

# staticcheck (advanced static analysis)
print_status "Installing staticcheck ($STATICCHECK_VERSION)..."
ensure_go_tool staticcheck "honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION}"

# gosec (security analysis)
print_status "Installing gosec ($GOSEC_VERSION)..."
if ! ensure_go_tool gosec "github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}"; then
    print_warning "gosec optional; continuing (golangci-lint security linters still active)."
fi

# 2. Vulnerability and Security Tools
print_status " Installing Security Tools..."

# govulncheck (vulnerability scanner)
print_status "Installing govulncheck..."
go install golang.org/x/vuln/cmd/govulncheck@latest

# trivy (filesystem / image scanner) - try Go install first, then fall back to upstream installer
print_status "Installing trivy (filesystem/image scanner)..."
if command -v trivy >/dev/null 2>&1; then
    print_status "Skipping trivy (already installed)"
else
    # Preferred: go install if module path is available (idempotent)
    if go install github.com/aquasecurity/trivy/v4/cmd/trivy@latest 2>/tmp/trivy_go_install.err; then
        print_status "trivy installed via go install"
    else
        print_warning "go install for trivy failed; attempting upstream install script"
        # Use bash explicitly for the upstream installer (some shells don't support -s)
        if curl -sSfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh -o /tmp/trivy-install.sh; then
            if bash /tmp/trivy-install.sh -b "$GOBIN" latest 2>/tmp/trivy_install.err; then
                print_status "trivy installed via upstream installer"
            else
                print_warning "trivy upstream installer failed (see /tmp/trivy_install.err)"
            fi
        else
            print_warning "Failed to download trivy install script"
        fi
    fi
fi

# 3. Code Quality and Complexity Tools
print_status " Installing Code Quality Tools..."

# gocyclo (cyclomatic complexity)
ensure_go_tool gocyclo github.com/fzipp/gocyclo/cmd/gocyclo@latest

# gocognit (cognitive complexity)
print_status "Installing gocognit..."
ensure_go_tool gocognit github.com/uudashr/gocognit/cmd/gocognit@latest

# dupl (duplicate code detection)
ensure_go_tool dupl github.com/mibk/dupl@latest

# deadcode (dead code detection)
ensure_go_tool deadcode golang.org/x/tools/cmd/deadcode@latest

# 4. Formatting and Import Tools
print_status " Installing Formatting Tools..."

# goimports (import management)
ensure_go_tool goimports golang.org/x/tools/cmd/goimports@latest

# gofumpt (stricter formatting)
ensure_go_tool gofumpt mvdan.cc/gofumpt@latest

# golines (line length formatter)
ensure_go_tool golines github.com/segmentio/golines@latest

# 5. Testing Tools
print_status " Installing Testing Tools..."

# gotestsum (enhanced test runner)
ensure_go_tool gotestsum gotest.tools/gotestsum@latest

# go-test-coverage (coverage analysis)
print_status "Installing go-test-coverage..."
ensure_go_tool go-test-coverage github.com/vladopajic/go-test-coverage/v2@latest

# 6. Documentation Tools
print_status " Installing Documentation Tools..."

# godoc (documentation)
ensure_go_tool godoc golang.org/x/tools/cmd/godoc@latest

# gomarkdoc (markdown docs)
ensure_go_tool gomarkdoc github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest

# 7. Performance Tools
print_status " Installing Performance Tools..."

# pprof (profiling)
ensure_go_tool pprof github.com/google/pprof@latest

# benchstat (benchmark analysis)
ensure_go_tool benchstat golang.org/x/perf/cmd/benchstat@latest

# 8. Dependency Tools
print_status " Installing Dependency Tools..."

# go-mod-outdated (outdated dependencies)
ensure_go_tool go-mod-outdated github.com/psampaz/go-mod-outdated@latest

# go-licenses (license checking)
ensure_go_tool go-licenses github.com/google/go-licenses@latest

# modgraphviz (dependency graph)
ensure_go_tool modgraphviz golang.org/x/exp/cmd/modgraphviz@latest

# 9. Additional Linux Tools
print_status " Installing Additional Linux Tools..."

# jq (JSON processing)
print_status "Installing jq..."
if [ -n "$APT_ELEVATE" ] || [ "$(id -u)" -eq 0 ]; then
    $APT_ELEVATE apt-get install -y jq || print_warning "jq install failed; attempting to continue."
else
    print_warning "Skipping jq install (no privileges)."
fi

# yq (YAML processing)
print_status "Installing yq..."
if [ -n "$APT_ELEVATE" ] || [ "$(id -u)" -eq 0 ]; then
    if wget -q -O /tmp/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64; then
        $APT_ELEVATE mv /tmp/yq /usr/local/bin/yq || mv /tmp/yq "$GOBIN/yq" || print_warning "Unable to place yq in /usr/local/bin; installed under GOPATH instead."
        if [ -x /usr/local/bin/yq ]; then
            chmod +x /usr/local/bin/yq || true
        fi
    else
        print_warning "Failed to download yq; skipping."
    fi
else
    print_warning "Skipping yq install (no privileges)."
fi

# tree (directory visualization)
print_status "Installing tree..."
if [ -n "$APT_ELEVATE" ] || [ "$(id -u)" -eq 0 ]; then
    $APT_ELEVATE apt-get install -y tree || print_warning "tree install failed; continuing."
else
    print_warning "Skipping tree install (no privileges)."
fi

# 10. Development Tools
print_status " Installing Development Tools..."

# air (live reload for Go applications)
ensure_go_tool air github.com/air-verse/air@latest

# delve (Go debugger)
ensure_go_tool dlv github.com/go-delve/delve/cmd/dlv@latest

# wire (dependency injection)
ensure_go_tool wire github.com/google/wire/cmd/wire@latest

print_success " All quality tools installed successfully!"

# Verify installations
print_status " Verifying installations..."

tools=(
    "golangci-lint"
    "staticcheck" 
    "gosec"
    "trivy"
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

for tool in "${tools[@]}"; do
    if command -v "$tool" &> /dev/null; then
        print_success "$tool "
    else
        print_warning "$tool  (not found in PATH)"
    fi
done

print_status "Generating summary..."
missing=()
for t in "${TOOL_LIST[@]}"; do
    if ! command -v "$t" >/dev/null 2>&1; then
        missing+=("$t")
    fi
done

if [ ${#missing[@]} -eq 0 ]; then
    print_success "All ${#TOOL_LIST[@]} tools available."
else
    print_warning "Missing ${#missing[@]} tool(s): ${missing[*]}"
    echo "You can retry installs for missing tools, e.g.:"
    for m in "${missing[@]}"; do
        case "$m" in
            golangci-lint) echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}  # preferred (or use the official install script if necessary)";;
            *) echo "  go install <module path for $m>@latest";;
        esac
    done
fi

print_success " DevContainer quality tools setup completed!"
print_status " Next steps:"
echo "  1. make quality"
echo "  2. make setup-hooks   # if defined"
echo "  3. make dev           # start live dev"
echo "  4. make help          # discover commands"
