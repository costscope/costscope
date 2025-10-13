#!/usr/bin/env bash
set -euo pipefail

# Try to source common CI helper for consistent logging if available
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMMON_SH="$REPO_ROOT/scripts/ci/lib/common.sh"
if [[ -f "$COMMON_SH" ]]; then
	# shellcheck disable=SC1090
	. "$COMMON_SH"
fi

# This script has been deprecated and replaced by the Go generator (cmd/tools/gen-openapi).
# It intentionally exits with non-zero status to surface any lingering references.
if command -v ci::die >/dev/null 2>&1; then
	ci::die "generate-openapi.sh removed. Use: go run ./cmd/tools/gen-openapi or make api-spec-generate"
else
	echo "generate-openapi.sh removed. Use: go run ./cmd/tools/gen-openapi or make api-spec-generate" >&2
	exit 1
fi
