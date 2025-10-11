#!/usr/bin/env bash
set -euo pipefail

VERSION=${VERSION:-unknown}
{
  echo "### Release Pipeline Summary"
  echo "* Version: ${VERSION}"
  echo "* Draft release created (prerelease=${PRERELEASE:-unknown})"
  echo "* Artifacts: binaries (linux/darwin x amd64/arm64), SBOM, checksums + signature"
} >> "${GITHUB_STEP_SUMMARY:-/dev/null}" 2>/dev/null || true
