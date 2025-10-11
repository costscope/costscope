#!/usr/bin/env bash
set -euo pipefail

# Local debug script to reproduce the 'Derive version' step from .github/workflows/release.yml
# Usage:
#   RELEASE_RAW='' ./scripts/ci/derive-version-debug.sh
#   RELEASE_RAW='v0.1.0' ./scripts/ci/derive-version-debug.sh
#   RELEASE_RAW='0.1.0' ./scripts/ci/derive-version-debug.sh

RAW_VERSION=${RELEASE_RAW:-}
GITHUB_EVENT_NAME=${GITHUB_EVENT_NAME:-push}

if [ "$GITHUB_EVENT_NAME" = "workflow_dispatch" ]; then
  # In workflow the raw values are taken from inputs; emulate via env
  : # no-op
else
  # emulate tag ref (if not provided, use RAW_VERSION as-is)
  # In GH the ref is like refs/tags/v0.1.0 -> script extracts last component
  # Here we assume user provided RAW_VERSION or fallback to empty
  :
fi

if [ -z "$RAW_VERSION" ]; then
  echo "No RAW_VERSION provided. Example: RELEASE_RAW=v0.1.0 $0" >&2
  exit 2
fi

echo "Testing RAW_VERSION='$RAW_VERSION'"

if ! echo "$RAW_VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "Version $RAW_VERSION not valid SemVer (vMAJOR.MINOR.PATCH[-PRERELEASE])" >&2
  exit 1
fi

echo "Detected valid version: $RAW_VERSION"

echo "version=$RAW_VERSION"
echo "tag=$RAW_VERSION"
echo "prerelease=false"
