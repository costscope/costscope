#!/usr/bin/env bash
set -euo pipefail

# Run gitleaks with repo config and redact enabled
# Env:
#   GITLEAKS_CONFIG (default: .gitleaks.toml if exists)
#   GITLEAKS_SOURCE (default: .)

SOURCE=${GITLEAKS_SOURCE:-.}
CONFIG_FLAG=()
if [[ -f "${GITLEAKS_CONFIG:-.gitleaks.toml}" ]]; then
  CONFIG_FLAG=(--config "${GITLEAKS_CONFIG:-.gitleaks.toml}")
fi

gitleaks detect \
  --source "$SOURCE" \
  --no-banner \
  --report-format json \
  --report-path gitleaks-report.json \
  --redact \
  "${CONFIG_FLAG[@]}" || true
