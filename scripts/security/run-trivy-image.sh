#!/usr/bin/env bash
set -euo pipefail

# Run Trivy image scan for a locally available image
# Env:
#   TRIVY_IMAGE (required)
#   TRIVY_SEVERITY (default: UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL)

IMAGE=${TRIVY_IMAGE:?TRIVY_IMAGE is required}
SEV=${TRIVY_SEVERITY:-UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL}

trivy image --severity "$SEV" --format json --output trivy-image.json --ignore-unfixed --no-progress "$IMAGE" || true
