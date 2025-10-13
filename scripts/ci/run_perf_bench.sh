#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

OUT_JSON="${GITHUB_WORKSPACE:-$PWD}/bench_results.json"
OUT_PROM="${GITHUB_WORKSPACE:-$PWD}/perf_metrics.prom"
BASELINE="tests/perf/baseline_bench_results.json"

EXTRA_OUT_ARGS=(--baseline "$BASELINE" -output "$OUT_JSON" -prom-output "$OUT_PROM")
ci::log "Running perf bench with baseline=$BASELINE"
make perf-bench-synth EXTRA_ARGS="${EXTRA_OUT_ARGS[*]}"
