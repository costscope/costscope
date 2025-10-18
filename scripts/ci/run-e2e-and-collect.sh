#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

ROOT="$(ci::repo_root "${SCRIPT_DIR}/../..")"
ci::debug "PWD=${PWD} GITHUB_WORKSPACE=${GITHUB_WORKSPACE:-} SCRIPT_DIR=${SCRIPT_DIR} ROOT=${ROOT}"
cd "${ROOT}"

mkdir -p e2e-artifacts
ci::log "Running E2E pipeline tests..."

# Prefer running with duckdb tag (tests are gated by //go:build duckdb)
# Allow override via E2E_TAGS (comma-separated Go build tags)
E2E_TAGS=${E2E_TAGS:-duckdb}

# If running under act and required fixtures are missing, skip gracefully.
if ci::is_act; then
  missing=()
  [[ -f tests/fixtures/aws/cur_baseline_sample.csv ]] || missing+=("tests/fixtures/aws/cur_baseline_sample.csv")
  [[ -f tests/fixtures/azure/usage.csv ]] || missing+=("tests/fixtures/azure/usage.csv")
  [[ -f tests/fixtures/gcp/usage_minimal.csv ]] || missing+=("tests/fixtures/gcp/usage_minimal.csv")
  if (( ${#missing[@]} > 0 )); then
    ci::warn "E2E fixtures missing under act; skipping E2E. Missing: ${missing[*]}"
    echo '{}' > e2e-artifacts/summary.json || true
    exit 0
  fi
fi

set +e
if [[ -n "$E2E_TAGS" ]]; then
  go test -tags "$E2E_TAGS" ./tests/e2e/pipeline -count=1 -run '^Test.*' -v | tee e2e_all.log
  status=${PIPESTATUS[0]}
else
  go test ./tests/e2e/pipeline -count=1 -run '^Test.*' -v | tee e2e_all.log
  status=${PIPESTATUS[0]}
fi
set -e

# If the package has no files due to build tags, treat as skip (useful under act/local without duckdb)
if [[ $status -ne 0 ]] && grep -q "build constraints exclude all Go files" e2e_all.log; then
  ci::log "E2E pipeline tests skipped (duckdb tag not available in this environment)."
  echo '{}' > e2e-artifacts/summary.json || true
  exit 0
fi

if [[ $status -ne 0 ]]; then
  exit $status
fi

# Extract report paths from logs and copy into artifacts directory with canonical names
grep -ho 'E2E_[A-Z_]*REPORT[^=]*=.*' e2e_all.log || true
while IFS= read -r line; do
  key="${line%%=*}"; path="${line#*=}"; norm="$(echo "$key" | tr 'A-Z_' 'a-z-' )"
  case "$key" in
    E2E_REPORT_PATH) dest="e2e_report.json" ;;
    E2E_AZURE_REPORT) dest="azure_report.json" ;;
    E2E_GCP_REPORT) dest="gcp_report.json" ;;
    *) dest="${norm}.json" ;;
  esac
  if [[ -f "$path" ]]; then cp "$path" "e2e-artifacts/$dest"; fi
done < <(grep -ho 'E2E_[A-Z_]*REPORT[^=]*=.*' e2e_all.log || true)
ls -l e2e-artifacts || true

# Build consolidated summary (requires jq)
if command -v jq >/dev/null 2>&1; then
  # Merge any provider reports present into a single JSON
  inputs=()
  [[ -f e2e-artifacts/e2e_report.json ]] && inputs+=(e2e-artifacts/e2e_report.json)
  [[ -f e2e-artifacts/azure_report.json ]] && inputs+=(e2e-artifacts/azure_report.json)
  [[ -f e2e-artifacts/gcp_report.json ]] && inputs+=(e2e-artifacts/gcp_report.json)
  if ((${#inputs[@]})); then
    jq -s 'reduce .[] as $i ({}; . * $i)' "${inputs[@]}" > e2e-artifacts/summary.json || true
  else
    echo '{}' > e2e-artifacts/summary.json
  fi
else
  # No jq available; create an empty summary to avoid downstream failures
  echo '{}' > e2e-artifacts/summary.json
fi
