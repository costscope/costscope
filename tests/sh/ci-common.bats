#!/usr/bin/env bats

load 'test_helper'

@test "ci::repo_root returns a directory" {
  run bash -lc 'source ./scripts/ci/lib/common.sh; d=$(ci::repo_root "$(pwd)"); [[ -d "$d" ]]'
  [ "$status" -eq 0 ]
}

@test "ci::is_act false by default" {
  run bash -lc 'source ./scripts/ci/lib/common.sh; ci::is_act'
  [ "$status" -ne 0 ]
}
