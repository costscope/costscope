#!/usr/bin/env bash
set -euo pipefail

missing=0
for f in README.md docs/ARCHITECTURE.md docs/api/index.md; do
  if [ ! -f "$f" ]; then
    echo "Missing $f"
    missing=1
  fi
done
if [ "$missing" -eq 1 ]; then
  echo "Missing required docs"
  exit 1
fi
echo "Core docs present"
