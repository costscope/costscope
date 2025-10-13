#!/usr/bin/env bash
set -euo pipefail

# Run gosec over the codebase and output JSON report.
# Env:
#   GOSEC_PKGS (default: ./...)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::require_cmd gosec

PKGS=${GOSEC_PKGS:-./...}

ci::log "Running gosec on: $PKGS"
gosec -fmt=json -out gosec.json "$PKGS" || true
ci::log "gosec complete -> gosec.json"
