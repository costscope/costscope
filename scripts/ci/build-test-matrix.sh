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

# Under act, skip non-slim variants to avoid apt/network instability
if [[ "${IS_ACT:-false}" == "true" && "${VARIANT}" != "slim" && "${ACT_FULL:-false}" != "true" ]]; then
  echo "[act] Skipping ${VARIANT} variant under act"
  exit 0
fi

# Under act for the slim variant, run a reduced test set that excludes packages requiring large external fixtures
if [[ "${IS_ACT:-false}" == "true" && "${VARIANT}" == "slim" && "${ACT_FULL:-false}" != "true" ]]; then
  echo "[act] Running reduced slim test set (excluding internal/core/focus/conversion)"
  # Tolerate no matches without failing the script under 'set -o pipefail'
  pkgs=$(go list ./... | grep -Ev '^local/costscope/internal/core/focus/conversion($|/)' || true)
  if [[ -n "${pkgs}" ]]; then
    echo "[act] Package count: $(echo "$pkgs" | wc -w | tr -d ' ')"
    failed=0
    # Run package-by-package to surface the exact failing package under act
    for pkg in ${pkgs}; do
      echo "[act] >>> go test -v -race -cover ${pkg}"
      if ! env -u GOROOT GOTOOLCHAIN=auto go test -v -race -cover "${pkg}"; then
        echo "[act] FAIL: ${pkg}"
        failed=1
      fi
    done
    if [[ ${failed} -ne 0 ]]; then
      echo "[act] Reduced test run had failures"
      exit 1
    fi
  else
    echo "[act] No packages selected for reduced test run"
  fi
  exit 0
fi

case "${VARIANT}" in
  slim)
    make test-slim || go test ./...
    ;;
  sqlite)
    ${SUDO} apt-get update || true
    ${SUDO} apt-get install -y build-essential || true
    make test-sqlite || go test -tags sqlite ./...
    ;;
  duckdb)
    ${SUDO} apt-get update || true
    ${SUDO} apt-get install -y build-essential || true
    make test-duckdb || go test -tags duckdb ./...
    ;;
  *)
    echo "unknown variant: ${VARIANT}" >&2
    exit 3
    ;;
esac
