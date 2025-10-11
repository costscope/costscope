#!/usr/bin/env bash
set -euo pipefail
# Guard: ensure handlers do not use removed deprecated helpers or raw c.JSON for 400/201/404/204 that bypass envelope.
# Patterns flagged:
# - response.Created201(
# - response.BadRequest(
# - c.JSON(400,
# - c.JSON(201,
# - c.JSON(404,
# - c.JSON(204,
# Allow explicit exceptions via inline comment // response-ignore-lint

fail=0
while IFS= read -r file; do
  while IFS= read -r line; do
    lineno=$(cut -d: -f2 <<<"$line")
    text=$(cut -d: -f3- <<<"$line")
    if grep -q "response-ignore-lint" <<<"$text"; then
      continue
    fi
    echo " Response helper guard: disallowed pattern in $line" >&2
    fail=1
  done < <(grep -nE 'response\.Created201\(|response\.BadRequest\(|c\.JSON\((400|201|404|204),' "$file" || true)
 done < <(git ls-files '*.go' | grep -v '^internal/api/response/' | grep -E 'internal/api/(handlers|focus|middleware|websocket|jobs)/' || true)

if [ $fail -eq 1 ]; then
  echo "\nFailing. Replace with AutoCreated201, AutoBadRequest/AutoBadRequestCode, AutoNotFound404, AutoNoContent204, or AutoOK/AutoFail wrappers." >&2
  exit 1
fi

echo " API response helper guard passed"
