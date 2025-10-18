#!/usr/bin/env bash
set -euo pipefail

# detect-primary-image-ref.sh
# Determines the primary image reference from docker/metadata-action tags list.
# Usage:
#   detect-primary-image-ref.sh [tags]
# If no positional arg provided, reads from STDIN. Emits 'image_ref=' to GITHUB_OUTPUT.

tags_input="${1-}"
if [[ -z "$tags_input" ]]; then
  # Read from stdin into variable (preserve newlines)
  if IFS= read -r -d '' tags_input; then :; fi || true
  # If above didn't read due to no NUL, fallback to simple read-all
  if [[ -z "$tags_input" ]]; then
    tags_input=$(cat || true)
  fi
fi

# Get first non-empty line
primary_ref="$(printf '%s' "$tags_input" | awk 'length>0 {print; exit}')"

if [[ -z "$primary_ref" ]]; then
  echo "[detect-primary-image-ref] No tags provided" >&2
  exit 1
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "image_ref=${primary_ref}" >>"$GITHUB_OUTPUT"
else
  echo "image_ref=${primary_ref}"
fi
