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

if [[ -x "$GOBIN_DIR/golangci-lint" ]] && "$GOBIN_DIR/golangci-lint" version 2>/dev/null | grep -q "${WANT_VERSION}\b"; then
  ci::log "golangci-lint ${WANT_VERSION} already present; skipping install"
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
