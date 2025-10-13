#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# CostScope SBOM generation script (M14)
# Generates CycloneDX JSON SBOM via syft -> sbom-syft.json (also copies to sbom.json for legacy consumers)
# Validates file size and presence of Go module components. Optionally creates a cosign attestation.

# Configurable env vars:
#   SYFT_VERSION           Pin syft version (default v1.18.0 aligns with Makefile)
#   SBOM_OUTPUT            Output filename (default sbom-syft.json)
#   SBOM_MIN_SIZE          Minimum size in bytes (default 2048)
#   SBOM_FAIL_MISSING      If set (non-empty), fail if file missing / invalid (default behavior already fails)
#   COSIGN_KEY             If set, triggers cosign attest step
#   SBOM_ATTEST_IMAGE      Image ref to attest (default: costscope:latest)
#   COSIGN_EXPERIMENTAL    Forwarded to cosign (default 1 for attest)

SYFT_VERSION="${SYFT_VERSION:-v1.33.0}"
SBOM_OUTPUT="${SBOM_OUTPUT:-sbom-syft.json}"
SBOM_MIN_SIZE="${SBOM_MIN_SIZE:-2048}"
IMAGE_REF="${SBOM_ATTEST_IMAGE:-${image:-costscope:latest}}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

log() { ci::log "$*"; }
err() { ci::warn "$*"; }

ensure_syft() {
  if ! command -v syft >/dev/null 2>&1; then
    log "Installing syft ${SYFT_VERSION} (go install preferred)..."
    # Try to install via `go install` with explicit module path + version. This places the
    # binary in $(go env GOPATH)/bin or GOBIN when supported. If it fails, fall back to the
    # upstream install script (download+exec) to preserve previous behaviour.
    if ! go install github.com/anchore/syft/cmd/syft@${SYFT_VERSION} 2>/dev/null; then
      log "go install failed; falling back to upstream installer"
      # Download the upstream installer and invoke it without passing an extra '-s' which
      # some /bin/sh implementations (notably on non-GNU platforms) may treat as illegal.
      curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh -o /tmp/syft-install.sh && sh /tmp/syft-install.sh -b "$(go env GOPATH)/bin" "${SYFT_VERSION}" >/dev/null
    fi
  fi
}

ensure_cosign() {
  if ! command -v cosign >/dev/null 2>&1; then
    local ver arch
    ver="${COSIGN_VERSION:-v2.2.4}"
    # map uname arch to cosign release arch (support arm64)
    case "$(uname -m)" in
      x86_64|amd64) arch=amd64 ;;
      aarch64|arm64) arch=arm64 ;;
      *) arch=amd64 ;;
    esac
    log "Installing cosign ${ver} for arch=${arch}..."
    curl -sSfL "https://github.com/sigstore/cosign/releases/download/${ver}/cosign-$(uname -s | tr '[:upper:]' '[:lower:]')-${arch}" -o "$(go env GOPATH)/bin/cosign" && chmod +x "$(go env GOPATH)/bin/cosign"
  fi
}

generate_sbom() {
  log " Generating SBOM (syft) -> ${SBOM_OUTPUT}";
  # Prefer cyclonedx-json for richer metadata; fallback to generic JSON spec if older doc expectation.
  syft dir:. -o cyclonedx-json="${SBOM_OUTPUT}" > /dev/null
  cp "${SBOM_OUTPUT}" sbom.json 2>/dev/null || true
}

validate_sbom() {
  if [ ! -f "${SBOM_OUTPUT}" ]; then
    err "SBOM file missing: ${SBOM_OUTPUT}"
    exit 1
  fi
  local size
  size=$(stat -c '%s' "${SBOM_OUTPUT}")
  if [ "${size}" -lt "${SBOM_MIN_SIZE}" ]; then
    err "SBOM size ${size} < minimum ${SBOM_MIN_SIZE} bytes"
    exit 2
  fi
  if command -v jq >/dev/null 2>&1; then
    # Detect at least one Go module component (purl with pkg:golang)
    if ! jq -e '(.components[]? | select(.purl | test("^pkg:golang/"))) | length >= 0' "${SBOM_OUTPUT}" >/dev/null 2>&1; then
      err "No Go module components detected (purl pkg:golang/)"
      exit 3
    fi
  else
    log "️ jq not installed; skipping component validation"
  fi
  log " SBOM validation passed (size=${size} bytes)"
}

maybe_attest() {
  if [ -n "${COSIGN_KEY:-}" ]; then
    ensure_cosign
    export COSIGN_EXPERIMENTAL="${COSIGN_EXPERIMENTAL:-1}"
    log "️  Creating cosign attestation for image ${IMAGE_REF} (predicate=${SBOM_OUTPUT})"
    if ! cosign attest --yes --predicate "${SBOM_OUTPUT}" --type cyclonedx "${IMAGE_REF}"; then
      err "Cosign attestation failed"
      exit 4
    fi
    log " Cosign attestation complete"
  else
    log "ℹ️  COSIGN_KEY not set; skipping attestation"
  fi
}

main() {
  ensure_syft
  generate_sbom
  validate_sbom
  maybe_attest
  log " SBOM generation finished"
}

main "$@"
