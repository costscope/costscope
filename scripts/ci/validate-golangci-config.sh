#!/usr/bin/env bash
set -euo pipefail

if ! command -v yq >/dev/null 2>&1; then
  echo "yq is required to validate .golangci.yml. Install it first (e.g., brew install yq)." >&2
  exit 2
fi

echo "Validating .golangci.yml"
if ! yq eval '.' .golangci.yml >/dev/null 2>&1; then
  echo "Invalid YAML in .golangci.yml" >&2
  # Print parse errors to help debugging but do not mask the non-zero exit
  yq eval '.' .golangci.yml || true
  exit 1
fi
echo ".golangci.yml is valid YAML"
