#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# Guard script: ensure every //nolint:unused has an explanatory comment.
# Policy defined in CONTRIBUTING.md (Lint Suppression Policy).

fail=0
while IFS= read -r line; do
  file=${line%%:*}
  rest=${line#*:}
  lineno=${rest%%:*}
  text=${rest#*:}
  # Accept pattern: //nolint:unused // rationale
  if ! grep -Eq "//nolint:unused // .+" <<<"$text"; then
    echo " Missing rationale after //nolint:unused at $file:$lineno" >&2
    fail=1
  fi
done < <(grep -R --line-number --no-color "//nolint:unused" . || true)

if [ "$fail" -ne 0 ]; then
  echo "\nPolicy: Each //nolint:unused requires a trailing explanatory comment (see CONTRIBUTING.md)." >&2
  exit 1
fi

ci::log " nolint:unused rationale guard passed"