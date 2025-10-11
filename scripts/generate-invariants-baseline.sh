#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <input_parquet> <output_baseline_json> [tolerance]" >&2
  exit 1
fi
INPUT=$1
OUTPUT=$2
TOLERANCE=${3:-0.01}

# Use validate command with --invariants only; skip heavy domains for speed
# We rely on invariants report generation; no baseline compare on first run.

if ! command -v ./costscope >/dev/null 2>&1; then
  echo "costscope binary not found in current directory; build first (make build)." >&2
  exit 2
fi

./costscope validate "$INPUT" --schema --invariants \
  --invariants-report "/tmp/invariants-report.json" --invariants-tolerance "$TOLERANCE" --quiet || true

# Copy the generated report as new baseline (strip volatile fields if needed later)
cp /tmp/invariants-report.json "$OUTPUT"
echo "Baseline written to $OUTPUT"
