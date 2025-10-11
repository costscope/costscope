#!/usr/bin/env bash
set -euo pipefail
mkdir -p govuln_reports
# pick packages likely to touch application code (avoid tooling libs)
go list ./... | grep -E '(^local/costscope/internal|^local/costscope/cmd|^local/costscope/examples|^local/costscope/internal/api|^local/costscope/internal/core)' > /tmp/pkgs_to_scan.txt
while read -r pkg; do
  safe=$(echo "$pkg" | sed 's;/;__;g')
  out="govuln_reports/${safe}.json"
  if [ -s "$out" ]; then
    echo "skip $pkg (exists)"
    continue
  fi
  echo "scan $pkg -> $out"
  # timeout 10 minutes to avoid long hangs / OOM in container
  timeout 10m govulncheck -json "$pkg" > "$out" 2>&1 || echo "govulncheck exit $?: $pkg"
done < /tmp/pkgs_to_scan.txt
