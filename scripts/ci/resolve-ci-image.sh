#!/usr/bin/env bash
set -euo pipefail

# resolve-ci-image.sh
# Emits a single output key 'image' with the fully qualified image reference.
# Resolution priority:
#  1) CI_TOOLS_IMAGE (full ref, e.g., ghcr.io/org/ci-base:tag)
#  2) CI_IMAGE_REPO + CI_IMAGE_TAG (repo + tag)
#  3) CI_IMAGE_REPO + 'latest' when explicitly allowed via RESOLVE_CI_IMAGE_ALLOW_LATEST=true
# Usage (in GitHub Actions step):
#   - id: out
#     env:
#       CI_IMAGE_REPO: ghcr.io/${{ github.repository_owner }}/ci-base
#       CI_IMAGE_TAG: ${{ vars.CI_IMAGE_TAG }}
#       CI_TOOLS_IMAGE: ${{ vars.CI_TOOLS_IMAGE }}
#     run: bash ./scripts/ci/resolve-ci-image.sh
# Then use: ${{ steps.out.outputs.image }} as a container image.

# Infer GitHub owner for local or Actions contexts
gh_owner="${GITHUB_REPOSITORY_OWNER:-}"
if [[ -z "$gh_owner" && -n "${GITHUB_REPOSITORY:-}" ]]; then
  gh_owner="${GITHUB_REPOSITORY%/*}"
fi

default_repo="ghcr.io/${gh_owner:-costscope}/ci-base"
repo="${CI_IMAGE_REPO:-$default_repo}"

# 1) Full image override takes precedence
if [[ -n "${CI_TOOLS_IMAGE:-}" ]]; then
  img="${CI_TOOLS_IMAGE}"
else
  # 2) Combine repo + tag (or 3) 'latest' if allowed)
  tag="${CI_IMAGE_TAG:-}"
  if [[ -z "$tag" ]]; then
    if [[ "${RESOLVE_CI_IMAGE_ALLOW_LATEST:-false}" == "true" || "${IS_ACT:-false}" == "true" ]]; then
      echo "[resolve-ci-image] CI_IMAGE_TAG is not set; using 'latest' due to explicit allowance (RESOLVE_CI_IMAGE_ALLOW_LATEST=true or act)" >&2
      tag="latest"
    else
      echo "[resolve-ci-image] CI_IMAGE_TAG is not set and latest is not allowed; please set CI_IMAGE_TAG, CI_TOOLS_IMAGE, or RESOLVE_CI_IMAGE_ALLOW_LATEST=true" >&2
      exit 2
    fi
  fi
  img="${repo}:${tag}"
fi

# Under local act runs, prefer a broadly available base for specialized images that might not exist yet.
# This avoids manifest errors like: 'manifest unknown' when pulling ghcr.io/<owner>/ci-shellcheck:latest
if [[ "${IS_ACT:-false}" == "true" ]]; then
  repo_basename="${repo##*/}"
  if [[ "$repo_basename" == "ci-shellcheck" ]]; then
    echo "[resolve-ci-image] act detected and repo is ci-shellcheck; using fallback image 'catthehacker/ubuntu:full-latest'" >&2
    img="catthehacker/ubuntu:full-latest"
  fi
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "image=${img}" >>"$GITHUB_OUTPUT"
else
  # Best-effort fallback for local runs
  echo "image=${img}"
fi

echo "[resolve-ci-image] Using image: ${img}" >&2
