#!/usr/bin/env bash
set -euo pipefail
# parity_guard.sh - Generate fast & unified parquet outputs and compare aggregate parity.
# Env vars:
#   OUT_DIR               Directory to write Parquet outputs into (default: costscope-data)
#   PARQUET_ROTATE_SIZE   Rotation size in bytes for Parquet segments (default: 10000000000)
# Exit codes:
#  0 success
#  2 parity mismatch
#  Other: unexpected failure

PARITY_TOLERANCE=${PARITY_TOLERANCE:-1e-9}
PARQUET_ROTATE_SIZE=${PARQUET_ROTATE_SIZE:-10000000000}
INPUT=${INPUT:-tests/perf/aws-cur-synth.csv.gz}
BIN=${BIN:-bin/costscope}
OUT_DIR=${OUT_DIR:-costscope-data}

if [[ ! -x "$BIN" ]]; then
  echo "[parity-guard] binary '$BIN' not found or not executable" >&2
  exit 1
fi
if [[ ! -f "$INPUT" ]]; then
  echo "[parity-guard] input dataset '$INPUT' missing" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"
PARITY_JSON="$OUT_DIR/parity.json"

echo "[parity-guard] Converting (fast path) → $OUT_DIR/focus_fast.parquet"
$BIN convert --provider aws --input "$INPUT" --output "$OUT_DIR/focus_fast.parquet" --streaming --rotate-size "$PARQUET_ROTATE_SIZE" >/dev/null 2>&1 || { echo " fast convert failed"; exit 1; }

echo "[parity-guard] Converting (unified mapper) → $OUT_DIR/focus_unified.parquet"
COSTSCOPE_USE_UNIFIED_MAPPER=1 $BIN convert --provider aws --input "$INPUT" --output "$OUT_DIR/focus_unified.parquet" --streaming --rotate-size "$PARQUET_ROTATE_SIZE" >/dev/null 2>&1 || { echo " unified convert failed"; exit 1; }

echo "[parity-guard] Running parity-check (lite hash enabled)"
if ! go run ./scripts/tools/parity-check --legacy "$OUT_DIR/focus_fast.parquet" -unified "$OUT_DIR/focus_unified.parquet" -tolerance "${PARITY_TOLERANCE}" -out "$PARITY_JSON" >/dev/null 2>&1; then
  echo " Parity mismatch" >&2
  exit 2
fi

# Print compact summary (nested fields)
jq '{
  legacy: {
    effective_cost_sum: .legacy.effective_cost_sum,
    usage_quantity_sum: .legacy.usage_quantity_sum,
    record_count: .legacy.record_count
  },
  unified: {
    effective_cost_sum: .unified.effective_cost_sum,
    usage_quantity_sum: .unified.usage_quantity_sum,
    record_count: .unified.record_count
  },
  equal_cost, equal_usage, equal_records, equal_lite_hash, duration_ms
}' "$PARITY_JSON" 2>/dev/null || true

echo "[parity-guard] Passed"
