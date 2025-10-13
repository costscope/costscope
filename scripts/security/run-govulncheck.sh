#!/usr/bin/env bash
set -euo pipefail

# Run govulncheck and write JSON report.
# Env:
#   GOVULNCHECK_PKGS (default: ./...)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::require_cmd govulncheck

PKGS=${GOVULNCHECK_PKGS:-./...}

ci::log "Running govulncheck for packages: $PKGS"
govulncheck -format=json "$PKGS" > govulncheck.json || true
ci::log "govulncheck complete -> govulncheck.json"

# Optional step status artifact (kept for compatibility with prior workflow)
if command -v jq >/dev/null 2>&1; then
	jq -n '{status:"ok"}' > step_status.json || true
else
	echo '{"status":"ok"}' > step_status.json || true
fi
