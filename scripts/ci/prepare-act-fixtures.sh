#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

ci::log "[act] Preparing smoke fixtures for local act run"
# Prefer the lightweight generator which doesn't depend on the built binary
if [[ -f ./scripts/e2e/generate-fixtures.sh ]]; then
  bash ./scripts/e2e/generate-fixtures.sh
elif [[ -f ./scripts/e2e/run.sh ]]; then
  bash ./scripts/e2e/run.sh
else
  ci::die "No fixture generator found (generate-fixtures.sh or run.sh)"
fi

mkdir -p tests/fixtures/aws tests/fixtures/azure tests/fixtures/gcp
# Copy generator outputs into the test fixture paths the unit tests expect
cp -f tmp/e2e/data/aws-cur.csv tests/fixtures/aws/cur_smoke.csv
cp -f tmp/e2e/data/azure-cost.csv tests/fixtures/azure/usage_smoke.csv
cp -f tmp/e2e/data/gcp-billing.csv tests/fixtures/gcp/usage_smoke.csv
ls -l tests/fixtures/aws/cur_smoke.csv tests/fixtures/azure/usage_smoke.csv tests/fixtures/gcp/usage_smoke.csv || true
