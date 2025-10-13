#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=ci/lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/ci/lib/common.sh"

ci::log "Running basic local markdown link check..."
broken=0

while IFS= read -r -d '' f; do
  # extract Markdown links like [text](path) ignoring external URLs
  while IFS= read -r link; do
    case "$link" in
      http*|mailto:)
        continue
        ;;
    esac
    [[ -z "$link" ]] && continue
    if [[ -f "$link" ]] || [[ -f "$(dirname "$f")/$link" ]]; then
      continue
    fi
    echo "Broken link: $f -> $link"
    broken=1
  done < <(perl -ne 'while(/\[[^]]+\]\(([^)#]+)(?:#[^)]+)?\)/g){print "$1\n"}' "$f")
done < <(git ls-files -z -- 'docs/*.md')

if [[ "$broken" -eq 1 ]]; then
  ci::die "Link check failed"
fi
ci::log "All local links OK"
