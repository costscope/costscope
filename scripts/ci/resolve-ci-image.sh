#!/usr/bin/env bash
set -euo pipefail

# resolve-ci-image.sh
# Emits a single output key 'image' with the fully qualified image reference
# built from CI_IMAGE_REPO (defaults to ghcr.io/<owner>/ci-base) and CI_IMAGE_TAG.
# Usage (in GitHub Actions step):
#   - id: out
#     env:
#       CI_IMAGE_REPO: ghcr.io/${{ github.repository_owner }}/ci-base
#     run: bash ./scripts/ci/resolve-ci-image.sh
# Then use: ${{ steps.out.outputs.image }} as a container image.

# Infer GitHub owner for local or Actions contexts
gh_owner="${GITHUB_REPOSITORY_OWNER:-}"
if [[ -z "$gh_owner" && -n "${GITHUB_REPOSITORY:-}" ]]; then
  gh_owner="${GITHUB_REPOSITORY%/*}"
fi

default_repo="ghcr.io/${gh_owner:-costscope}/ci-base"
repo="${CI_IMAGE_REPO:-$default_repo}"

tag="${CI_IMAGE_TAG:-}"
if [[ -z "$tag" ]]; then
  if [[ "${RESOLVE_CI_IMAGE_ALLOW_LATEST:-false}" == "true" || "${IS_ACT:-false}" == "true" ]]; then
    echo "[resolve-ci-image] CI_IMAGE_TAG is not set; using 'latest' due to explicit allowance (RESOLVE_CI_IMAGE_ALLOW_LATEST=true or act)" >&2
    tag="latest"
  else
    echo "[resolve-ci-image] CI_IMAGE_TAG is not set and latest is not allowed; please set CI_IMAGE_TAG or RESOLVE_CI_IMAGE_ALLOW_LATEST=true" >&2
    exit 2
  fi
fi

img="${repo}:${tag}"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "image=${img}" >>"$GITHUB_OUTPUT"
else
  # Best-effort fallback for local runs
  echo "image=${img}"
fi

echo "[resolve-ci-image] Using image: ${img}" >&2
