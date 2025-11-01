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

# Validate required inputs. KUBECONFIG_B64 is only required for non-dry-run (or dry-run with cluster).
if [[ -z "$IMAGE_REPO" || -z "$IMAGE_TAG" ]]; then
  ci::die "Missing required env. Need IMAGE_REPO, IMAGE_TAG"
fi

ci::require_cmd base64
ci::require_cmd helm

# If dry-run and kubeconfig is not provided, run lint + template without cluster access and exit.
if [[ "$HELM_DRY_RUN" == "true" && -z "$KUBECONFIG_B64" ]]; then
  ci::log "HELM_DRY_RUN=true and no KUBECONFIG provided; running 'helm lint' and 'helm template'"
  helm lint ./charts/costscope
  # Render manifests to ensure values are valid
  helm template costscope ./charts/costscope \
    --set image.repository="$IMAGE_REPO" \
    --set image.tag="$IMAGE_TAG" \
    --set env.COSTSCOPE_JWT_SECRET="$COSTSCOPE_JWT_SECRET" >/dev/null
  ci::log "Helm chart rendered successfully (dry-run without cluster)."
  exit 0
fi

# Decode kubeconfig into a secure temporary file and clean it up on exit
KUBECONFIG_TMP="$(mktemp)"
# Best-effort protect file permissions on platforms that support it
chmod 600 "$KUBECONFIG_TMP" 2>/dev/null || true
cleanup() { rm -f "$KUBECONFIG_TMP" || true; }
trap cleanup EXIT INT TERM

# Decode kubeconfig (be tolerant to whitespace pasted into secrets). If decoding
# fails, as a last resort accept raw kubeconfig YAML when it looks like one.
if ! printf '%s' "$KUBECONFIG_B64" | base64 -d -i > "$KUBECONFIG_TMP" 2>/dev/null; then
  ci::warn "KUBECONFIG_B64 failed to decode as base64; attempting raw kubeconfig fallback"
  if grep -qE '^(apiVersion:|clusters:|kind: *Config)' <<<"$KUBECONFIG_B64"; then
    # Looks like a kubeconfig YAML pasted directly as a secret; write it as-is
    printf '%s\n' "$KUBECONFIG_B64" > "$KUBECONFIG_TMP"
  else
    ci::die "KUBECONFIG_B64 is invalid base64 and does not resemble a kubeconfig. Regenerate the secret.\nLinux:  base64 -w 0 ~/.kube/config\nmacOS:  base64 < ~/.kube/config | tr -d '\n'"
  fi
fi
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
