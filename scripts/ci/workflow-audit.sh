#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

ci::log "Searching for unsafe workflow patterns..."
BAD=0

# Enforce critical pins only on push to main; PRs get warnings to ease migration
ENFORCE_CRITICAL=0
if [[ "${GITHUB_EVENT_NAME:-}" == "push" && "${GITHUB_REF:-}" == "refs/heads/main" ]]; then
  ENFORCE_CRITICAL=1
fi

echo -e "\n1) Disallow master/main/latest pins"
# Find uses pinned to master/main/latest or using non-pinned channels
grep -R --line-number "uses: .*@\(master\|main\|latest\)" .github/workflows || true
if grep -R --line-number "uses: .*@\(master\|main\|latest\)" .github/workflows >/dev/null 2>&1; then
  echo "Found action references pinned to master/main/latest — prefer version tags or commit SHAs" >&2
  BAD=1
fi

echo -e "\n2) Disallow curl | sh installs"
# Find curl|sh install patterns (exclude this audit workflow file to avoid self-matching)
if grep -R --line-number --exclude=workflow-audit.yml "curl .* | sh" .github/workflows >/dev/null 2>&1; then
  echo "Found curl | sh installs in workflows (excluding workflow-audit.yml) — prefer pinned releases or official actions" >&2
  BAD=1
fi

echo -e "\n3) Require commit SHA pins for critical actions (enforced on push to main)"
# For critical actions, require a 40-char commit SHA (no semver tag allowed).
critical_prefixes=(
  "actions/checkout"
  "actions/setup-go"
  "actions/cache"
  "actions/upload-artifact"
  "actions/download-artifact"
  "docker/setup-buildx-action"
  "docker/login-action"
  "docker/metadata-action"
  "docker/build-push-action"
  "docker/setup-qemu-action"
  "sigstore/cosign-installer"
  "aquasecurity/trivy-action"
  "zaproxy/action-baseline"
  "azure/setup-helm"
  "softprops/action-gh-release"
  "marocchino/sticky-pull-request-comment"
  "securego/gosec"
  "gitleaks/gitleaks-action"
  "actions/github-script"
)

while IFS= read -r line; do
  file=$(echo "$line" | cut -d: -f1)
  lineno=$(echo "$line" | cut -d: -f2)
  use=$(echo "$line" | cut -d: -f3- | sed -E 's/^.*uses:[[:space:]]*//')
  # Extract owner/repo and ref
  owner_repo=$(echo "$use" | sed -E 's/@.*//')
  # Extract ref after '@' and strip trailing comments or extra tokens
  ref=$(echo "$use" | sed -E 's/^[^@]+@//' | sed -E 's/[[:space:]]+#.*$//' | sed -E 's/[[:space:]].*$//')
  for p in "${critical_prefixes[@]}"; do
    if echo "$owner_repo" | grep -q "^$p$"; then
      if ! echo "$ref" | grep -Eq '^[0-9a-f]{40}$'; then
        if [[ "$ENFORCE_CRITICAL" -eq 1 ]]; then
          echo "Critical action not pinned to commit SHA: $file:$lineno uses: $owner_repo@$ref" >&2
          BAD=1
        else
          echo "warning: critical action not pinned to SHA (PR mode): $file:$lineno uses: $owner_repo@$ref" >&2
        fi
      fi
    fi
  done
done < <(grep -R --line-number --exclude=workflow-audit.yml "^[[:space:]]*uses:[[:space:]]*[^@[:space:]]\+@[^[:space:]]\+" .github/workflows || true)

echo -e "\n4) Flag generic unpinned refs (heuristic)"
# Find uses with '@' but neither a full sha nor a semver-like tag; still warn/fail.
# Trim trailing comments after the ref to avoid false positives when a ref is a SHA followed by a comment.
found_unpinned=0
while IFS= read -r line; do
  # extract only the ref part after '@' and trim anything after a space or comment symbol
  ref=$(echo "$line" | sed -E 's/.*@//' | sed -E 's/[[:space:]]+#.*$//' | sed -E 's/[[:space:]].*$//')
  # allow full 40-char SHAs
  if echo "$ref" | grep -Eq '^[0-9a-f]{40}$'; then
    continue
  fi
  # allow semver tags (vX.Y.Z or X.Y.Z or 4-part semver)
  if echo "$ref" | grep -Eq '^v?[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$'; then
    continue
  fi
  echo "Possible unpinned or non-immutable ref: $line" >&2
  found_unpinned=1
done < <(awk '/uses:/ {print FILENAME ":" NR ": " $0}' .github/workflows/* | grep -v "workflow-audit.yml" | grep -E "@[^[:space:]]+" || true)

if [[ "$found_unpinned" -ne 0 ]]; then
  BAD=1
fi

if [[ "$BAD" -ne 0 ]]; then
  ci::die "Workflow audit failed: unsafe patterns detected"
else
  ci::log "Workflow audit passed: no obvious unsafe patterns found"
fi
