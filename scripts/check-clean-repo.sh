#!/usr/bin/env bash
set -euo pipefail

if [[ ${TRACE:-} == 1 ]]; then set -x; fi

status="$(git status --porcelain)"
if [[ -n "$status" ]]; then
  echo " Repository is dirty. Commit or stash changes before release." >&2
  git status --short >&2 || true
  exit 1
fi
echo " Repository clean"
