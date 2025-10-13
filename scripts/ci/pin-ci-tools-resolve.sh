#!/usr/bin/env bash
set -euo pipefail

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
  echo "OWNER is required" >&2
  exit 2
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
    echo "Either 'image' or 'tag' must be provided." >&2
    exit 2
  fi
  image="ghcr.io/${owner}/ci-base:${input_tag}"
fi

# Basic immutability check (avoid floating tags)
if echo "$image" | grep -qE ':(latest|main|dev)$'; then
  echo "Refusing to pin a floating tag: $image" >&2
  exit 3
fi

echo "image=$image" >> "${GITHUB_OUTPUT}"
echo "Resolved image: $image"
