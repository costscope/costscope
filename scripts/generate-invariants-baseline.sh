#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMMON_SH="$ROOT_DIR/scripts/ci/lib/common.sh"
if [[ -f "$COMMON_SH" ]]; then
  # shellcheck disable=SC1090
  . "$COMMON_SH"
fi

if [[ $# -lt 2 ]]; then
  if command -v ci::die >/dev/null 2>&1; then
    ci::die "Usage: $0 <input_parquet> <output_baseline_json> [tolerance]"
  else
    echo "Usage: $0 <input_parquet> <output_baseline_json> [tolerance]" >&2
    exit 1
  fi
fi
INPUT=$1
OUTPUT=$2
TOLERANCE=${3:-0.01}

if command -v ci::require_cmd >/dev/null 2>&1; then
  ci::require_cmd jq || true
fi

# Use validate command with --invariants only; skip heavy domains for speed
# We rely on invariants report generation; no baseline compare on first run.

if [[ ! -x "$ROOT_DIR/bin/costscope" && ! -x "$ROOT_DIR/costscope" ]]; then
  if command -v ci::die >/dev/null 2>&1; then
    ci::die "costscope binary not found in repo root or bin/. Build first (make build)."
  else
    echo "costscope binary not found in repo root or bin/. Build first (make build)." >&2
    exit 2
  fi
fi

BIN="${BIN:-}"
if [[ -x "$ROOT_DIR/bin/costscope" ]]; then BIN="$ROOT_DIR/bin/costscope"; fi
if [[ -x "$ROOT_DIR/costscope" ]]; then BIN="$ROOT_DIR/costscope"; fi

"$BIN" validate "$INPUT" --schema --invariants \
  --invariants-report "/tmp/invariants-report.json" --invariants-tolerance "$TOLERANCE" --quiet || true

# Copy the generated report as new baseline (strip volatile fields if needed later)
cp /tmp/invariants-report.json "$OUTPUT"
if command -v ci::log >/dev/null 2>&1; then
  ci::log "Baseline written to $OUTPUT"
else
  echo "Baseline written to $OUTPUT"
fi
