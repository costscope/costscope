#!/usr/bin/env bash
set -euo pipefail

# Generate checksums and sign + verify with cosign keyless

cosign_check() {
  if ! command -v cosign >/dev/null 2>&1; then
    echo "cosign not found (ensure action installed before calling this script)" >&2
    exit 2
  fi
}

cosign_check

(cd dist && find . -maxdepth 3 -type f -name 'costscope-*' -exec sha256sum {} \;) > checksums.txt
cat checksums.txt

export COSIGN_EXPERIMENTAL=1
cosign sign-blob --yes --output-signature checksums.txt.sig checksums.txt
cosign verify-blob --signature checksums.txt.sig checksums.txt || { echo 'Signature verification failed'; exit 1; }

# Optional provenance attestation on checksum file
cosign attest --predicate <(echo '{"buildType":"binary-aggregate","version":"'"${VERSION:-unknown}"'"}') --type slsaprovenance --yes checksums.txt || true
