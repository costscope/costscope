#!/usr/bin/env bash
# Reproducible build verification script (M8).
# Performs two builds with a fixed SOURCE_DATE_EPOCH and compares sha256 sums.
# Adds a 1 second sleep between builds to ensure wall clock would differ without fixed epoch.
#
# Usage: ./scripts/repro-check.sh [SOURCE_DATE_EPOCH] [binary_name]
# Defaults:
#   SOURCE_DATE_EPOCH: 1716249600 (2024-05-21T00:00:00Z) – arbitrary stable example
#   binary_name: costscope
# Exit codes:
#   0 -> success (hashes identical)
#   1 -> drift detected (hashes differ)
#   >1 -> script/build error
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

EPOCH="${1:-1716249600}"
BIN_NAME="${2:-costscope}"

if ! command -v sha256sum >/dev/null 2>&1; then
  ci::die "sha256sum not found"
fi

# Require clean working tree for provenance
if ! git diff --quiet || ! git diff --cached --quiet; then
  ci::die "Working tree dirty; commit or stash changes before reproducibility check."
fi

tmpdir="$(mktemp -d)"
ci::log "[repro] epoch=$EPOCH tmp=$tmpdir"

build_once() {
  local label="$1"
  rm -f "$BIN_NAME"
  SOURCE_DATE_EPOCH="$EPOCH" make -s build-release >/dev/null 2>&1 || { ci::die "build failed"; }
  if [[ ! -f "$BIN_NAME" ]]; then
    echo "expected binary '$BIN_NAME' missing" >&2
    exit 3
  fi
  local sum
  sum="$(sha256sum "$BIN_NAME" | awk '{print $1}')"
  cp "$BIN_NAME" "$tmpdir/${BIN_NAME}_${label}"
  echo "$sum"
}

SUM1=$(build_once A)
sleep 1
SUM2=$(build_once B)

if [[ "$SUM1" != "$SUM2" ]]; then
  ci::warn "[repro] DRIFT: sha256 mismatch"
  echo " A: $SUM1" >&2
  echo " B: $SUM2" >&2
  echo "Artifacts: $tmpdir/${BIN_NAME}_A $tmpdir/${BIN_NAME}_B" >&2
  exit 1
fi

ci::log "[repro] OK: sha256 $SUM1 (artifacts in $tmpdir)"
exit 0
