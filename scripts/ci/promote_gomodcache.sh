#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

TMP_CACHE="${RUNNER_TEMP:-/tmp}/go-mod-cache-restore"
mkdir -p "$TMP_CACHE" "${GOMODCACHE:-}"
ci::log "Promoting cached modules from $TMP_CACHE -> ${GOMODCACHE:-<unset>}"
rsync -a --delete "$TMP_CACHE/" "${GOMODCACHE:-$TMP_CACHE}/" || true
ci::log "GOMODCACHE contents after promotion:"
ls -la "${GOMODCACHE:-$TMP_CACHE}" | sed -n '1,120p' || true
