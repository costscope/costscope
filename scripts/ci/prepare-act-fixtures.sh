#!/usr/bin/env bash
set -euo pipefail

echo "[act] Preparing smoke fixtures for local act run"
# Prefer the lightweight generator which doesn't depend on the built binary
if [[ -f ./scripts/e2e/generate-fixtures.sh ]]; then
  bash ./scripts/e2e/generate-fixtures.sh
elif [[ -f ./scripts/e2e/run.sh ]]; then
  bash ./scripts/e2e/run.sh
else
  echo "No fixture generator found (generate-fixtures.sh or run.sh)" >&2
  exit 1
fi

mkdir -p tests/fixtures/aws tests/fixtures/azure tests/fixtures/gcp
# Copy generator outputs into the test fixture paths the unit tests expect
cp -f tmp/e2e/data/aws-cur.csv tests/fixtures/aws/cur_smoke.csv
cp -f tmp/e2e/data/azure-cost.csv tests/fixtures/azure/usage_smoke.csv
cp -f tmp/e2e/data/gcp-billing.csv tests/fixtures/gcp/usage_smoke.csv
ls -l tests/fixtures/aws/cur_smoke.csv tests/fixtures/azure/usage_smoke.csv tests/fixtures/gcp/usage_smoke.csv || true
