#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=ci/lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/ci/lib/common.sh"

mkdir -p govuln_reports
ci::log "Selecting packages to scan with govulncheck"
# pick packages likely to touch application code (avoid tooling libs)
go list ./... | grep -E '(^github.com/costscope/costscope/internal|^github.com/costscope/costscope/cmd|^github.com/costscope/costscope/examples|^github.com/costscope/costscope/internal/api|^github.com/costscope/costscope/internal/core)' > /tmp/pkgs_to_scan.txt
while read -r pkg; do
  safe=$(echo "$pkg" | sed 's;/;__;g')
  out="govuln_reports/${safe}.json"
  if [[ -s "$out" ]]; then
    ci::log "skip $pkg (exists)"
    continue
  fi
  ci::log "scan $pkg -> $out"
  # timeout 10 minutes to avoid long hangs / OOM in container
  timeout 10m govulncheck -json "$pkg" > "$out" 2>&1 || ci::warn "govulncheck exit $?: $pkg"
done < /tmp/pkgs_to_scan.txt
