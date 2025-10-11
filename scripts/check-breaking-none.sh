#!/usr/bin/env bash
set -euo pipefail
# Simple guard: ensure migration doc declares no breaking changes pre-1.0.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
FILE="$ROOT_DIR/docs/MIGRATION_0.x_to_1.0.md"
if [[ ! -f "$FILE" ]]; then
  echo "ERROR: migration file not found: $FILE" >&2
  exit 2
fi
if grep -q '^## BreakingChanges: none$' "$FILE"; then
  echo "OK: BreakingChanges none confirmed"
  exit 0
else
  echo "FAIL: Expected '## BreakingChanges: none' line in migration doc" >&2
  grep -n 'BreakingChanges' "$FILE" || true
  exit 1
fi
