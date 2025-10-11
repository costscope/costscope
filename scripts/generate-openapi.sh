#!/usr/bin/env bash
# This script has been deprecated and replaced by the Go generator (cmd/tools/gen-openapi).
# It intentionally exits with non-zero status to surface any lingering references.
echo "generate-openapi.sh removed. Use: go run ./cmd/tools/gen-openapi or make api-spec-generate" >&2
echo "This script has been removed. Use: go run ./cmd/tools/gen-openapi" >&2
exit 1
