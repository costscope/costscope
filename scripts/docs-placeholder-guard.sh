#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

if grep -R --include='*.md' -n "content moved verbatim" docs/ >/dev/null 2>&1; then
  ci::warn "Placeholder markers still present – complete migration before merging:"
  grep -Rn --include='*.md' "content moved verbatim" docs/ || true
  exit 1
fi
ci::log "No placeholder markers detected"
