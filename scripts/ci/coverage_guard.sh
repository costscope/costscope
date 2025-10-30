#!/usr/bin/env bash
set -euo pipefail

# coverage_guard.sh
# Generic coverage guard for narrow package groups.
# Modes:
#   production (default) -> internal/core/production
#   mapping             -> internal/framework/mapping
# Baseline files:
#   configs/coverage/<mode>.baseline (plain number, e.g. 92 or 87.2)
# Exit codes:
#   0  success / within threshold
#   2  coverage below min threshold
#  >2  unexpected execution error
# Options:
#   --mode <production|mapping>
#   --auto-bump           (increase baseline file if current >= baseline+0.3)
#   --json                (emit JSON summary to stdout)
# Environment overrides:
#   COV_ALLOWED_DRIFT (default 2.0)
#   COV_MIN_FLOOR     (optional absolute floor; if set overrides computed min when higher)
#   COV_BUMP_STEP     (default dynamic: actual current value)

MODE="production"
AUTO_BUMP=0
EMIT_JSON=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="$2"; shift 2;;
    --auto-bump)
      AUTO_BUMP=1; shift;;
    --json)
      EMIT_JSON=1; shift;;
    -h|--help)
      cat <<EOF
Usage: $0 [--mode production|mapping] [--auto-bump] [--json]
EOF
      exit 0;;
    *) echo "Unknown arg: $1" >&2; exit 3;;
  esac
done

case "$MODE" in
  production)
    PKG="./internal/core/production" ;;
  mapping)
    PKG="./internal/framework/mapping" ;;
  *) echo "[coverage-guard] unsupported mode: $MODE" >&2; exit 3;;
esac

# Resolve repo root robustly: prefer git, else derive from this script path (../../..)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if root_from_git=$(git rev-parse --show-toplevel 2>/dev/null); then
  REPO_ROOT="$root_from_git"
else
  REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
fi
BASELINE_FILE="$REPO_ROOT/configs/coverage/${MODE}.baseline"
if [[ ! -f "$BASELINE_FILE" ]]; then
  echo "[coverage-guard] baseline file missing: $BASELINE_FILE" >&2
  exit 3
fi

baseline_raw=$(<"$BASELINE_FILE")
# strip
baseline=$(echo "$baseline_raw" | tr -d ' \t\r\n')
if [[ -z "$baseline" ]]; then
  echo "[coverage-guard] empty baseline in $BASELINE_FILE" >&2
  exit 3
fi

allowed=${COV_ALLOWED_DRIFT:-2.0}
min=$(awk -v b="$baseline" -v a="$allowed" 'BEGIN{printf "%.1f", b - a}')
if [[ -n "${COV_MIN_FLOOR:-}" ]]; then
  # If floor > computed min use floor
  cmp=$(awk -v m="$min" -v f="$COV_MIN_FLOOR" 'BEGIN{ if (f+0 > m+0) print 1; else print 0 }')
  if [[ "$cmp" == "1" ]]; then
    min=$(awk -v f="$COV_MIN_FLOOR" 'BEGIN{printf "%.1f", f}')
  fi
fi

tmp=$(mktemp)
echo "[coverage-guard] mode=$MODE pkg=$PKG baseline=$baseline allowed=$allowed min=$min" >&2
## Silence noisy Go VCS stamping warnings in CI logs without affecting build outputs
goflags_init="${GOFLAGS:-}"
if [[ -z "$goflags_init" ]]; then
  goflags_effective="-buildvcs=false"
else
  goflags_effective="$goflags_init -buildvcs=false"
fi
if ! GOFLAGS="$goflags_effective" go test -count=1 -coverprofile="$tmp" "$PKG" 2>&1 | sed 's/^/[coverage-guard] test: /' >&2; then
  echo "[coverage-guard] test run failed" >&2
  rm -f "$tmp"
  exit 1
fi
cur=$(go tool cover -func="$tmp" | awk '/total:/ {gsub("%","",$3); print $3}')
rm -f "$tmp"

# Numeric compare
ok=1
awk -v c="$cur" -v m="$min" 'BEGIN{ if (c+0 < m+0) exit 1 }' || ok=0

if [[ $EMIT_JSON == 1 ]]; then
  printf '{"mode":"%s","current":%.1f,"baseline":%.1f,"allowed":%.1f,"min":%.1f,"status":"%s"}\n' \
    "$MODE" "$cur" "$baseline" "$allowed" "$min" "$([ $ok -eq 1 ] && echo OK || echo FAIL)"
else
  echo "[coverage-guard] current=$cur baseline=$baseline min=$min" >&2
fi

if [[ $ok -eq 0 ]]; then
  echo "[coverage-guard] FAIL: $cur < $min" >&2
  exit 2
fi

echo "[coverage-guard] OK" >&2

if [[ $AUTO_BUMP == 1 ]]; then
  # bump if gain >= 0.3
  bump_needed=$(awk -v c="$cur" -v b="$baseline" 'BEGIN{ if (c - b >= 0.3) print 1; else print 0 }')
  if [[ "$bump_needed" == "1" ]]; then
    echo "[coverage-guard] auto-bump: $baseline -> $cur" >&2
    printf '%.1f\n' "$cur" > "$BASELINE_FILE"
  else
    echo "[coverage-guard] auto-bump: skipped (delta < 0.3)" >&2
  fi
fi
