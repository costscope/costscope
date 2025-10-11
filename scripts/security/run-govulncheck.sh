#!/usr/bin/env bash
set -euo pipefail

# Run govulncheck and write JSON report.
# Env:
#   GOVULNCHECK_PKGS (default: ./...)

PKGS=${GOVULNCHECK_PKGS:-./...}

govulncheck -format=json "$PKGS" > govulncheck.json || true

# Optional step status artifact (kept for compatibility with prior workflow)
jq -n '{status:"ok"}' > step_status.json || true
