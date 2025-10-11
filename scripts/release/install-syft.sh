#!/usr/bin/env bash
set -euo pipefail

VER=${SYFT_VERSION:-v1.18.0}
if command -v syft >/dev/null 2>&1; then
  echo "syft already present: $(syft version || true)"
  exit 0
fi

OS=$(uname | tr '[:upper:]' '[:lower:]')

# Try a few common URL variants (some releases use different naming conventions).
try_urls=(
  "https://github.com/anchore/syft/releases/download/${VER}/syft_${VER#v}_${OS}_x86_64.tar.gz"
  "https://github.com/anchore/syft/releases/download/${VER}/syft_${VER}_${OS}_x86_64.tar.gz"
  # Try variant without leading 'v' in path if VER starts with v
  "https://github.com/anchore/syft/releases/download/${VER#v}/syft_${VER#v}_${OS}_x86_64.tar.gz"
)

downloaded=0
for URL in "${try_urls[@]}"; do
  echo "Attempting to download syft from ${URL}"
  if curl -fsSL "$URL" -o /tmp/syft.tar.gz; then
    downloaded=1
    break
  else
    echo "Download failed for ${URL}, trying next variant..."
    rm -f /tmp/syft.tar.gz || true
  fi
done

if [ "$downloaded" -ne 1 ]; then
  echo "Failed to download syft from known URL patterns for ${VER}."
  echo "Please set SYFT_VERSION to a valid release (e.g. v1.22.0) or install syft manually."
  exit 2
fi

tar -xzf /tmp/syft.tar.gz -C /tmp
if [ -f /tmp/syft ]; then
  install -m 0755 /tmp/syft /usr/local/bin/syft
  rm -f /tmp/syft.tar.gz /tmp/syft || true
  echo "syft installed: $(/usr/local/bin/syft version || true)"
else
  echo "Downloaded archive did not contain expected 'syft' binary. Inspect /tmp/syft.tar.gz" >&2
  exit 3
fi
