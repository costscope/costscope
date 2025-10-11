#!/usr/bin/env bash
set -euo pipefail

# allowlist-rationale-lint.sh
# Fails if any non-comment, non-empty line in .deadcode-allowlist lacks '# rationale:'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ALLOWLIST_FILE="${ROOT_DIR}/.deadcode-allowlist"

if [[ ! -f "$ALLOWLIST_FILE" ]]; then
  echo "Allowlist file not found: $ALLOWLIST_FILE" >&2
  exit 2
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
  echo " Allowlist rationale lint failed. Missing rationale comments for symbols:" >&2
  printf ' - %s\n' "${bad[@]}" >&2
  echo "Each line must include '# rationale: <reason>' after the symbol." >&2
  exit 1
fi

echo " Allowlist rationale lint passed." >&2