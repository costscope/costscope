#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

ci::log "[act] Preparing smoke fixtures for local act run"
# Allow overriding row counts for faster local runs.
export COSTSCOPE_FIXTURE_MAX_ROWS="${COSTSCOPE_FIXTURE_MAX_ROWS:-200}" # default small sample
ci::log "[act] Using COSTSCOPE_FIXTURE_MAX_ROWS=${COSTSCOPE_FIXTURE_MAX_ROWS}"
# Prefer the lightweight generator which doesn't depend on the built binary
if [[ -f ./scripts/e2e/generate-fixtures.sh ]]; then
  bash ./scripts/e2e/generate-fixtures.sh "${COSTSCOPE_FIXTURE_MAX_ROWS}"
elif [[ -f ./scripts/e2e/run.sh ]]; then
  bash ./scripts/e2e/run.sh
else
  ci::die "No fixture generator found (generate-fixtures.sh or run.sh)"
fi

mkdir -p tests/fixtures/aws tests/fixtures/azure tests/fixtures/gcp
# Copy generator outputs into the test fixture paths the unit tests expect
cp -f tmp/e2e/data/aws-cur.csv tests/fixtures/aws/cur_smoke.csv
cp -f tmp/e2e/data/aws-cur-savingsplan-covered.csv tests/fixtures/aws/cur_savingsplan_covered_usage.csv || true
cp -f tmp/e2e/data/azure-cost.csv tests/fixtures/azure/usage_smoke.csv
cp -f tmp/e2e/data/gcp-billing.csv tests/fixtures/gcp/usage_smoke.csv
ls -l tests/fixtures/aws/cur_smoke.csv tests/fixtures/aws/cur_savingsplan_covered_usage.csv tests/fixtures/azure/usage_smoke.csv tests/fixtures/gcp/usage_smoke.csv || true
