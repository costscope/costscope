#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Usage: build-test-matrix.sh <variant>
# variant: slim | sqlite | duckdb

VARIANT=${1:-}
if [[ -z "${VARIANT}" ]]; then
  ci::die "usage: $0 <variant: slim|sqlite|duckdb>"
fi

ci::log "build-test-matrix variant=${VARIANT} (IS_ACT=${IS_ACT:-false})"

go mod download

# Use sudo only if available to avoid noisy warnings in containerized environments
if command -v sudo >/dev/null 2>&1; then SUDO=sudo; else SUDO=; fi

## Act-specific skipping/reduction removed — run full tests under act as requested

case "${VARIANT}" in
  slim)
    if ci::is_act; then
      # Under act, run all packages without -race to keep memory stable.
      # Avoid invoking the make target (which uses -race) to reduce OOM risk
      # and eliminate any edge cases with set -e and fallback semantics.
      env -u GOROOT GOTOOLCHAIN=auto go test -v -cover ./...
    else
      make test-slim || go test -race ./...
    fi
    ;;
  sqlite)
    if ci::is_act; then
      make test-sqlite || env -u GOROOT GOTOOLCHAIN=auto go test -v -cover -tags sqlite ./...
    else
      make test-sqlite || go test -race -tags sqlite ./...
    fi
    ;;
  duckdb)
    if ci::is_act; then
      make test-duckdb || env -u GOROOT GOTOOLCHAIN=auto go test -v -cover -tags duckdb ./...
    else
      make test-duckdb || go test -race -tags duckdb ./...
    fi
    ;;
  *)
    ci::die "unknown variant: ${VARIANT}"
    ;;
esac
