#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

ci::require_cmd go
ci::log "Running perf engine micro benchmark"
go run ./scripts/tools/perf-bench -input tests/perf/aws-cur-synth.csv.gz \
  -iterations 1 -include-perf-engine -perf-ops 3000 -output "${GITHUB_WORKSPACE:-$PWD}/bench_results_perf_engine.json" || true
