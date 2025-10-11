#!/usr/bin/env bash
set -euo pipefail

# Wrapper to run perf-bench with safe runtime OUT path computation and go fallback.
# Usage: run_perf_wrapper.sh <input_csv_gz> [extra args...]

INPUT=${1:-tests/perf/aws-cur-synth.csv.gz}
shift || true

# Compute output args (prefer GITHUB_WORKSPACE when available)
if [ -n "${GITHUB_WORKSPACE:-}" ]; then
  OUT_ARGS=( -output "${GITHUB_WORKSPACE}/bench_results.json" -prom-output "${GITHUB_WORKSPACE}/perf_metrics.prom" )
else
  OUT_ARGS=( -output bench_results.json -prom-output perf_metrics.prom )
fi

# Find a go binary: prefer PATH, else hostedtoolcache fallback
if command -v go >/dev/null 2>&1; then
  GO_BIN=$(command -v go)
else
  if ls /opt/hostedtoolcache/go/*/x64/bin/go >/dev/null 2>&1; then
    GO_BIN=$(ls -d /opt/hostedtoolcache/go/*/x64/bin/go | tail -n1)
  else
    echo "ERROR: go not available in PATH and no hostedtoolcache fallback found; please install go >= 1.24.6" >&2
    exit 2
  fi
fi

echo "Using go: $($GO_BIN version)"

# Run the perf bench tool
"$GO_BIN" run ./scripts/tools/perf-bench -input "$INPUT" "${OUT_ARGS[@]}" "$@" || { echo " Performance regression detected"; exit 1; }
