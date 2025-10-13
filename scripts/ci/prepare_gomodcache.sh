#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Prepare GOMODCACHE in a runner-local temp directory and export it
if [[ -z "${GOMODCACHE:-}" ]]; then
  GOMODCACHE="${RUNNER_TEMP:-/tmp}/go/pkg/mod"
fi

mkdir -p "$GOMODCACHE"

# When CONSERVATIVE_GOMODCACHE=true, don't wipe cache (useful locally)
if [[ "${CONSERVATIVE_GOMODCACHE:-false}" != "true" ]]; then
  rm -rf "$GOMODCACHE"/* || true
  ci::log "Cleaned GOMODCACHE directory: $GOMODCACHE"
else
  ci::log "Conservative mode: not cleaning GOMODCACHE=$GOMODCACHE"
fi

ls -la "$GOMODCACHE" || true
echo "GOMODCACHE=$GOMODCACHE" >> "${GITHUB_ENV:-/dev/null}"
ci::log "Prepared GOMODCACHE=$GOMODCACHE"
