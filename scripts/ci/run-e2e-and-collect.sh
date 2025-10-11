#!/usr/bin/env bash
set -euo pipefail

# Ensure we are at the repository root regardless of caller CWD.
# Under act, GITHUB_WORKSPACE may not point to the synced repo; probe multiple candidates.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

probe_candidates=()
if [[ -n "${GITHUB_WORKSPACE:-}" && -d "${GITHUB_WORKSPACE}" ]]; then
  probe_candidates+=("${GITHUB_WORKSPACE}")
fi
probe_candidates+=("${SCRIPT_ROOT}")
probe_candidates+=("${PWD}")
# Common nested layout when repo root contains a top-level folder named like the module (e.g., costscope)
probe_candidates+=("${SCRIPT_ROOT}/costscope")

choose_root=""
for cand in "${probe_candidates[@]}"; do
  # A valid root contains go.mod and our tests/fixtures directory
  if [[ -f "${cand}/go.mod" && -d "${cand}/tests/fixtures" ]]; then
    choose_root="${cand}"
    break
  fi
done

if [[ -z "${choose_root}" ]]; then
  # Fall back to script root as a best effort
  choose_root="${SCRIPT_ROOT}"
fi

if [[ "${ACT:-}" == "true" ]]; then
  echo "[e2e] PWD=${PWD} GITHUB_WORKSPACE=${GITHUB_WORKSPACE:-} SCRIPT_ROOT=${SCRIPT_ROOT} CHOSEN_ROOT=${choose_root}"
fi

cd "${choose_root}"

mkdir -p e2e-artifacts
echo "Running E2E pipeline tests..."

# Prefer running with duckdb tag (tests are gated by //go:build duckdb)
# Allow override via E2E_TAGS (comma-separated Go build tags)
E2E_TAGS=${E2E_TAGS:-duckdb}

# If running under act and required fixtures are missing, skip gracefully.
if [[ "${ACT:-}" == "true" ]]; then
  missing=()
  [[ -f tests/fixtures/aws/cur_baseline_sample.csv ]] || missing+=("tests/fixtures/aws/cur_baseline_sample.csv")
  [[ -f tests/fixtures/azure/usage.csv ]] || missing+=("tests/fixtures/azure/usage.csv")
  [[ -f tests/fixtures/gcp/usage_minimal.csv ]] || missing+=("tests/fixtures/gcp/usage_minimal.csv")
  if (( ${#missing[@]} > 0 )); then
    echo "E2E fixtures missing under act; skipping E2E. Missing: ${missing[*]}"
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
  echo "E2E pipeline tests skipped (duckdb tag not available in this environment)."
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
  jq -n 'reduce inputs as $i ({}; . * $i)'
fi
