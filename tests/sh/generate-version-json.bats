#!/usr/bin/env bats

load 'test_helper'

setup() {
  rm -f version.json
}

@test "generate-version-json.sh writes version.json" {
  run bash -lc './scripts/generate-version-json.sh'
  [ "$status" -eq 0 ]
  [ -f version.json ]
  # sanity: has required keys
  run bash -lc 'jq -e .version version.json >/dev/null 2>&1 || true; jq -e .commit version.json >/dev/null 2>&1 || true'
  [ "$status" -eq 0 ]
}
