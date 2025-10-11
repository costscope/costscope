#!/usr/bin/env bash
set -euo pipefail

# run_bench.sh - Unified performance bench runner for legacy vs unified mapper.
# Scenarios: chunk sizes 10k/50k/100k, rotation on/off (handled inside perf-bench tool).
# Outputs: bench_results.json, optional perf_metrics.prom, compares against baseline if provided.
# SLA thresholds (can override via env): duration <=20% slower, allocations <=25% higher.
# Usage:
#   scripts/perf/run_bench.sh [-i input.csv[.gz]] [-b baseline.json] [-o out.json] [-p prom.out]
# Env overrides:
#   PERF_BENCH_DURATION_MAX (default 1.20)
#   PERF_BENCH_ALLOC_MAX    (default 1.25)

INPUT="demo/focus-conversion/demo-cur-data.csv"
BASELINE=""
OUTPUT="bench_results.json"
PROM_OUT=""
ITERATIONS=1

while getopts "i:b:o:p:n:" opt; do
  case "$opt" in
    i) INPUT="$OPTARG" ;;
    b) BASELINE="$OPTARG" ;;
    o) OUTPUT="$OPTARG" ;;
    p) PROM_OUT="$OPTARG" ;;
    n) ITERATIONS="$OPTARG" ;;
    *) echo "Usage: $0 [-i input] [-b baseline.json] [-o output.json] [-p prom_output] [-n iterations]" >&2; exit 2 ;;
  esac
done

DUR_MAX=${PERF_BENCH_DURATION_MAX:-1.20}
ALLOC_MAX=${PERF_BENCH_ALLOC_MAX:-1.25}

set -x
ARGS=( -input "$INPUT" -output "$OUTPUT" -duration-max "$DUR_MAX" -alloc-max "$ALLOC_MAX" -iterations "$ITERATIONS" )
if [[ -n "$BASELINE" ]]; then
  ARGS+=( -baseline "$BASELINE" )
fi
if [[ -n "$PROM_OUT" ]]; then
  ARGS+=( -prom-output "$PROM_OUT" )
fi

go run ./scripts/tools/perf-bench "${ARGS[@]}"
set +x

status=$(jq -r '.overall_status' "$OUTPUT" 2>/dev/null || echo unknown)
if [[ "$status" != "pass" ]]; then
  echo " Performance regression detected (status=$status). See $OUTPUT" >&2
  exit 1
fi

echo " Performance benchmarks passed (status=$status). Results: $OUTPUT"
