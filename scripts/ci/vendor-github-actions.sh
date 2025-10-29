#!/usr/bin/env bash
set -euo pipefail

# Vendor pinned GitHub Actions into .github/_actions to avoid network fetches under act.
# This script downloads tarballs for specific action SHAs and places them under
# .github/_actions/<action-name>@<sha>/ so workflows referencing those SHAs can
# execute without external resolution.
#
# Note: Intended for local development/testing (e.g., nektos/act). Do not commit the
# vendored directories to VCS; keep them ignored via .gitignore if needed.

ROOT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
TARGET_DIR="$ROOT_DIR/.github/_actions"
TMP_DIR="${TMPDIR:-/tmp}/vendor-actions-$$"

mkdir -p "$TARGET_DIR" "$TMP_DIR"

# Static fallbacks for known SHAs we rely on or that act remaps to implicitly.
static_actions=(
  # name;repo;sha
  "actions-setup-go;actions/setup-go;44694675825211faa026b3c33043df3e48a5fa00"
  # Some act runners remap setup-go to this commit; include proactively.
  "actions-setup-go;actions/setup-go;19bb51245e9c80abacb2e91cc42b33fa478b8639"
  "actions-cache;actions/cache;0057852bfaa89a56745cba8c7296529d2fc39830"
  "actions-upload-artifact;actions/upload-artifact;ea165f8d65b6e75b540449e92b4886f43607fa02"
  "actions-checkout;actions/checkout;08c6903cd8c0fde910a37f88322edcfb5dd907a8"
)

discover_actions_from_workflows() {
  # Extract pinned actions with commit SHA from all workflows.
  # Output format: repo;sha
  local wf_dir="$ROOT_DIR/.github/workflows"
  [[ -d "$wf_dir" ]] || return 0
  # Grep uses lines like: uses: actions/checkout@08c6903c...
  # We only capture when ref looks like a commit (7-40 hex chars).
  grep -RhoE "uses:[[:space:]]*actions/[a-zA-Z0-9_-]+@[0-9a-f]{7,40}" "$wf_dir" 2>/dev/null |
    sed -E 's/.*uses:[[:space:]]*(actions\/[^@]+)@([0-9a-f]{7,40}).*/\1;\2/' |
    sort -u
}

fetch_and_unpack() {
  local name="$1" repo="$2" sha="$3"
  local dest="$TARGET_DIR/${name}@${sha}"
  if [[ -d "$dest" ]]; then
    echo "[vendor-actions] Exists: $dest (skipping)"
    return 0
  fi
  echo "[vendor-actions] Fetching $repo@$sha -> $dest"
  local tarball="$TMP_DIR/${name}-${sha}.tar.gz"
  # GitHub tarball URL for a specific commit
  local url="https://api.github.com/repos/${repo}/tarball/${sha}"
  # Use curl with GitHub API; unauthenticated may be rate limited. Optionally set GH_TOKEN for higher limits.
  curl -fsSL -H "Accept: application/vnd.github+json" -o "$tarball" "$url"
  local unpack_dir="$TMP_DIR/unpack-${name}-${sha}"
  mkdir -p "$unpack_dir"
  tar -xzf "$tarball" -C "$unpack_dir"
  # Move extracted root folder (unknown prefix) into destination
  local top
  top=$(find "$unpack_dir" -mindepth 1 -maxdepth 1 -type d | head -n1)
  if [[ -z "$top" ]]; then
    echo "[vendor-actions] ERROR: Failed to identify extracted directory for $repo@$sha" >&2
    return 1
  fi
  mkdir -p "$dest"
  # Move the contents (not the container directory) to keep expected layout
  shopt -s dotglob
  mv "$top"/* "$dest"/
  shopt -u dotglob
  echo "[vendor-actions] Vendored $repo@$sha into $dest"
}

main() {
  # 1) Vendoring actions statically listed above (safety net for act remaps).
  for entry in "${static_actions[@]}"; do
    IFS=';' read -r name repo sha <<<"$entry"
    fetch_and_unpack "$name" "$repo" "$sha"
  done
  # 2) Discover pinned actions from workflows and vendor them as well.
  while IFS= read -r entry; do
    [[ -z "$entry" ]] && continue
    IFS=';' read -r repo sha <<<"$entry"
    name="actions-$(basename "$repo")"
    fetch_and_unpack "$name" "$repo" "$sha"
  done < <(discover_actions_from_workflows)
  echo "[vendor-actions] Done. Contents of $TARGET_DIR:"
  ls -la "$TARGET_DIR" || true
}

trap 'rm -rf "$TMP_DIR"' EXIT
main "$@"
