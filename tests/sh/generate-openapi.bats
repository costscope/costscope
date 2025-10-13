#!/usr/bin/env bats

load 'test_helper'

@test "generate-openapi.sh exits non-zero with deprecation message" {
  run bash -lc './scripts/generate-openapi.sh'
  [ "$status" -ne 0 ]
  [[ "$output" =~ "removed" ]]
}
