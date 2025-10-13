#!/usr/bin/env bash
set -euo pipefail

# Detect nektos/act and propagate flags into $GITHUB_ENV for subsequent steps
# Inputs: environment variables GITHUB_ACTOR, ACT_FULL (optional)

if [[ "${GITHUB_ACTOR:-}" == "nektos/act" ]]; then
  echo "IS_ACT=true" >> "$GITHUB_ENV"
  echo "ACT=true" >> "$GITHUB_ENV"
  echo "Detected local act runtime; will skip heavy network/apt steps"
else
  echo "IS_ACT=false" >> "$GITHUB_ENV"
fi

if [[ "${ACT_FULL:-}" == "true" ]]; then
  echo "ACT_FULL=true" >> "$GITHUB_ENV"
  echo "[act] Full mode enabled (ACT_FULL=true)"
fi
