#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Detect nektos/act and propagate flags into $GITHUB_ENV for subsequent steps
# Inputs: environment variables GITHUB_ACTOR, ACT_FULL (optional)
ENV_FILE="${GITHUB_ENV:-/dev/null}"

# Track detection in a local variable so we can also export as step outputs reliably
DETECTED_IS_ACT="false"

# Primary signal: many act runners set ACT=true for all steps; honor this first
if [[ "${ACT:-}" == "true" ]]; then
  DETECTED_IS_ACT="true"
  {
    echo "IS_ACT=true"
    echo "ACT=true"
  } >> "$ENV_FILE"
  ci::log "[act-detect] Detected ACT=true in environment; enabling IS_ACT=true"
fi

if [[ "${IS_ACT:-}" == "true" ]]; then
  # Preserve an explicitly injected IS_ACT=true (e.g. passed via act runner -e IS_ACT=true)
  DETECTED_IS_ACT="true"
  ci::log "[act-detect] Preserving pre-set IS_ACT=true from environment"
else
  if [[ "${GITHUB_ACTOR:-}" == "nektos/act" ]]; then
    DETECTED_IS_ACT="true"
    {
      echo "IS_ACT=true"
      echo "ACT=true"
    } >> "$ENV_FILE"
    ci::log "Detected local act runtime (actor match); enabling IS_ACT=true"
  else
    # Fallback heuristics: allow forcing act mode when running locally using dev-local tag or explicit FORCE_ACT
    if [[ "${FORCE_ACT:-}" == "true" || "${CI_IMAGE_TAG:-}" == "dev-local" ]]; then
      DETECTED_IS_ACT="true"
      echo "IS_ACT=true" >> "$ENV_FILE"
      echo "ACT=true" >> "$ENV_FILE"
      ci::log "[act-detect] FORCE_ACT or dev-local CI_IMAGE_TAG detected; setting IS_ACT=true"
    elif [[ -z "${ACTIONS_RUNTIME_TOKEN:-}" ]]; then
      # On real GitHub-hosted runners, ACTIONS_RUNTIME_TOKEN is always present.
      # Under act emulation this env var is typically absent; use that as an additional robust signal.
      DETECTED_IS_ACT="true"
      echo "IS_ACT=true" >> "$ENV_FILE"
      echo "ACT=true" >> "$ENV_FILE"
      ci::log "[act-detect] ACTIONS_RUNTIME_TOKEN is missing; assuming act and setting IS_ACT=true"
    else
      echo "IS_ACT=false" >> "$ENV_FILE"
    fi
  fi
fi

if [[ "${ACT_FULL:-}" == "true" ]]; then
  echo "ACT_FULL=true" >> "$ENV_FILE"
  ci::log "[act] Full mode enabled (ACT_FULL=true)"
fi

# Also expose detection as step outputs when running inside GitHub Actions
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "is_act=${DETECTED_IS_ACT}"
    echo "act_full=${ACT_FULL:-false}"
  } >> "${GITHUB_OUTPUT}"
fi
