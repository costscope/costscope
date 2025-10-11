#!/usr/bin/env bash
set -euo pipefail

echo "Linting markdown for trailing whitespace..."
if grep -R --include='*.md' "[[:blank:]]$" docs/ >/dev/null 2>&1; then
  echo "Trailing spaces found in docs (run: git diff -- docs | sed -n '1,200p')"
  exit 1
fi
echo "No trailing spaces found"
