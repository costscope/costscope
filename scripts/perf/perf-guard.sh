#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# perf-guard.sh (M11 – Perf guard unified vs fast)
# Runs the targeted benchmark BenchmarkPerfGuardUnified (legacy vs unified sub-benchmarks)
# Parses go test -bench output (ns/op and B/op allocations) and enforces thresholds:
#   unified_duration / legacy_duration <= PERF_GUARD_DURATION_MAX (default 1.20)
#   unified_allocs   / legacy_allocs   <= PERF_GUARD_ALLOCS_MAX   (default 1.25)
# Exits non-zero on violation. Designed for CI gating.
#
# Usage: scripts/perf/perf-guard.sh [package (default ./internal/core/focus/conversion)]
# Env:
#   PERF_GUARD_DURATION_MAX (default 1.20)
#   PERF_GUARD_ALLOCS_MAX   (default 1.25)
#   PERF_GUARD_ROWS         (optional override of row count via benchmark env)

PKG="${1:-./internal/core/focus/conversion}"
DUR_MAX="${PERF_GUARD_DURATION_MAX:-1.20}"
ALLOC_MAX="${PERF_GUARD_ALLOCS_MAX:-1.25}"

ci::log " Running perf guard benchmark (pkg=$PKG dur<=${DUR_MAX}x alloc<=${ALLOC_MAX}x)"

BENCH_ENV=""
if [[ -n "${PERF_GUARD_ROWS:-}" ]]; then
  BENCH_ENV="PERF_GUARD_ROWS=${PERF_GUARD_ROWS}" 
fi

# Run only the perf guard benchmark; suppress other output with -run '^$'
set -o pipefail
CMD="env -u GOROOT GOTOOLCHAIN=auto COSTSCOPE_CORE_LOG_LEVEL=error COSTSCOPE_LOG_LEVEL=error LOG_LEVEL=error COSTSCOPE_LOG_FORMAT=discard $BENCH_ENV go test -count=1 -run '^$' -bench '^BenchmarkPerfGuardUnified$' -benchmem $PKG"
ci::debug "+ $CMD"
OUTPUT=$(eval "$CMD" 2>&1 | tee /dev/stderr)

# State-machine parse: remember last seen benchmark name; capture metrics when a numeric metrics line appears (contains ns/op)
legacy_ns=""; unified_ns=""; legacy_allocs=""; unified_allocs=""; current=""
grab_metrics_line=""
while IFS= read -r line; do
  # If the line starts with our benchmark name set current context
  if [[ $line =~ ^BenchmarkPerfGuardUnified/legacy- ]]; then current="legacy"; grab_metrics_line=1; continue; fi
  if [[ $line =~ ^BenchmarkPerfGuardUnified/unified- ]]; then current="unified"; grab_metrics_line=1; continue; fi
  # The actual metrics often appear on the next separate line (after logs) containing ns/op
  if [[ $grab_metrics_line == 1 && $line == *"ns/op"* ]]; then
    read -r -a toks <<<"$line"
    for i in "${!toks[@]}"; do
      if [[ ${toks[$i]} == "ns/op" && $i -ge 1 ]]; then ns_val=${toks[$((i-1))]}; fi
      if [[ ${toks[$i]} == "allocs/op" && $i -ge 1 ]]; then allocs_val=${toks[$((i-1))]}; fi
    done
    if [[ -n $current && -n ${ns_val:-} && -n ${allocs_val:-} ]]; then
      if [[ $current == legacy && -z $legacy_ns ]]; then legacy_ns=$ns_val; legacy_allocs=$allocs_val; fi
      if [[ $current == unified && -z $unified_ns ]]; then unified_ns=$ns_val; unified_allocs=$allocs_val; fi
    fi
    grab_metrics_line=0; ns_val=""; allocs_val=""; current=""
  fi
  [[ -n $legacy_ns && -n $unified_ns ]] && break
done <<<"$OUTPUT"

if [[ -z "$legacy_ns" || -z "$unified_ns" || -z "$legacy_allocs" || -z "$unified_allocs" ]]; then
  ci::warn " Failed to extract benchmark metrics (state-machine parser)"
  echo "$OUTPUT" | sed -n '1,200p' >&2
  exit 4
fi

ci::log " Parsed metrics: legacy_ns=$legacy_ns unified_ns=$unified_ns legacy_allocs=$legacy_allocs unified_allocs=$unified_allocs"

ratio_dur=$(awk -v u="$unified_ns" -v l="$legacy_ns" 'BEGIN{ if(l==0){print 999}else{printf "%.4f", u/l} }')
ratio_alloc=$(awk -v u="$unified_allocs" -v l="$legacy_allocs" 'BEGIN{ if(l==0){print 999}else{printf "%.4f", u/l} }')

status="pass"
fail_reasons=()

if awk -v r="$ratio_dur" -v max="$DUR_MAX" 'BEGIN{exit (r>max)?0:1}'; then
  status="fail"; fail_reasons+=("duration ratio $ratio_dur > $DUR_MAX")
fi
if awk -v r="$ratio_alloc" -v max="$ALLOC_MAX" 'BEGIN{exit (r>max)?0:1}'; then
  status="fail"; fail_reasons+=("alloc ratio $ratio_alloc > $ALLOC_MAX")
fi

ci::log " Perf guard results:"
ci::log "   legacy:  ns/op=$legacy_ns  allocs/op=$legacy_allocs"
ci::log "   unified: ns/op=$unified_ns allocs/op=$unified_allocs"
ci::log "   ratios:  duration=$ratio_dur allocs=$ratio_alloc"

if [[ "$status" == "fail" ]]; then
  ci::warn " Performance guard FAILED: ${fail_reasons[*]}"
  exit 10
fi

ci::log " Performance guard passed (duration=$ratio_dur allocs=$ratio_alloc)"

exit 0
