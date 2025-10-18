#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

variant="${1:-}"
if [[ -z "$variant" ]]; then
  ci::die "usage: $0 <variant>"
fi

# Allow FORCE_ACT to coerce act-mode skipping logic if ci::is_act() wouldn't detect (e.g. actor mismatch)
if [[ "${FORCE_ACT:-false}" == "true" && "${IS_ACT:-false}" != "true" ]]; then
  IS_ACT=true
fi

# Guard: when running under (real or forced) act only run the slim variant locally.
# Previous logic allowed ACT_FULL to enable all variants, but act runner injected ACT_FULL=true by default,
# defeating slimming. We now require explicit ACT_ALLOW_NON_SLIM=true to run additional variants.
if ci::is_act && [[ "$variant" != "slim" && "${ACT_ALLOW_NON_SLIM:-false}" != "true" ]]; then
  # Ensure smoke fixtures exist even for skipped variants so any tests that
  # might still enumerate fixture paths (e.g. listing, discovery in other
  # packages) do not fail with missing file errors.
  if [[ -f ./scripts/ci/prepare-act-fixtures.sh ]]; then
    ci::log "[act] Preparing fixtures before skipping non-slim variant '$variant'"
    bash ./scripts/ci/prepare-act-fixtures.sh || ci::warn "fixture prep failed (non-fatal)"
  fi
  ci::log "[act] Skipping variant '$variant' because IS_ACT=true and ACT_ALLOW_NON_SLIM!=true"
  # Expose a step output (when invoked from a workflow step with id) plus env var to allow later
  # steps / actions gating without relying on expression re-evaluation of $GITHUB_ENV under act.
  if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    echo "skipped=true" >> "${GITHUB_OUTPUT}"
  fi
  if [[ -n "${GITHUB_ENV:-}" ]]; then
    echo "VARIANT_SKIPPED=true" >> "${GITHUB_ENV}"
  fi
  exit 0
fi

# Additional hard safety: if we somehow reach here under act with a non-slim variant (e.g., workflow conditional
# evaluated before IS_ACT was exported) abort early to avoid wasting resources. This protects against timing/order
# issues in environment propagation under act.
if [[ "${IS_ACT:-false}" == "true" && "$variant" != "slim" && "${ACT_ALLOW_NON_SLIM:-false}" != "true" ]]; then
  ci::warn "[act][safety] Detected non-slim variant '$variant' executing after primary skip gate; forcing early exit (ACT_ALLOW_NON_SLIM not true)"
  if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    echo "skipped=true" >> "${GITHUB_OUTPUT}"
  fi
  if [[ -n "${GITHUB_ENV:-}" ]]; then
    echo "VARIANT_SKIPPED=true" >> "${GITHUB_ENV}"
  fi
  exit 0
fi

# If we reach here variant was not skipped; emit skipped=false for consistency when step id used.
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "skipped=false" >> "${GITHUB_OUTPUT}"
fi
