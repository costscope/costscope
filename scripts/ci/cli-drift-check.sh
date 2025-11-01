#!/usr/bin/env bash
set -euo pipefail
# CLI command drift check script
# Generates command builders and fails with artifacts if drift detected.
# Intended for use in both local dev and CI.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

ARTIFACT_SUMMARY="cli-drift-summary.txt"
ARTIFACT_DIFF="cli-diff.patch"

# Generate builders into a temp directory and compare against committed files
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

AN_OUT="$TMP_DIR/analytics_zz_generated_command_builder.go"
MC_OUT="$TMP_DIR/multicloud_zz_generated_command_builder.go"

# Regenerate into temp files (avoid mutating working tree)
go run -buildvcs=false ./scripts/tools/commandgen -receiver AnalyticsCommands -spec cmd/modules/analytics/commands/command_spec.yaml -out "$AN_OUT"
go run -buildvcs=false ./scripts/tools/commandgen -receiver MulticloudCommands -spec cmd/modules/multicloud/commands/command_spec.yaml -out "$MC_OUT"

# Optionally format generated temp files (avoids cosmetic drift)
if command -v gofmt >/dev/null 2>&1; then
  gofmt -w "$AN_OUT" "$MC_OUT" || true
fi

# Normalize file permissions (should not affect textual diff)
chmod 0644 "$AN_OUT" "$MC_OUT" || true

AN_REPO="cmd/modules/analytics/commands/zz_generated_command_builder.go"
MC_REPO="cmd/modules/multicloud/commands/zz_generated_command_builder.go"

DIFF_EXIT=0
{
  # Disable rename detection and produce unified diffs for clarity
  if ! git --no-pager diff -M0 --no-index -- "$AN_REPO" "$AN_OUT"; then DIFF_EXIT=1; fi
  if ! git --no-pager diff -M0 --no-index -- "$MC_REPO" "$MC_OUT"; then DIFF_EXIT=1; fi
} > "$ARTIFACT_DIFF" 2>/dev/null || true

if [ "$DIFF_EXIT" -ne 0 ]; then
  echo "CLI command drift detected" | tee "$ARTIFACT_SUMMARY"
  echo "--- Begin CLI drift (first 200 lines) ---"
  sed -n '1,200p' "$ARTIFACT_DIFF" || true
  echo "--- End CLI drift excerpt ---"
  exit 1
else
  echo "No CLI command drift" > "$ARTIFACT_SUMMARY"
fi
