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

# Under act reduce Go parallelism & test worker fanout to mitigate OOM (exit 137) events.
if ci::is_act; then
  export GOMAXPROCS=2
  export GOFLAGS="-p=2"
  ci::log "[act] Applied resource throttling: GOMAXPROCS=$GOMAXPROCS GOFLAGS=$GOFLAGS"
fi

# Ensure required smoke fixtures exist before running tests (needed by focus conversion parity tests)
if [[ ! -f tests/fixtures/aws/cur_smoke.csv || ! -f tests/fixtures/azure/usage_smoke.csv || ! -f tests/fixtures/gcp/usage_smoke.csv ]]; then
  if [[ -f ./scripts/ci/prepare-act-fixtures.sh ]]; then
    ci::log "[ci] Smoke fixtures missing; generating"
    bash ./scripts/ci/prepare-act-fixtures.sh || ci::warn "fixture generation failed (continuing)"
  fi
fi

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
      set +e
      env -u GOROOT GOTOOLCHAIN=auto go test -v -cover ./...
      status=$?
      set -e
      if ci::is_act; then
        # Re-run in lightweight JSON mode only if failure to surface failing packages/tests for local diagnostics
        if [[ $status -ne 0 ]]; then
          ci::warn "[act] test run failed; collecting JSON diagnostics (retry without coverage/race)"
          # Best-effort: do not fail if jq absent
          if command -v jq >/dev/null 2>&1; then
            env -u GOROOT GOTOOLCHAIN=auto go test -json ./... 2>/dev/null | tee act-test-results.json | jq -r 'select(.Action=="fail" and .Test!="") | "FAIL TEST: \(.Package) :: \(.Test)"' > act-test-failures.txt || true
            ci::log "[act] failing tests (if any):"; (grep '^FAIL TEST:' act-test-failures.txt || echo '[none extracted]')
          else
            ci::warn "[act] jq not installed; skipping JSON failure extraction"
          fi
        fi
      fi
      exit $status
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
