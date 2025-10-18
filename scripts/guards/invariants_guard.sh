#!/usr/bin/env bash
set -euo pipefail
# invariants_guard.sh - Run invariants drift guard comparing current regenerated invariants vs baseline.
# Exit codes:
#  0 success
#  2 conversion/regenerate failure
#  3 invariants drift detected

INVARIANTS_TOLERANCE=${INVARIANTS_TOLERANCE:-0.01}
BASELINE=${INVARIANTS_BASELINE:-tests/fixtures/quality/baseline_invariants.json}
INPUT=${INPUT:-tests/perf/aws-cur-synth.csv.gz}
DBG=bin/costscope-duckdb-debug
OPT=bin/costscope-optimized-duckdb
BIN=$DBG
PARQUET_ROTATE_SIZE=${PARQUET_ROTATE_SIZE:-10000000000}

if [[ ! -f "$BASELINE" ]]; then
  echo " Baseline not found: $BASELINE" >&2
  exit 1
fi
if [[ ! -f "$INPUT" ]]; then
  echo " Input dataset missing: $INPUT" >&2
  exit 1
fi
if [[ ! -x "$DBG" || ! -x "$OPT" ]]; then
  echo " DuckDB binaries missing (expected $DBG and $OPT)" >&2
  exit 1
fi

echo "[invariants-guard] Converting fast path (preferred debug=$BIN)"
if ! $BIN convert --provider aws --input "$INPUT" --output focus_fast.parquet --streaming --rotate-size "$PARQUET_ROTATE_SIZE" >/dev/null 2>&1; then
  echo "️  Debug convert failed; trying optimized binary"
  BIN=$OPT
  if ! $BIN convert --provider aws --input "$INPUT" --output focus_fast.parquet --streaming --rotate-size "$PARQUET_ROTATE_SIZE" >/dev/null 2>&1; then
    echo " Conversion failed with both binaries" >&2
    exit 2
  fi
fi

LATEST=$(ls -1t focus_fast*.parquet 2>/dev/null | head -n1)
if [[ -z "$LATEST" ]]; then
  echo " No parquet output" >&2
  exit 2
fi

echo "[invariants-guard] Regenerating current invariants from $LATEST"
if ! $BIN invariants regenerate "$LATEST" --output invariants_current.json --force --tolerance "${INVARIANTS_TOLERANCE}" >/dev/null 2>&1; then
  echo "️  Regenerate failed with $BIN; trying fallback binary"
  ALT=$([[ $BIN == $DBG ]] && echo $OPT || echo $DBG)
  if ! $ALT invariants regenerate "$LATEST" --output invariants_current.json --force --tolerance "${INVARIANTS_TOLERANCE}" >/dev/null 2>&1; then
    echo " Invariants regenerate failed with both binaries" >&2
    exit 2
  fi
  BIN=$ALT
fi

echo "[invariants-guard] Diffing invariants (current vs baseline)"
if ! $BIN invariants diff invariants_current.json --baseline "$BASELINE" --tolerance "${INVARIANTS_TOLERANCE}" --report invariants.json >/dev/null 2>&1; then
  echo " Invariants drift detected" >&2
  # Show top part of report
  head -n120 invariants.json || true
  # Provide ratio hints if jq is available
  if command -v jq >/dev/null 2>&1; then
    jq '{row_count, sum_effective_cost, sum_list_cost, sum_usage_quantity, violations}' invariants.json 2>/dev/null || true
  fi
  exit 3
fi

echo " Invariants guard passed (no drift) via $BIN"
echo $BIN > invariants_engine.txt
