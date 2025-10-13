#!/usr/bin/env bash
set -euo pipefail
# Simple guard: ensure migration doc declares no breaking changes pre-1.0.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
FILE="$ROOT_DIR/docs/MIGRATION_0.x_to_1.0.md"
if [[ ! -f "$FILE" ]]; then
  ci::die "ERROR: migration file not found: $FILE"
fi
if grep -q '^## BreakingChanges: none$' "$FILE"; then
  ci::log "OK: BreakingChanges none confirmed"
  exit 0
else
  ci::warn "FAIL: Expected '## BreakingChanges: none' line in migration doc"
  grep -n 'BreakingChanges' "$FILE" || true
  exit 1
fi
