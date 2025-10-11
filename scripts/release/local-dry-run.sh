#!/usr/bin/env bash
set -euo pipefail

# Local dry-run helper for release pipeline.
# Usage: scripts/release/local-dry-run.sh v0.1.0-rc1

VERSION=${1:-}
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version-tag (e.g. v0.1.0-rc1)>" >&2
  exit 1
fi

if ! echo "$VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "Invalid SemVer tag: $VERSION" >&2
  exit 1
fi

if ! command -v git-cliff >/dev/null 2>&1; then
  echo "Installing git-cliff (temporary via Homebrew if available)..." >&2
  if command -v brew >/dev/null 2>&1; then
    brew install git-cliff || true
  else
    echo "Homebrew not found; skipping auto-install of git-cliff (notes generation will be skipped)." >&2
  fi
fi

if command -v git-cliff >/dev/null 2>&1; then
  echo " Generating release notes preview for $VERSION..."
  git-cliff --config .git-cliff.toml --tag "$VERSION" > /tmp/RELEASE_NOTES_PREVIEW.md
  head -n 40 /tmp/RELEASE_NOTES_PREVIEW.md || true
  echo "... (see /tmp/RELEASE_NOTES_PREVIEW.md for full notes)"
else
  echo "(skip) git-cliff not installed; skipping release notes preview." >&2
fi

echo " Building current-platform binary (dry-run)"
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) CGO_ENABLED=0 go build \
  -ldflags "-s -w -X main.version=$VERSION -extldflags '-Wl,-dead_strip -Wl,-x'" \
  -trimpath \
  -o "costscope-$VERSION-$(go env GOOS)-$(go env GOARCH)" .

if command -v sha256sum >/dev/null 2>&1; then
  echo " Calculating checksum"
  sha256sum "costscope-$VERSION-$(go env GOOS)-$(go env GOARCH)" > /tmp/checksums.txt
  cat /tmp/checksums.txt
else
  shasum -a 256 "costscope-$VERSION-$(go env GOOS)-$(go env GOARCH)" > /tmp/checksums.txt
  cat /tmp/checksums.txt
fi

echo " SBOM (syft)"
if command -v syft >/dev/null 2>&1; then
  syft dir:. -o cyclonedx-json=/tmp/sbom.json >/dev/null
  echo "SBOM written to /tmp/sbom.json"
else
  echo "(skip) syft not installed"
fi

echo " Dry-run complete (no tags pushed)."
