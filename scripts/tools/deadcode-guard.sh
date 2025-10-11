#!/usr/bin/env bash
set -euo pipefail

# deadcode-guard.sh
# Purpose: Run deadcode tool, filter out allowlisted symbols, and fail on newly unreachable exports.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ALLOWLIST_FILE="${ROOT_DIR}/.deadcode-allowlist"

if ! command -v deadcode >/dev/null 2>&1; then
  echo "deadcode tool not found in PATH (install via: go install golang.org/x/tools/... or make tools)" >&2
  exit 2
fi

tmp_out="$(mktemp)"
trap 'rm -f "$tmp_out"' EXIT

# Run full deadcode scan (honor existing Makefile exclude patterns by reproducing logic if needed)
# We rely on the tool's default module recursion.
deadcode ./... > "$tmp_out" || true

if [[ ! -s "$tmp_out" ]]; then
  echo "No deadcode output (unexpected)." >&2
  exit 0
fi

if [[ ! -f "$ALLOWLIST_FILE" ]]; then
  echo "Allowlist file missing at $ALLOWLIST_FILE" >&2
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
  echo "Deadcode guard: no new unallowlisted symbols detected." >&2
  exit 0
fi

echo "Deadcode guard: NEW unreachable exported symbols detected (not in allowlist):" >&2
cat "$violations_file" >&2
echo >&2
echo "Add intentional cases to .deadcode-allowlist with justification comments, or remove the code." >&2
exit 1
