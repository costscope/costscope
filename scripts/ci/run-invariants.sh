#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

mkdir -p invariants-artifacts
ci::log "Running invariants-ci"
./bin/invariants-ci -config configs/invariants.example.json | tee invariants-artifacts/summary.json

if command -v jq >/dev/null 2>&1; then
  if jq -e '.passed==true' invariants-artifacts/summary.json >/dev/null; then
    ci::log "Invariants passed"
  else
    ci::die "Invariants failed"
  fi
else
  ci::warn "jq not found; skipping pass/fail gate"
fi
