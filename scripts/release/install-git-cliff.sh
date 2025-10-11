#!/usr/bin/env bash
set -euo pipefail

VER=${GITCLIFF_VERSION:-v0.10.0}
if command -v git-cliff >/dev/null 2>&1; then
  echo "git-cliff already present: $(git-cliff --version || true)"
  exit 0
fi

if command -v cargo >/dev/null 2>&1; then
  cargo install git-cliff || true
  exit 0
fi

OS=$(uname | tr '[:upper:]' '[:lower:]')
URL="https://github.com/orhun/git-cliff/releases/download/${VER}/git-cliff-${VER#v}-${OS}-x86_64.tar.gz"
echo "Downloading git-cliff from ${URL}"
curl -sSfL "$URL" -o /tmp/git-cliff.tar.gz
tar -xzf /tmp/git-cliff.tar.gz -C /tmp
install -m 0755 /tmp/git-cliff /usr/local/bin/git-cliff || true
rm -f /tmp/git-cliff.tar.gz /tmp/git-cliff || true
