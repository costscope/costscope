#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Detect nektos/act and propagate flags into $GITHUB_ENV for subsequent steps
# Inputs: environment variables GITHUB_ACTOR, ACT_FULL (optional)
ENV_FILE="${GITHUB_ENV:-/dev/null}"

if [[ "${GITHUB_ACTOR:-}" == "nektos/act" ]]; then
  {
    echo "IS_ACT=true"
    echo "ACT=true"
  } >> "$ENV_FILE"
  ci::log "Detected local act runtime; will skip heavy network/apt steps"
else
  echo "IS_ACT=false" >> "$ENV_FILE"
fi

if [[ "${ACT_FULL:-}" == "true" ]]; then
  echo "ACT_FULL=true" >> "$ENV_FILE"
  ci::log "[act] Full mode enabled (ACT_FULL=true)"
fi
