#!/usr/bin/env bash
set -euo pipefail

# Simple diagnostic to verify that actions used in workflows are present under
# .github/_actions when running under act. This helps catch MODULE_NOT_FOUND
# errors early in local runs.

ROOT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
TARGET_DIR="$ROOT_DIR/.github/_actions"

echo "[act-actions-check] Scanning workflows for pinned actions..."
missing=0

uses_tmp="${TMPDIR:-/tmp}/act-actions-uses-$$.txt"
grep -RhoE "uses:[[:space:]]*actions/[a-zA-Z0-9_-]+@[0-9a-f]{7,40}" "$ROOT_DIR/.github/workflows" 2>/dev/null | sed -E 's/.*uses:[[:space:]]*(actions\/[^@]+)@([0-9a-f]{7,40}).*/\1@\2/' | sort -u > "$uses_tmp" || true

if [[ ! -s "$uses_tmp" ]]; then
  echo "[act-actions-check] No pinned actions with commit SHAs found. Nothing to check."
  exit 0
fi

count=$(wc -l < "$uses_tmp" | tr -d ' ')
printf "[act-actions-check] Checking %d actions under %s\n" "$count" "$TARGET_DIR"
while IFS= read -r spec; do
  repo="${spec%@*}" sha="${spec##*@}"
  name="actions-$(basename "$repo")"
  path="$TARGET_DIR/${name}@${sha}"
  idx="$path/dist/setup/index.js"
  if [[ -d "$path" ]]; then
    # Heuristics: setup-go is Node action with dist/setup entry; others may differ.
    if [[ -f "$idx" || -f "$path/action.yml" || -f "$path/action.yaml" ]]; then
      echo "[OK] $repo@$sha -> $path"
    else
      echo "[WARN] $repo@$sha present but expected dist/action entry missing: $path" >&2
    fi
  else
    echo "[MISSING] $repo@$sha -> $path" >&2
    ((missing++)) || true
  fi
done < "$uses_tmp"

if (( missing > 0 )); then
  echo "[act-actions-check] Missing $missing action(s). Run: scripts/ci/vendor-github-actions.sh" >&2
  exit 2
fi

echo "[act-actions-check] All required actions appear present."
exit 0
