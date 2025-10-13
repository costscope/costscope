#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Write run metadata into docs/security/run-metadata.yaml
mkdir -p docs/security
{
  echo "run_environment: ${GITHUB_ACTIONS:-false}"
  if [[ -z "${ACTIONS_RUNTIME_TOKEN:-}" ]]; then
    echo "note: ACTIONS_RUNTIME_TOKEN not present (likely local/act run)"
  else
    echo "note: ACTIONS_RUNTIME_TOKEN present"
  fi
} > docs/security/run-metadata.yaml
ci::log "Wrote docs/security/run-metadata.yaml"
