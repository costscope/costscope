#!/usr/bin/env bash
set -euo pipefail

# Run Trivy filesystem scan
# Env:
#   TRIVY_TARGET (default: .)
#   TRIVY_SEVERITY (default: UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::require_cmd trivy

TARGET=${TRIVY_TARGET:-.}
SEV=${TRIVY_SEVERITY:-UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL}

ci::log "Running Trivy FS scan (target=$TARGET, severity=$SEV)"
trivy fs --severity "$SEV" --format json --output trivy-fs.json --ignore-unfixed --no-progress "$TARGET" || true
ci::log "Trivy FS scan complete -> trivy-fs.json"
