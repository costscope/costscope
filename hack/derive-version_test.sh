#!/usr/bin/env bash
set -euo pipefail

# Simple tests for hack/derive-version.sh
SCRIPT="$(dirname "$0")/derive-version.sh"

echo "Running derive-version tests..."

run_case() {
  desc="$1"
  input="$2"
  expect_ok="$3"
  echo -n "$desc: "
  if "${SCRIPT}" "$input" >/dev/null 2>&1; then
    if [ "$expect_ok" = true ]; then
      echo "PASS"
    else
      echo "FAIL (expected fail, but succeeded)"; exit 2
    fi
  else
    if [ "$expect_ok" = true ]; then
      echo "FAIL (expected success, but failed)"; exit 2
    else
      echo "PASS"
    fi
  fi
}

# 1) Branch-like input should fail
run_case "branch master should fail" "master" false

# 2) Dummy tag with prerelease suffix should pass
run_case "tag v0.0.0-test should pass" "v0.0.0-test" true

# 3) Numeric version without 'v' should be normalized and pass
run_case "numeric 0.1.0 should pass and be normalized" "0.1.0" true

echo "All derive-version tests passed."
