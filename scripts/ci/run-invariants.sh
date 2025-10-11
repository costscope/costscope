#!/usr/bin/env bash
set -euo pipefail

mkdir -p invariants-artifacts
./bin/invariants-ci -config configs/invariants.example.json | tee invariants-artifacts/summary.json

if command -v jq >/dev/null 2>&1; then
  if jq -e '.passed==true' invariants-artifacts/summary.json >/dev/null; then
    echo "Invariants passed"
  else
    echo "Invariants failed" >&2
    exit 1
  fi
fi
