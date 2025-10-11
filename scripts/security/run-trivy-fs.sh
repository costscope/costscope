#!/usr/bin/env bash
set -euo pipefail

# Run Trivy filesystem scan
# Env:
#   TRIVY_TARGET (default: .)
#   TRIVY_SEVERITY (default: UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL)

TARGET=${TRIVY_TARGET:-.}
SEV=${TRIVY_SEVERITY:-UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL}

trivy fs --severity "$SEV" --format json --output trivy-fs.json --ignore-unfixed --no-progress "$TARGET" || true
