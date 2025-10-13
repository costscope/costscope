#!/usr/bin/env bash
set -euo pipefail

# Run Trivy image scan for a locally available image
# Env:
#   TRIVY_IMAGE (required)
#   TRIVY_SEVERITY (default: UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::require_cmd trivy

IMAGE=${TRIVY_IMAGE:?TRIVY_IMAGE is required}
SEV=${TRIVY_SEVERITY:-UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL}

ci::log "Running Trivy image scan (image=$IMAGE, severity=$SEV)"
trivy image --severity "$SEV" --format json --output trivy-image.json --ignore-unfixed --no-progress "$IMAGE" || true
ci::log "Trivy image scan complete -> trivy-image.json"
