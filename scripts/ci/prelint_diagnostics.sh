#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

ci::log "Go environment:"
go env
echo
ci::log "Go list (packages visible to default build):"
go list -json ./... | sed -n '1,200p' || true
echo
ci::log "Files in internal/database (sample):"
ls -la internal/database | sed -n '1,200p' || true
