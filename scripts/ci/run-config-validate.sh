#!/usr/bin/env bash
set -euo pipefail

# Purpose: Run config validation with consistent exit code semantics.
# Exit codes:
#   0 - all configs valid
#   2 - validation failures (schema/logic) found (distinct from execution error)
#  >2 - unexpected execution or internal error

exit_code=""

echo "[config-validate] Running config validation..." >&2
if ! go run ./scripts/tools/config-validate --format text; then
  exit_code=$?
fi

if [[ "${exit_code:-0}" == "2" ]]; then
  echo "Validation failures detected" >&2
  exit 2
fi

if [[ -n "${exit_code}" ]]; then
  echo "[config-validate] Exiting with code ${exit_code}" >&2
  exit "${exit_code}"
fi

echo "[config-validate] Success" >&2