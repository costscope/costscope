#!/usr/bin/env bash
set -euo pipefail

# derive-version.sh
# Usage:
#   derive-version.sh <raw-version> [prerelease]
# If raw-version is empty, it will try to read from GITHUB_REF (useful in CI).

RAW_INPUT="${1:-}" 
PRERELEASE_INPUT="${2:-false}"

if [ -z "${RAW_INPUT}" ]; then
  # Try to read from GITHUB_REF (e.g. refs/tags/v0.1.0)
  RAW_INPUT="${GITHUB_REF##*/}"
fi

RAW_VERSION="$RAW_INPUT"
PRERELEASE="$PRERELEASE_INPUT"

# Accept versions with or without leading 'v', normalize to 'vMAJOR.MINOR.PATCH...'
if echo "$RAW_VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+'; then
  NORMALIZED_VERSION="v$RAW_VERSION"
else
  NORMALIZED_VERSION="$RAW_VERSION"
fi

if ! echo "$NORMALIZED_VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "Version $RAW_VERSION (normalized: $NORMALIZED_VERSION) not valid SemVer (vMAJOR.MINOR.PATCH[-PRERELEASE])" >&2
  exit 1
fi

cat <<EOF
version=${NORMALIZED_VERSION}
tag=${NORMALIZED_VERSION}
prerelease=${PRERELEASE}
EOF

exit 0
