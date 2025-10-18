#!/usr/bin/env bash
set -euo pipefail

# Compute version-related variables and export them to GITHUB_ENV and GITHUB_OUTPUT
# Mirrors logic previously embedded in the workflow YAML.

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT=$(git rev-parse --short=12 HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GOVERSION=$(go version | awk '{print $3}')
SOURCE_REPO=$(git config --get remote.origin.url || echo 'https://github.com/your/repo')

# Export to environment for subsequent run steps
if [[ -n "${GITHUB_ENV:-}" ]]; then
  {
    echo "VERSION=$VERSION"
    echo "COMMIT=$COMMIT"
    echo "BUILD_DATE=$BUILD_DATE"
    echo "GOVERSION=$GOVERSION"
    echo "SOURCE_REPO=$SOURCE_REPO"
  } >> "$GITHUB_ENV"
fi

# Expose as step outputs for expression usage when available
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "version=$VERSION"
    echo "commit=$COMMIT"
    echo "build_date=$BUILD_DATE"
    echo "go_version=$GOVERSION"
    echo "source_repo=$SOURCE_REPO"
  } >> "$GITHUB_OUTPUT"
fi

# Also print a short summary for local debugging
printf 'Computed: VERSION=%s COMMIT=%s BUILD_DATE=%s GOVERSION=%s SOURCE_REPO=%s\n' \
  "$VERSION" "$COMMIT" "$BUILD_DATE" "$GOVERSION" "$SOURCE_REPO"
