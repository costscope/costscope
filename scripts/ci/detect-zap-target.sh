#!/usr/bin/env bash
set -euo pipefail

# detect-zap-target.sh
# Decides whether OWASP ZAP baseline should run and emits outputs for GitHub Actions.
# Inputs via env:
#   ZAP_TARGET_URL (optional): if non-empty, enables DAST and sets target
# Outputs (GITHUB_OUTPUT):
#   enabled=true|false
#   target=<url> (when enabled)

out_file="${GITHUB_OUTPUT:-}"
emit() {
  if [[ -n "$out_file" ]]; then
    echo "$1" >>"$out_file"
  else
    echo "$1"
  fi
}

target="${ZAP_TARGET_URL:-}"
if [[ -n "$target" ]]; then
  emit "enabled=true"
  emit "target=${target}"
else
  emit "enabled=false"
fi
