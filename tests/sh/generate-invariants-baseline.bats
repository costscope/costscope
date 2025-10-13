#!/usr/bin/env bats

load 'test_helper'

@test "generate-invariants-baseline.sh fails clearly when binary missing" {
  run bash -lc 'tmp=$(mktemp -d); ./scripts/generate-invariants-baseline.sh "$tmp/in.parquet" "$tmp/out.json" 2>"$tmp/err"; echo $?; cat "$tmp/err"'
  [ "$status" -ne 0 ]
}
