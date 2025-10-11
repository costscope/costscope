#!/usr/bin/env bash
set -euo pipefail

# pin-actions.sh
# Resolves action refs (owner/repo@tag) in .github/workflows/*.yml to immutable commit SHAs in-place.
#
# Requirements:
# - bash, sed, awk, grep, curl OR git (ls-remote) recommended for robust tag resolution
# - jq optional (for nicer output)
#
# Usage:
#   scripts/tools/pin-actions.sh [--dry-run]
#
# Behavior:
# - Scans all workflow files for lines like: `uses: owner/repo@ref`
# - Skips lines already pinned to a 40-char SHA
# - Resolves tags to commit SHAs via `git ls-remote https://github.com/owner/repo.git refs/tags/<tag>`.
#   If an annotated tag is used, prefers the dereferenced commit `^{} `.
# - Rewrites the file, replacing @<ref> with @<sha>.
# - Emits a summary of changes.

DRY_RUN=0
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=1
fi

ROOT_DIR=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$ROOT_DIR"

changed=0
# initialize changes safely so `set -u` doesn't cause an unbound variable error
# prefer a simple empty string or array depending on shell; an empty string is
# portable but we use a bash array when bash is available. Keep code that uses
# the array compatible with bash; workflows should execute this script with
# bash (shebang is present), but initializing prevents failures when /bin/sh
# is used accidentally.
changes=()

resolve_sha_for_ref() {
  local owner_repo="$1" tagref="$2"

  # If it's already a full SHA, return as-is
  if [[ "$tagref" =~ ^[0-9a-f]{40}$ ]]; then
    echo "$tagref"
    return 0
  fi

  # Try to resolve via git ls-remote (works without auth for public repos)
  local url="https://github.com/${owner_repo%.git}.git"
  # ls-remote can output two lines for annotated tags; prefer the ^{} deref
  if out=$(git ls-remote "$url" "refs/tags/$tagref" 2>/dev/null); then
    # Prefer dereferenced annotated tag line
    local deref
    deref=$(echo "$out" | awk '/\^\{\}$/ {print $1}' | head -n1)
    if [[ -n "$deref" ]]; then
      echo "$deref"
      return 0
    fi
    # Fallback to the first direct match
    local direct
    direct=$(echo "$out" | awk '{print $1}' | head -n1)
    if [[ -n "$direct" ]]; then
      echo "$direct"
      return 0
    fi
  fi

  # As a fallback, try heads (some actions might use branch names, though discouraged)
  if out=$(git ls-remote "$url" "refs/heads/$tagref" 2>/dev/null); then
    local headsha
    headsha=$(echo "$out" | awk '{print $1}' | head -n1)
    if [[ -n "$headsha" ]]; then
      echo "$headsha"
      return 0
    fi
  fi

  return 1
}

pin_file() {
  local file="$1"
  local tmp="${file}.tmp$$"
  local did_change=0

  while IFS= read -r line || [[ -n "$line" ]]; do
    # Match either 'uses:' or '- uses:' with arbitrary indentation
    if [[ "$line" =~ ^[[:space:]]*[-]?[[:space:]]*uses:[[:space:]]*([^@[:space:]]+)[@]([^[:space:]]+) ]]; then
      local full="${BASH_REMATCH[0]}"
      local owner_repo="${BASH_REMATCH[1]}"
      local ref="${BASH_REMATCH[2]}"

      # Skip local or composite actions (./path) and docker:// images
      if [[ "$owner_repo" =~ ^\./ ]] || [[ "$owner_repo" =~ ^docker:// ]]; then
        printf '%s\n' "$line" >> "$tmp"
        continue
      fi

      # Already pinned to SHA?
      if [[ "$ref" =~ ^[0-9a-f]{40}$ ]]; then
        printf '%s\n' "$line" >> "$tmp"
        continue
      fi

      if sha=$(resolve_sha_for_ref "$owner_repo" "$ref"); then
        did_change=1
        changed=1
        changes+=("$file: $owner_repo@$ref -> $owner_repo@$sha")
        if [[ "$DRY_RUN" -eq 1 ]]; then
          printf '%s\n' "$line" >> "$tmp"
        else
          # Replace only this occurrence of @ref with @sha in the current line
          printf '%s\n' "${line/@$ref/@$sha}" >> "$tmp"
        fi
      else
        echo "warn: unable to resolve $owner_repo@$ref to a commit SHA" >&2
        printf '%s\n' "$line" >> "$tmp"
      fi
    else
      printf '%s\n' "$line" >> "$tmp"
    fi
  done < "$file"

  if [[ "$DRY_RUN" -eq 0 ]]; then
    if [[ "$did_change" -eq 1 ]]; then
      mv "$tmp" "$file"
    else
      rm -f "$tmp"
    fi
  else
    rm -f "$tmp"
  fi
}

shopt -s nullglob
for f in .github/workflows/*.yml .github/workflows/*.yaml; do
  pin_file "$f"
done

if [[ ${#changes[@]} -gt 0 ]]; then
  echo "Pinned the following action refs:" >&2
  for c in "${changes[@]}"; do echo "  - $c" >&2; done
else
  echo "No action refs changed (already pinned or none found)." >&2
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "(dry-run) No files were modified." >&2
fi
