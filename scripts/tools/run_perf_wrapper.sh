#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# Wrapper to run perf-bench with safe runtime OUT path computation and go fallback.
# Usage: run_perf_wrapper.sh <input_csv_gz> [extra args...]

INPUT=${1:-tests/perf/aws-cur-synth.csv.gz}
shift || true

# Compute output args
# Priority: OUT_DIR if set -> GITHUB_WORKSPACE if set -> current directory
if [ -n "${OUT_DIR:-}" ]; then
  mkdir -p "${OUT_DIR}"
  OUT_ARGS=( -output "${OUT_DIR}/bench_results.json" -prom-output "${OUT_DIR}/perf_metrics.prom" )
elif [ -n "${GITHUB_WORKSPACE:-}" ]; then
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

ci::log "Using go: $($GO_BIN version)"

# Run the perf bench tool
"$GO_BIN" run ./scripts/tools/perf-bench -input "$INPUT" "${OUT_ARGS[@]}" "$@" || { ci::warn " Performance regression detected"; exit 1; }
