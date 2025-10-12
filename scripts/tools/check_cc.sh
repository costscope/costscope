#!/usr/bin/env bash
# Lightweight cyclomatic complexity guard for staged / provided Go files.
# Fails if any non-test function exceeds threshold (default 25) unless ALLOW_HIGH_COMPLEXITY=1.
# Usage: scripts/tools/check_cc.sh [files...]

set -euo pipefail

THRESHOLD="${CC_THRESHOLD:-25}"
ALLOW="${ALLOW_HIGH_COMPLEXITY:-0}"

# Ensure gocyclo exists
if ! command -v gocyclo >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    echo "[cc-check] gocyclo not installed. Install via: make lint-tools-install" >&2
  else
    echo "[cc-check] gocyclo not installed. Install via: brew install gocyclo (macOS) or make lint-tools-install (Linux)" >&2
  fi
  exit 1
fi

FILES=("$@")
if [ ${#FILES[@]} -eq 0 ]; then
  # Fallback to full repo scan (excluding vendor, bin, _archive)
  mapfile -t FILES < <(git ls-files '*.go' | grep -Ev '(^vendor/|^bin/|^_archive/)' || true)
fi

if [ ${#FILES[@]} -eq 0 ]; then
  echo "[cc-check] No Go files to scan"; exit 0; fi

# Run gocyclo across selected files (filtering to just them)
# gocyclo lacks direct include list, so pipe subset.
TMP_LIST=$(mktemp)
printf '%s\n' "${FILES[@]}" > "$TMP_LIST"

# Run gocyclo; it outputs: <CC> <package> <signature> <path:line:col>
# We'll filter out *_test.go unless explicitly desired.
VIOLATIONS=()
while IFS= read -r line; do
  cc=$(echo "$line" | awk '{print $1}')
  path=$(echo "$line" | awk '{print $NF}') # last token path:line:col
  file=${path%%:*}
  if [[ $file == *"_test.go" ]]; then
    continue
  fi
  if (( cc > THRESHOLD )); then
    VIOLATIONS+=("$line")
  fi
done < <(gocyclo -over "$THRESHOLD" $(cat "$TMP_LIST") || true)

rm -f "$TMP_LIST"

if [ ${#VIOLATIONS[@]} -gt 0 ]; then
  echo "[cc-check] Functions exceeding cyclomatic complexity threshold ($THRESHOLD):"
  printf '  %s\n' "${VIOLATIONS[@]}"
  if [ "$ALLOW" != "1" ]; then
    echo "[cc-check] FAIL (set ALLOW_HIGH_COMPLEXITY=1 to bypass temporarily)" >&2
    exit 1
  else
    echo "[cc-check] WARN bypass enabled (ALLOW_HIGH_COMPLEXITY=1)"
  fi
else
  echo "[cc-check] OK (no functions > $THRESHOLD)"
fi
