#!/usr/bin/env bash
set -euo pipefail

if grep -R --include='*.md' -n "content moved verbatim" docs/ >/dev/null 2>&1; then
  echo "Placeholder markers still present – complete migration before merging:"
  grep -Rn --include='*.md' "content moved verbatim" docs/ || true
  exit 1
fi
echo "No placeholder markers detected"
