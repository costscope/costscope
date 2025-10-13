#!/usr/bin/env bats

load 'test_helper'

@test "generate-enterprise-stubs.sh creates stub and enterprise files" {
  run bash -lc '
    set -euo pipefail
    pkg=$(mktemp -d)
    mkdir -p "$pkg"
    ./scripts/generate-enterprise-stubs.sh StreamingEngine "$pkg" EnterpriseStreamingEngine >/dev/null
    ls -1 "$pkg" | wc -l
  '
  [ "$status" -eq 0 ]
  # At least two files expected
  [ "$output" -ge 2 ]
}
