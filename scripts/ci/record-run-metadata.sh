#!/usr/bin/env bash
set -euo pipefail

# Write run metadata into docs/security/run-metadata.yaml
mkdir -p docs/security
echo "run_environment: ${GITHUB_ACTIONS:-false}" > docs/security/run-metadata.yaml
if [ -z "${ACTIONS_RUNTIME_TOKEN:-}" ]; then
  echo "note: ACTIONS_RUNTIME_TOKEN not present (likely local/act run)" >> docs/security/run-metadata.yaml
else
  echo "note: ACTIONS_RUNTIME_TOKEN present" >> docs/security/run-metadata.yaml
fi
