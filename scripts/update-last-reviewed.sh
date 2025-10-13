#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# update-last-reviewed.sh
# Bumps the last_reviewed date (YYYY-MM-DD) in YAML front matter for changed docs.
# Usage: scripts/update-last-reviewed.sh [base-ref]
# Default base ref: origin/main

BASE_REF=${1:-origin/main}
today=$(date +%F)

changed=$(git diff --name-only --diff-filter=ACMRT $BASE_REF -- 'docs/**/*.md' 'README.md' | grep -E '\.md$' || true)
[ -z "$changed" ] && { ci::log "No changed markdown docs relative to $BASE_REF"; exit 0; }

ci::log "Updating last_reviewed for: $changed"

for f in $changed; do
  # skip if no front matter
  if ! grep -q '^---$' "$f"; then
  ci::warn "[skip] $f (no front matter)"; continue
  fi
  # only modify if key present or front matter ends without it
  if grep -q '^last_reviewed:' "$f"; then
    sed -i -E "0,/^last_reviewed:.*/s//last_reviewed: $today/" "$f"
  else
    # insert before closing front matter separator (first occurrence after start)
    awk -v d="$today" 'NR==1{print;next} NR>1 && /^---$/ && !done{print "last_reviewed: " d; done=1} {print}' "$f" > "$f.tmp" && mv "$f.tmp" "$f"
  fi
done

ci::log "Done." 
