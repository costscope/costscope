#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

KUBECONFIG_B64=${KUBECONFIG_B64:-}
IMAGE_REPO=${IMAGE_REPO:-}
IMAGE_TAG=${IMAGE_TAG:-}
COSTSCOPE_JWT_SECRET=${COSTSCOPE_JWT_SECRET:-}

if [[ -z "$KUBECONFIG_B64" || -z "$IMAGE_REPO" || -z "$IMAGE_TAG" ]]; then
  ci::die "Missing required env. Need KUBECONFIG_B64, IMAGE_REPO, IMAGE_TAG"
fi

ci::require_cmd base64
ci::require_cmd helm

echo "$KUBECONFIG_B64" | base64 -d > kubeconfig
export KUBECONFIG=$PWD/kubeconfig

ci::log "Deploying chart costscope with image ${IMAGE_REPO}:${IMAGE_TAG}"
helm upgrade --install costscope ./charts/costscope \
  --set image.repository="$IMAGE_REPO" \
  --set image.tag="$IMAGE_TAG" \
  --set env.COSTSCOPE_JWT_SECRET="$COSTSCOPE_JWT_SECRET"
