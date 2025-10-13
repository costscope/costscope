#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=ci/lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/ci/lib/common.sh"

ci::log "Linting markdown for trailing whitespace..."
if grep -R --include='*.md' "[[:blank:]]$" docs/ >/dev/null 2>&1; then
  ci::die "Trailing spaces found in docs (run: git diff -- docs | sed -n '1,200p')"
fi
ci::log "No trailing spaces found"
