#!/usr/bin/env bash
set -euo pipefail

echo "Running basic local markdown link check..."
broken=0
while IFS= read -r f; do
  # extract Markdown links like [text](path) ignoring external URLs
  perl -ne 'while(/\[[^]]+\]\(([^)#]+)(?:#[^)]+)?\)/g){print "$1\n"}' "$f" | while IFS= read -r link; do
    case "$link" in
      http*|mailto:)
        continue
        ;;
    esac
    [ -z "$link" ] && continue
    if [ -f "$link" ]; then
      continue
    fi
    if [ -f "$(dirname "$f")/$link" ]; then
      continue
    fi
    echo "Broken link: $f -> $link"
    broken=1
  done
done < <(git ls-files '*.md' | grep '^docs/' || true)

if [ "$broken" -eq 1 ]; then
  echo "Link check failed"
  exit 1
fi
echo "All local links OK"
