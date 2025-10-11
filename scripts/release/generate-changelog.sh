#!/usr/bin/env bash
set -euo pipefail

VERSION=${1:?usage: $0 <version>}

git config user.name "costscope-bot"
git config user.email "actions@users.noreply.github.com"

git cliff --config .git-cliff.toml --tag "${VERSION}" -o CHANGELOG.md
git add CHANGELOG.md || true
if ! git diff --cached --quiet; then
  git commit -m "chore(release): update CHANGELOG for ${VERSION}"
  git push || echo "(Non-fatal) Unable to push changelog (possibly read-only on tag builds)"
fi

git cliff --config .git-cliff.toml --tag "${VERSION}" > RELEASE_NOTES.md
echo '--- Generated Release Notes (preview) ---'
head -n 40 RELEASE_NOTES.md || true
