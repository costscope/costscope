#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

ci::require_cmd go
ci::log "Building invariants-ci (duckdb)"
go build -tags duckdb -o bin/invariants-ci ./scripts/tools/invariants-ci
