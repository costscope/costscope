#!/usr/bin/env bash
set -euo pipefail

# Run gosec over the codebase and output JSON report.
# Env:
#   GOSEC_PKGS (default: ./...)

PKGS=${GOSEC_PKGS:-./...}

gosec -fmt=json -out gosec.json "$PKGS" || true
