#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

KUBECONFIG_B64=${KUBECONFIG_B64:-}
IMAGE_REPO=${IMAGE_REPO:-}
IMAGE_TAG=${IMAGE_TAG:-}
COSTSCOPE_JWT_SECRET=${COSTSCOPE_JWT_SECRET:-}
HELM_DRY_RUN=${HELM_DRY_RUN:-false}

if [[ -z "$KUBECONFIG_B64" || -z "$IMAGE_REPO" || -z "$IMAGE_TAG" ]]; then
  ci::die "Missing required env. Need KUBECONFIG_B64, IMAGE_REPO, IMAGE_TAG"
fi

ci::require_cmd base64
ci::require_cmd helm

# Decode kubeconfig into a secure temporary file and clean it up on exit
KUBECONFIG_TMP="$(mktemp)"
# Best-effort protect file permissions on platforms that support it
chmod 600 "$KUBECONFIG_TMP" 2>/dev/null || true
cleanup() { rm -f "$KUBECONFIG_TMP" || true; }
trap cleanup EXIT INT TERM

printf '%s' "$KUBECONFIG_B64" | base64 -d > "$KUBECONFIG_TMP"
export KUBECONFIG="$KUBECONFIG_TMP"

ci::log "Deploying chart costscope with image ${IMAGE_REPO}:${IMAGE_TAG} (dry-run=${HELM_DRY_RUN})"
helm_args=(upgrade --install costscope ./charts/costscope \
  --set image.repository="$IMAGE_REPO" \
  --set image.tag="$IMAGE_TAG" \
  --set env.COSTSCOPE_JWT_SECRET="$COSTSCOPE_JWT_SECRET")
if [[ "$HELM_DRY_RUN" == "true" ]]; then
  helm_args+=(--dry-run --debug)
fi
helm "${helm_args[@]}"
