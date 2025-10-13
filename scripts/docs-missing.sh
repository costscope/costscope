#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

missing=0
for f in README.md docs/ARCHITECTURE.md docs/api/index.md; do
  if [ ! -f "$f" ]; then
    ci::warn "Missing $f"
    missing=1
  fi
done
if [ "$missing" -eq 1 ]; then
  ci::die "Missing required docs"
fi
ci::log "Core docs present"
