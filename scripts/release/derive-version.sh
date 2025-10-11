#!/usr/bin/env bash
set -euo pipefail

# Derive version/tag/prerelease and write to $GITHUB_OUTPUT
# Inputs (via env):
#   GITHUB_EVENT_NAME, GITHUB_REF
#   INPUT_VERSION (when workflow_dispatch)
#   INPUT_PRERELEASE (when workflow_dispatch)

RAW_VERSION=""
PRERELEASE="false"

if [[ "${GITHUB_EVENT_NAME:-}" == "workflow_dispatch" ]]; then
  RAW_VERSION=${INPUT_VERSION:-}
  PRERELEASE=${INPUT_PRERELEASE:-false}
  if [[ -z "${RAW_VERSION}" ]]; then
    echo "When using workflow_dispatch you must provide the 'version' input (e.g. v0.1.0)" >&2
    exit 1
  fi
else
  RAW_VERSION="${GITHUB_REF##*/}"
  PRERELEASE="false"
fi

# Normalize: accept with or without leading 'v'
if [[ "${RAW_VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  NORMALIZED_VERSION="v${RAW_VERSION}"
else
  NORMALIZED_VERSION="${RAW_VERSION}"
fi

if ! [[ "${NORMALIZED_VERSION}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "Version ${RAW_VERSION} (normalized: ${NORMALIZED_VERSION}) not valid SemVer (vMAJOR.MINOR.PATCH[-PRERELEASE])" >&2
  exit 1
fi

{
  echo "version=${NORMALIZED_VERSION}"
  echo "tag=${NORMALIZED_VERSION}"
  echo "prerelease=${PRERELEASE}"
} >> "${GITHUB_OUTPUT}"

echo "Detected version: ${NORMALIZED_VERSION} (prerelease=${PRERELEASE})"
