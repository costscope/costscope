#!/usr/bin/env bash
set -euo pipefail
# invariants_guard.sh - Run invariants drift guard comparing current regenerated invariants vs baseline.
# Env vars:
#   OUT_DIR               Directory to write Parquet outputs into (default: costscope-data)
#   PARQUET_ROTATE_SIZE   Rotation size in bytes for Parquet segments (default: 10000000000)
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
OUT_DIR=${OUT_DIR:-costscope-data}
INV_CURRENT="$OUT_DIR/invariants_current.json"
INV_REPORT="$OUT_DIR/invariants.json"
INV_ENGINE="$OUT_DIR/invariants_engine.txt"

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

mkdir -p "$OUT_DIR"

echo "[invariants-guard] Converting fast path (preferred debug=$BIN)"
if ! $BIN convert --provider aws --input "$INPUT" --output "$OUT_DIR/focus_fast.parquet" --streaming --rotate-size "$PARQUET_ROTATE_SIZE" >/dev/null 2>&1; then
  echo "️  Debug convert failed; trying optimized binary"
  BIN=$OPT
  if ! $BIN convert --provider aws --input "$INPUT" --output "$OUT_DIR/focus_fast.parquet" --streaming --rotate-size "$PARQUET_ROTATE_SIZE" >/dev/null 2>&1; then
    echo " Conversion failed with both binaries" >&2
    exit 2
  fi
fi

LATEST=$(ls -1t "$OUT_DIR"/focus_fast*.parquet 2>/dev/null | head -n1)
if [[ -z "$LATEST" ]]; then
  echo " No parquet output" >&2
  exit 2
fi

echo "[invariants-guard] Regenerating current invariants from $LATEST"
if ! $BIN invariants regenerate "$LATEST" --output "$INV_CURRENT" --force --tolerance "${INVARIANTS_TOLERANCE}" >/dev/null 2>&1; then
  echo "️  Regenerate failed with $BIN; trying fallback binary"
  ALT=$([[ $BIN == $DBG ]] && echo $OPT || echo $DBG)
  if ! $ALT invariants regenerate "$LATEST" --output "$INV_CURRENT" --force --tolerance "${INVARIANTS_TOLERANCE}" >/dev/null 2>&1; then
    echo " Invariants regenerate failed with both binaries" >&2
    exit 2
  fi
  BIN=$ALT
fi

echo "[invariants-guard] Diffing invariants (current vs baseline)"
if ! $BIN invariants diff "$INV_CURRENT" --baseline "$BASELINE" --tolerance "${INVARIANTS_TOLERANCE}" --report "$INV_REPORT" >/dev/null 2>&1; then
  echo " Invariants drift detected" >&2
  # Show top part of report
  head -n120 "$INV_REPORT" || true
  # Provide ratio hints if jq is available
  if command -v jq >/dev/null 2>&1; then
    jq '{row_count, sum_effective_cost, sum_list_cost, sum_usage_quantity, violations}' "$INV_REPORT" 2>/dev/null || true
  fi
  exit 3
fi

echo " Invariants guard passed (no drift) via $BIN"
echo $BIN > "$INV_ENGINE"
