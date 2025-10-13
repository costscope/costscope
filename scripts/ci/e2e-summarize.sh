#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Summarize E2E results by merging available provider reports and writing a table to $GITHUB_STEP_SUMMARY.
# Expects files in e2e-artifacts/: e2e_report.json (aws), azure_report.json, gcp_report.json (if present)

require_jq() {
  if ! command -v jq >/dev/null 2>&1; then
    if ! ci::is_act && command -v sudo >/dev/null 2>&1; then
      sudo apt-get update || true
      sudo apt-get install -y jq || true
    else
      ci::warn "jq not found; skipping apt install under act or without sudo. Summary will be limited."
    fi
  fi
}

require_jq || true

mkdir -p e2e-artifacts
summary='{}'
add_report() { # args: key file
  if [[ -f "$2" ]] && command -v jq >/dev/null 2>&1; then
    summary=$(jq --arg k "$1" --slurpfile d "$2" '.[$k]=$d[0]' <<<"$summary")
  fi
}

add_report aws e2e-artifacts/e2e_report.json
add_report azure e2e-artifacts/azure_report.json
add_report gcp e2e-artifacts/gcp_report.json

printf '%s\n' "$summary" > e2e-artifacts/summary.json

{
  echo "E2E summary:"
  echo
  echo '```json'
  cat e2e-artifacts/summary.json
  echo '```'
} >> "${GITHUB_STEP_SUMMARY:-/dev/null}" 2>/dev/null || true

# Build overview table and enforce gate
if command -v jq >/dev/null 2>&1 && jq -e '.aws' e2e-artifacts/summary.json >/dev/null 2>&1; then
  {
    echo
    echo '### Drift & Violations'
    printf '| Provider | Passed | Violations | Row Count | EffCost Drift | ListCost Drift | UsageQty Drift |\n'
    printf '|----------|--------|-----------|-----------|--------------|---------------|---------------|\n'
    for prov in aws azure gcp; do
      if jq -e ".$prov" e2e-artifacts/summary.json >/dev/null 2>&1; then
        jq -r --arg p "$prov" '
          .[$p] as $r | [$p, ($r.passed|tostring), (( $r.invariants.violations // [])|length), ($r.aggregates_scan.row_count // 0),
          ($r.relative_drift.sum_effective_cost // 0), ($r.relative_drift.sum_list_cost // 0), ($r.relative_drift.sum_usage_quantity // 0)]
          | "| " + (.[0]) + " | " + (.[1]) + " | " + (.[2]|tostring) + " | " + (.[3]|tostring) + " | " + (.[4]|tostring) + " | " + (.[5]|tostring) + " | " + (.[6]|tostring) + " |"' e2e-artifacts/summary.json
      fi
    done
  } >> "${GITHUB_STEP_SUMMARY:-/dev/null}" 2>/dev/null || true

  # Fail the job if any provider failed or has violations
  if jq 'to_entries | map(select(.value.passed==false or ((.value.invariants.violations // [])|length) > 0)) | length > 0' e2e-artifacts/summary.json | grep -q true; then
    ci::die "One or more E2E provider checks failed or reported violations"
  fi
fi
