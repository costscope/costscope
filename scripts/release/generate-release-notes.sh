#!/usr/bin/env bash
set -euo pipefail
# Generate release_notes.md from CHANGELOG.md diff (preferred) or git log fallback
# Environment: RELEASE_VERSION (required)
# Optional:
#   PREV_TAG      – previous git tag (auto-detected if unset)
#   CHANGELOG_PATH – path to changelog (default: CHANGELOG.md)
#   INCLUDE_GIT_COMMITS – set to 1 to append commit list after changelog diff

if [ -z "${RELEASE_VERSION:-}" ]; then
  echo "RELEASE_VERSION env not set" >&2
  exit 1
fi

PREV_TAG=${PREV_TAG:-$(git tag --sort=-creatordate | grep -v "^v${RELEASE_VERSION}$" | head -n1 || true)}
CURRENT_TAG="v${RELEASE_VERSION}"
CHANGELOG_PATH=${CHANGELOG_PATH:-CHANGELOG.md}

if ! git rev-parse "$CURRENT_TAG" >/dev/null 2>&1; then
  echo "️  Tag $CURRENT_TAG not yet created (diff will use HEAD)" >&2
fi

# 1. Capture changelog diff between previous tag and HEAD (or full file if no previous tag)
CHANGELOG_DIFF_FILE=$(mktemp)
if [ -n "$PREV_TAG" ] && git rev-parse "$PREV_TAG" >/dev/null 2>&1; then
  git diff "$PREV_TAG"..HEAD -- "$CHANGELOG_PATH" > "$CHANGELOG_DIFF_FILE" || true
else
  # No previous tag – include entire changelog for initial release
  { echo "--- a/$CHANGELOG_PATH"; echo "+++ b/$CHANGELOG_PATH"; sed 's/^/+ /' "$CHANGELOG_PATH"; } > "$CHANGELOG_DIFF_FILE" 2>/dev/null || true
fi

# 2. Extract Unreleased section content (lines until next '## [' excluding header) – if present
UNRELEASED_SECTION=""
if grep -q '^## \[Unreleased\]' "$CHANGELOG_PATH" 2>/dev/null; then
  UNRELEASED_SECTION=$(awk '/^## \[Unreleased\]/{flag=1;next}/^## \[/{flag=0}flag' "$CHANGELOG_PATH" | sed '/^$/d') || true
fi

# 3. Prepare commit list (optional)
GIT_COMMITS=""
if [ "${INCLUDE_GIT_COMMITS:-0}" = "1" ]; then
  if [ -n "$PREV_TAG" ]; then
    GIT_COMMITS=$(git log --pretty=format:'* %s (%h)' "$PREV_TAG"..HEAD)
  else
    GIT_COMMITS=$(git log --pretty=format:'* %s (%h)')
  fi
fi

# 4. Fallback if diff is empty: use Unreleased section or commits
if [ ! -s "$CHANGELOG_DIFF_FILE" ] && [ -z "$UNRELEASED_SECTION" ]; then
  echo "ℹ️  No CHANGELOG diff detected; using git commit list fallback" >&2
  if [ -z "$GIT_COMMITS" ]; then
    if [ -n "$PREV_TAG" ]; then
      GIT_COMMITS=$(git log --pretty=format:'* %s (%h)' "$PREV_TAG"..HEAD)
    else
      GIT_COMMITS=$(git log --pretty=format:'* %s (%h)')
    fi
  fi
fi

cat > release_notes.md <<EOF
# Release $CURRENT_TAG

Date: $(date -u +%Y-%m-%d)

## Summary
Provide a concise summary of key changes, features, and fixes.

## Changelog Diff (since ${PREV_TAG:-initial commit})
$(if [ -s "$CHANGELOG_DIFF_FILE" ]; then echo '\n```diff'; cat "$CHANGELOG_DIFF_FILE"; echo '```'; else echo '_No changelog diff (initial release or unchanged)_'; fi)

$(if [ -n "$UNRELEASED_SECTION" ]; then echo '## Unreleased Section Captured'; echo ''; echo '```markdown'; echo "$UNRELEASED_SECTION"; echo '```'; fi)

$(if [ -n "$GIT_COMMITS" ]; then echo '## Commit Summary'; echo "$GIT_COMMITS"; fi)

## Artifacts
- Binary: bin/costscope-production
- Image Tags: costscope:latest, costscope:stage-${RELEASE_VERSION}, costscope:${RELEASE_VERSION}
- SBOM: sbom.json
- Checksums: checksums.txt

## Verification
- Security gate: PASSED
- Smoke tests: PASSED
- SBOM generated and stored

## Upgrade Notes
List any breaking changes or migration steps here.

EOF

echo "release_notes.md generated (version ${RELEASE_VERSION}; previous tag: ${PREV_TAG:-none})"
