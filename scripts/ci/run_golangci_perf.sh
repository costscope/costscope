#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Runs golangci-lint for perf-relevant packages with duckdb build tag

GOLANGCI_BIN="$(command -v golangci-lint || true)"
if [[ -z "$GOLANGCI_BIN" ]]; then
  # fallback to GOPATH/bin
  GOLANGCI_BIN="$(go env GOPATH)/bin/golangci-lint"
fi

if [[ ! -x "$GOLANGCI_BIN" ]]; then
  ci::warn "golangci-lint not found; skipping perf lint."
  exit 0
fi

TARGET_PATTERNS=(
  ./internal/database/...
  ./scripts/tools/perf-bench
  ./scripts/tools/perf-gen-synth
  ./internal/core/focus/...
  ./internal/core/logging
)

mapfile -t RESOLVED_DIRS < <(go list -f '{{.Dir}}' "${TARGET_PATTERNS[@]}" 2>/dev/null || true)
if [[ ${#RESOLVED_DIRS[@]} -eq 0 ]]; then
  ci::log "No packages to lint in perf scope; skipping golangci-lint run."
  exit 0
fi

if ci::is_github_actions; then
  ci::log "Running golangci-lint (GitHub format) on ${#RESOLVED_DIRS[@]} dirs"
  "$GOLANGCI_BIN" run --config .golangci.perf.yml --verbose --out-format=github-actions "${RESOLVED_DIRS[@]}"
else
  ci::log "Running golangci-lint on ${#RESOLVED_DIRS[@]} dirs"
  "$GOLANGCI_BIN" run --config .golangci.perf.yml --verbose "${RESOLVED_DIRS[@]}"
fi
