#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# pin-ci-tools-resolve.sh
# Resolves the target CI tools image reference based on event payload or manual inputs
# Inputs (env):
#   OWNER           - repository owner (required)
#   CLIENT_IMAGE    - github.event.client_payload.image (optional)
#   CLIENT_TAG      - github.event.client_payload.tag (optional)
#   INPUT_IMAGE     - inputs.image (optional)
#   INPUT_TAG       - inputs.tag (optional)
# Outputs:
#   Writes "image=<value>" to $GITHUB_OUTPUT

owner="${OWNER:-}"
if [[ -z "$owner" ]]; then
  ci::die "OWNER is required"
fi

client_image="${CLIENT_IMAGE:-}"
client_tag="${CLIENT_TAG:-}"
manual_image="${INPUT_IMAGE:-}"
manual_tag="${INPUT_TAG:-}"

input_image="$client_image"
input_tag="$client_tag"

# Fallback to manual inputs if client payload empty
if [[ -z "$input_image" && -z "$input_tag" ]]; then
  input_image="$manual_image"
  input_tag="$manual_tag"
fi

if [[ -n "$input_image" ]]; then
  image="$input_image"
else
  if [[ -z "$input_tag" ]]; then
    # Under act, prefer CI_IMAGE_TAG from env instead of hardcoding any default
    if ci::is_act; then
      ci_tag="${CI_IMAGE_TAG:-}"
      if [[ -z "$ci_tag" ]]; then
        ci::die "No image/tag provided under act and CI_IMAGE_TAG is empty. Set CI_IMAGE_TAG in .github/act.env (see act.env.example) or provide inputs.image/tag."
      fi
      if [[ "$ci_tag" =~ ^sha256:[0-9a-f]{64}$ ]]; then
        image="ghcr.io/${owner}/ci-base@${ci_tag}"
      else
        image="ghcr.io/${owner}/ci-base:${ci_tag}"
      fi
      ci::warn "No inputs provided; using CI_IMAGE_TAG for act: ${image}"
    else
      ci::die "Either 'image' or 'tag' must be provided."
    fi
  else
    # If tag is a digest, use @; otherwise use :
    if [[ "$input_tag" =~ ^sha256:[0-9a-f]{64}$ ]]; then
      image="ghcr.io/${owner}/ci-base@${input_tag}"
    else
      image="ghcr.io/${owner}/ci-base:${input_tag}"
    fi
  fi
fi

# Basic immutability check (avoid floating tags)
if echo "$image" | grep -qE ':(latest|main|dev)$'; then
  ci::die "Refusing to pin a floating tag: $image"
fi

echo "image=$image" >> "${GITHUB_OUTPUT:-/dev/stdout}"
ci::log "Resolved image: $image"
