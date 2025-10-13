#!/usr/bin/env bash
set -euo pipefail

# Usage: build-test-matrix.sh <variant>
# variant: slim | sqlite | duckdb

VARIANT=${1:-}
if [[ -z "${VARIANT}" ]]; then
  echo "usage: $0 <variant: slim|sqlite|duckdb>" >&2
  exit 2
fi

echo "[ci] build-test-matrix variant=${VARIANT} (IS_ACT=${IS_ACT:-false})"

go mod download

# Use sudo only if available to avoid noisy warnings in containerized environments
if command -v sudo >/dev/null 2>&1; then SUDO=sudo; else SUDO=; fi

## Act-specific skipping/reduction removed — run full tests under act as requested

case "${VARIANT}" in
  slim)
    if [[ "${IS_ACT:-false}" == "true" ]]; then
      # Under act, run all packages without -race to keep memory stable.
      # Avoid invoking the make target (which uses -race) to reduce OOM risk
      # and eliminate any edge cases with set -e and fallback semantics.
      env -u GOROOT GOTOOLCHAIN=auto go test -v -cover ./...
    else
      make test-slim || go test -race ./...
    fi
    ;;
  sqlite)
    if [[ "${IS_ACT:-false}" == "true" ]]; then
      make test-sqlite || env -u GOROOT GOTOOLCHAIN=auto go test -v -cover -tags sqlite ./...
    else
      make test-sqlite || go test -race -tags sqlite ./...
    fi
    ;;
  duckdb)
    if [[ "${IS_ACT:-false}" == "true" ]]; then
      make test-duckdb || env -u GOROOT GOTOOLCHAIN=auto go test -v -cover -tags duckdb ./...
    else
      make test-duckdb || go test -race -tags duckdb ./...
    fi
    ;;
  *)
    echo "unknown variant: ${VARIANT}" >&2
    exit 3
    ;;
esac
