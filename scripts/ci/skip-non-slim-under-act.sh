#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

variant="${1:-}"
if [[ -z "$variant" ]]; then
  ci::die "usage: $0 <variant>"
fi

# Guard: when running under nektos/act only run the slim variant locally (unless ACT_FULL=true)
if ci::is_act && [[ "$variant" != "slim" && "${ACT_FULL:-false}" != "true" ]]; then
  ci::log "[act] Skipping variant '$variant' because IS_ACT=true"
  exit 0
fi
