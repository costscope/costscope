#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# allowlist-rationale-lint.sh
# Fails if any non-comment, non-empty line in .deadcode-allowlist lacks '# rationale:'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ALLOWLIST_FILE="${ROOT_DIR}/.deadcode-allowlist"

if [[ ! -f "$ALLOWLIST_FILE" ]]; then
  ci::die "Allowlist file not found: $ALLOWLIST_FILE"
fi

bad=()
while IFS= read -r l; do
  [[ -z "$l" ]] && continue
  [[ "$l" =~ ^(#|//) ]] && continue
  # Extract symbol token (first whitespace-delimited part)
  sym="${l%% *}"
  if [[ "$l" != *"# rationale:"* ]]; then
    bad+=("$sym")
  fi
done < "$ALLOWLIST_FILE"

if ((${#bad[@]})); then
  ci::warn " Allowlist rationale lint failed. Missing rationale comments for symbols:"
  printf ' - %s\n' "${bad[@]}" >&2
  echo "Each line must include '# rationale: <reason>' after the symbol." >&2
  exit 1
fi

ci::log " Allowlist rationale lint passed."