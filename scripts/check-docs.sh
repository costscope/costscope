#!/usr/bin/env bash
# Lightweight docs checks: duplicate basenames and DOCUMENTATION_INDEX links
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
docs_dir="$root/docs"
index_file="$docs_dir/DOCUMENTATION_INDEX.md"

echo "Checking for duplicate basenames in $docs_dir..."
# filenames that are expected to appear in multiple subdirs (whitelist)
whitelist=("README.md" "EXTENDED.md" "ADDING_NEW_PROVIDER.md" "CODE_HEALTH_README.md" "focus_conversion.md" "component-focus-conversion.md" "logging_and_metrics.md" "performance_engine.md" "RBAC_CASBIN_MIGRATION.md" "checklist.md")
# BSD find (macOS) does not support -printf; use basename via xargs for portability
dup_raw=$(find "$docs_dir" -type f -print0 | xargs -0 -n1 basename | sort | uniq -c | awk '$1>1 {print $0}' || true)
dup_filtered=""
while IFS= read -r line; do
  [ -z "$line" ] && continue
  # line format: count filename
  name=$(echo "$line" | awk '{print $2}')
  skip=0
  for w in "${whitelist[@]}"; do
    if [ "$w" = "$name" ]; then skip=1; break; fi
  done
  if [ $skip -eq 0 ]; then
    dup_filtered="$dup_filtered\n$line"
  fi
done <<< "$dup_raw"

if [ -n "$(echo "$dup_filtered" | tr -d '[:space:]')" ]; then
  echo "Duplicate basenames found:";
  echo "$dup_filtered";
  exit 2
else
  echo " No duplicate basenames (whitelisted duplicates ignored)"
fi

if [ -f "$index_file" ]; then
  echo "Verifying links in $index_file..."
  # extract markdown inline code spans and filter to likely path-like tokens (contain '/' or end with .md)
   # extract markdown inline code spans only from lines outside fenced code blocks
   # and filter to likely path-like tokens (contain '/' or end with .md)
   # Use awk to skip fenced code blocks and extract all `inline` tokens per line.
  paths=$(awk '
    BEGIN { inf=0 }
    /^```/ { inf = !inf; next }
    !inf {
      while (match($0, /`[^`]+`/)) {
        tok = substr($0, RSTART+1, RLENGTH-2)
        print tok
        $0 = substr($0, RSTART+RLENGTH)
      }
    }
  ' "$index_file" | grep -E '/|\.md$' | grep -v '^../' || true)
  missing=0
  while read -r p; do
    [ -z "$p" ] && continue
    # Normalize and check several candidate locations:
    # - relative to docs/ (docs/<path>)
    # - relative to repo root (<root>/<path>)
    candidate1="$docs_dir/$p"
    candidate2="$root/$p"
    if [ -f "$candidate1" ] || [ -d "$candidate1" ] || [ -f "$candidate2" ] || [ -d "$candidate2" ]; then
      continue
    else
      echo " Missing referenced path: $p"; missing=1
    fi
  done <<< "$paths"
  if [ $missing -ne 0 ]; then echo " DOCUMENTATION_INDEX.md contains missing references"; exit 3; else echo " All DOCUMENTATION_INDEX.md references exist (within docs/)"; fi
else
  echo "No DOCUMENTATION_INDEX.md found, skipping index checks"
fi

echo "Docs checks passed"
