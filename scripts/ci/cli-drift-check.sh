#!/usr/bin/env bash
set -euo pipefail
# CLI command drift check script
# Generates command builders and fails with artifacts if drift detected.
# Intended for use in both local dev and CI.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

ARTIFACT_SUMMARY="cli-drift-summary.txt"
ARTIFACT_DIFF="cli-diff.patch"

# Ensure VCS stamping does not fail inside ephemeral CI containers without full git metadata
# This flag is respected by all subsequent `go` commands (go run/build/test)
export GOFLAGS="${GOFLAGS:-} -buildvcs=false"

# Regenerate builders via make (preferred); fallback only if target genuinely missing
if make -q gen-commands >/dev/null 2>&1 || grep -q '^gen-commands:' mk/gen.mk; then
  make -s gen-commands
else
  echo "[info] fallback generation logic (gen-commands target not found)" >&2
  go run ./scripts/tools/commandgen -receiver AnalyticsCommands -spec cmd/modules/analytics/commands/command_spec.yaml -out cmd/modules/analytics/commands/zz_generated_command_builder.go
  go run ./scripts/tools/commandgen -receiver MulticloudCommands -spec cmd/modules/multicloud/commands/command_spec.yaml -out cmd/modules/multicloud/commands/zz_generated_command_builder.go
fi

# Optionally format (avoids cosmetic drift):
if command -v gofmt >/dev/null 2>&1; then
  gofmt -w cmd/modules/analytics/commands/zz_generated_command_builder.go cmd/modules/multicloud/commands/zz_generated_command_builder.go || true
fi

# Detect drift limited to generated files
if ! git diff --quiet -- cmd/modules/analytics/commands/zz_generated_command_builder.go cmd/modules/multicloud/commands/zz_generated_command_builder.go; then
  echo "CLI command drift detected" | tee "$ARTIFACT_SUMMARY"
  git --no-pager diff -- cmd/modules/analytics/commands/zz_generated_command_builder.go cmd/modules/multicloud/commands/zz_generated_command_builder.go > "$ARTIFACT_DIFF" || true
  exit 1
else
  echo "No CLI command drift" > "$ARTIFACT_SUMMARY"
fi
