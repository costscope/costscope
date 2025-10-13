#!/usr/bin/env bash
set -euo pipefail

# deadcode-guard.sh
# Purpose: Run deadcode tool, filter out allowlisted symbols, and fail on newly unreachable exports.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ALLOWLIST_FILE="${ROOT_DIR}/.deadcode-allowlist"

ci::require_cmd deadcode

tmp_out="$(mktemp)"
trap 'rm -f "$tmp_out"' EXIT

# Run full deadcode scan (honor existing Makefile exclude patterns by reproducing logic if needed)
# We rely on the tool's default module recursion.
deadcode ./... > "$tmp_out" || true

if [[ ! -s "$tmp_out" ]]; then
  ci::warn "No deadcode output (unexpected)."
  exit 0
fi

if [[ ! -f "$ALLOWLIST_FILE" ]]; then
  ci::die "Allowlist file missing at $ALLOWLIST_FILE" || true
  cat "$tmp_out"
  exit 3
fi

# Build a regex from allowlist (skip comments & blanks)
patterns_file="$(mktemp)"
grep -vE '^(#|//|\s*$)' "$ALLOWLIST_FILE" > "$patterns_file" || true

violations_file="$(mktemp)"
trap 'rm -f "$tmp_out" "$violations_file" "$patterns_file"' EXIT

while IFS= read -r line; do
  skip=0
  while IFS= read -r pat; do
    [ -z "$pat" ] && continue
    echo "$line" | grep -Eq "$pat" && { skip=1; break; }
  done < "$patterns_file"
  [ $skip -eq 1 ] && continue
  echo "$line" >> "$violations_file"
done < "$tmp_out"

if [[ ! -s "$violations_file" ]]; then
  ci::log "Deadcode guard: no new unallowlisted symbols detected."
  exit 0
fi

ci::warn "Deadcode guard: NEW unreachable exported symbols detected (not in allowlist):"
cat "$violations_file" >&2
echo >&2
ci::warn "Add intentional cases to .deadcode-allowlist with justification comments, or remove the code."
exit 1
