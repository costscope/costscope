#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

WANT_VERSION="${WANT_VERSION:?WANT_VERSION required}"
GOLANGCI_HOME="${GOLANGCI_HOME:-${GITHUB_WORKSPACE:-$PWD}/.cache/tools/golangci-lint}"
RUN_OS="${RUNNER_OS:-linux}"
RUN_ARCH="${RUNNER_ARCH:-amd64}"
GOBIN_DIR="${GOBIN:-${GOLANGCI_HOME}/${RUN_OS}-${RUN_ARCH}/${WANT_VERSION}}"

ci::require_cmd go
mkdir -p "$GOBIN_DIR"

# 1) Prefer a system-provided golangci-lint (e.g., baked into CI image) if version matches
if command -v golangci-lint >/dev/null 2>&1; then
  if golangci-lint version 2>/dev/null | grep -q "${WANT_VERSION}\b"; then
    ci::log "golangci-lint ${WANT_VERSION} found on PATH; reusing preinstalled binary"
    # Ensure it's available to subsequent steps even if scripts expect it in GOBIN_DIR
    ln -sf "$(command -v golangci-lint)" "$GOBIN_DIR/golangci-lint"
  else
    ci::log "golangci-lint on PATH does not match ${WANT_VERSION}; will install requested version"
  fi
fi

# 2) Use cached tool location if already built for this WANT_VERSION
if [[ -x "$GOBIN_DIR/golangci-lint" ]] && "$GOBIN_DIR/golangci-lint" version 2>/dev/null | grep -q "${WANT_VERSION}\b"; then
  ci::log "golangci-lint ${WANT_VERSION} already present in cache; skipping install"
else
  ci::log "Installing golangci-lint ${WANT_VERSION} via go install"
  GO111MODULE=on GOBIN="$GOBIN_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@"${WANT_VERSION}"
fi

ci::log "GOBIN=$GOBIN_DIR"
if [[ -n "${GITHUB_PATH:-}" ]]; then
  echo "$GOBIN_DIR" >> "$GITHUB_PATH"
else
  # GitHub is deprecating ::add-path::, but keep fallback for local environments
  echo "::add-path::$GOBIN_DIR"
fi
