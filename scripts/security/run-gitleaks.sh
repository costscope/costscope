#!/usr/bin/env bash
set -euo pipefail

# Run gitleaks with repo config and redact enabled
# Env:
#   GITLEAKS_CONFIG (default: .gitleaks.toml if exists)
#   GITLEAKS_SOURCE (default: .)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ci::require_cmd gitleaks

SOURCE=${GITLEAKS_SOURCE:-.}
CONFIG_FLAG=()
if [[ -f "${GITLEAKS_CONFIG:-.gitleaks.toml}" ]]; then
  CONFIG_FLAG=(--config "${GITLEAKS_CONFIG:-.gitleaks.toml}")
fi

ci::log "Running gitleaks on source=$SOURCE"
gitleaks detect \
  --source "$SOURCE" \
  --no-banner \
  --report-format json \
  --report-path gitleaks-report.json \
  --redact \
  "${CONFIG_FLAG[@]}" || true
ci::log "gitleaks complete -> gitleaks-report.json"
