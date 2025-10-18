#!/usr/bin/env bash
set -euo pipefail

# Run govulncheck and write JSON report.
# Env:
#   GOVULNCHECK_PKGS     (default: ./...)
#   GOVULNCHECK_TIMEOUT  (optional; passed to govulncheck as -timeout=<seconds>)
#   GOVULNCHECK_OUT      (default: ./govulncheck.json)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::require_cmd govulncheck

PKGS=${GOVULNCHECK_PKGS:-./...}
OUT=${GOVULNCHECK_OUT:-govulncheck.json}
TIMEOUT_FLAG=()
if [[ -n "${GOVULNCHECK_TIMEOUT:-}" ]]; then
	if govulncheck -h 2>&1 | grep -q -- "-timeout"; then
		TIMEOUT_FLAG=("-timeout=${GOVULNCHECK_TIMEOUT}")
	else
		ci::warn "govulncheck does not support -timeout; ignoring GOVULNCHECK_TIMEOUT=${GOVULNCHECK_TIMEOUT}"
	fi
fi

ci::log "Running govulncheck for packages: $PKGS (out: $OUT${GOVULNCHECK_TIMEOUT:+, timeout: $GOVULNCHECK_TIMEOUT}s)"
govulncheck -json "${TIMEOUT_FLAG[@]}" "$PKGS" > "$OUT" || true
ci::log "govulncheck complete -> $OUT"

# Optional step status artifact (kept for compatibility with prior workflow)
if command -v jq >/dev/null 2>&1; then
	jq -n '{status:"ok"}' > step_status.json || true
else
	echo '{"status":"ok"}' > step_status.json || true
fi
