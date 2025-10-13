#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Local debug script to reproduce the 'Derive version' step from .github/workflows/release.yml
# Usage:
#   RELEASE_RAW='' ./scripts/ci/derive-version-debug.sh
#   RELEASE_RAW='v0.1.0' ./scripts/ci/derive-version-debug.sh
#   RELEASE_RAW='0.1.0' ./scripts/ci/derive-version-debug.sh

RAW_VERSION=${RELEASE_RAW:-}
GITHUB_EVENT_NAME=${GITHUB_EVENT_NAME:-push}

if [[ "$GITHUB_EVENT_NAME" == "workflow_dispatch" ]]; then
  : # In workflow the raw values are taken from inputs; emulate via env
fi

if [[ -z "$RAW_VERSION" ]]; then
  ci::die "No RAW_VERSION provided. Example: RELEASE_RAW=v0.1.0 $0"
fi

ci::log "Testing RAW_VERSION='$RAW_VERSION'"

if ! echo "$RAW_VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  ci::die "Version $RAW_VERSION not valid SemVer (vMAJOR.MINOR.PATCH[-PRERELEASE])"
fi

ci::log "Detected valid version: $RAW_VERSION"

echo "version=$RAW_VERSION"
echo "tag=$RAW_VERSION"
echo "prerelease=false"
