#!/usr/bin/env bash
set -euo pipefail

KUBECONFIG_B64=${KUBECONFIG_B64:-}
IMAGE_REPO=${IMAGE_REPO:-}
IMAGE_TAG=${IMAGE_TAG:-}
COSTSCOPE_JWT_SECRET=${COSTSCOPE_JWT_SECRET:-}

if [[ -z "$KUBECONFIG_B64" || -z "$IMAGE_REPO" || -z "$IMAGE_TAG" ]]; then
  echo "Missing required env. Need KUBECONFIG_B64, IMAGE_REPO, IMAGE_TAG" >&2
  exit 2
fi

echo "$KUBECONFIG_B64" | base64 -d > kubeconfig
export KUBECONFIG=$PWD/kubeconfig

helm upgrade --install costscope ./charts/costscope \
  --set image.repository="$IMAGE_REPO" \
  --set image.tag="$IMAGE_TAG" \
  --set env.COSTSCOPE_JWT_SECRET="$COSTSCOPE_JWT_SECRET"
