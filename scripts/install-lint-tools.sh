#!/usr/bin/env bash
set -euo pipefail

# Install pinned developer lint/security tools.
# Respects WANT_VERSION environment variable for golangci-lint (default v1.54.2).

WANTED="${WANT_VERSION:-v1.54.2}"
GOBIN="$(go env GOPATH)/bin"

mkdir -p "$GOBIN"

echo "Installing pinned lint/security tools"
echo " - golangci-lint: $WANTED"
echo " - staticcheck: latest"
echo " - gosec: latest"
echo " - gocyclo: latest"

# Build/install with the local Go toolchain so export-data stays compatible

install_or_warn() {
	local cmd="$*"
	echo "Running: $cmd"
	if ! env GO111MODULE=on GOBIN="$GOBIN" bash -lc "$cmd"; then
		echo "\nWARNING: failed to install via: $cmd"
		return 1
	fi
	return 0
}

# Try to install golangci-lint via 'go install'. If that fails, download a prebuilt binary
try_install_golangci() {
	local ver="$1"
	if install_or_warn "go install \"github.com/golangci/golangci-lint/cmd/golangci-lint@${ver}\""; then
		return 0
	fi

	echo "Attempting to download prebuilt golangci-lint ${ver}..."
	local rel_ver
	rel_ver="${ver#v}"
	local os_name
	local arch
	os_name="$(uname -s | tr '[:upper:]' '[:lower:]')"
	arch="$(uname -m)"
	case "$arch" in
		x86_64|amd64) arch="amd64" ;; 
		aarch64|arm64) arch="arm64" ;; 
		*) echo "Unsupported arch: $arch"; return 1 ;;
	esac
	case "$os_name" in
		linux|darwin) ;;
		*) echo "Unsupported OS: $os_name"; return 1 ;;
	esac

	local filename="golangci-lint-${rel_ver}-${os_name}-${arch}.tar.gz"
	local url="https://github.com/golangci/golangci-lint/releases/download/${ver}/${filename}"
	local tmpd
	tmpd="$(mktemp -d)"
	trap 'rm -rf "$tmpd"' RETURN

	echo "Downloading: $url"
	if ! curl -fsSL -o "$tmpd/$filename" "$url"; then
		echo "Failed to download $url"
		return 1
	fi

	echo "Extracting..."
	if ! tar -xzf "$tmpd/$filename" -C "$tmpd"; then
		echo "Failed to extract $filename"
		return 1
	fi

	local bin_dir="$tmpd/golangci-lint-${rel_ver}-${os_name}-${arch}"
	if [ ! -x "$bin_dir/golangci-lint" ]; then
		echo "Downloaded archive does not contain golangci-lint binary in expected path: $bin_dir/golangci-lint"
		return 1
	fi
	cp "$bin_dir/golangci-lint" "$GOBIN/"
	chmod +x "$GOBIN/golangci-lint"
	echo "Installed golangci-lint to $GOBIN/golangci-lint"
}

try_install_golangci "$WANTED"

# Install the remaining tools; failures are non-fatal but warned
install_or_warn "go install \"honnef.co/go/tools/cmd/staticcheck@latest\"" || true
install_or_warn "go install \"github.com/securego/gosec/v2/cmd/gosec@latest\"" || true
install_or_warn "go install \"github.com/fzipp/gocyclo/cmd/gocyclo@latest\"" || true

echo "Installed tools to $GOBIN; ensure $GOBIN is on your PATH"
