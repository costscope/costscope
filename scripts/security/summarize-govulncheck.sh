#!/usr/bin/env bash
set -euo pipefail

# Summarize a govulncheck JSON report (NDJSON supported) and optionally fail if vulnerabilities are found.
# Env / Args:
#   INPUT (env or $1): path to govulncheck JSON output file
#   FAIL_ON_VULNS (optional, default: ""): if set to non-empty truthy value, exit 1 when advisories found

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

INPUT_PATH="${INPUT:-${1:-}}"
if [[ -z "$INPUT_PATH" ]]; then
  ci::err "INPUT not set and no positional path provided"
  exit 2
fi
if [[ ! -f "$INPUT_PATH" ]]; then
  ci::err "Input file not found: $INPUT_PATH"
  exit 2
fi

SIZE_BYTES=$(wc -c <"$INPUT_PATH" | tr -d ' ')
if [[ "${SIZE_BYTES:-0}" -le 0 ]]; then
  ci::warn "Input file is empty: $INPUT_PATH"
fi

# Validate JSON if jq is available (slurp to accept NDJSON)
if command -v jq >/dev/null 2>&1; then
  if ! jq -e -s '.' "$INPUT_PATH" >/dev/null 2>&1; then
    ci::warn "jq slurp validation reported an issue (file may be NDJSON with non-JSON lines). Continuing."
  fi
fi

# Extract unique advisory IDs (GO-...) using a regex approach for robustness across versions
UNIQ_IDS=$(grep -oE '"id"[[:space:]]*:[[:space:]]*"GO-[0-9-]+"' "$INPUT_PATH" | sed -E 's/.*"(GO-[0-9-]+)"/\1/' | sort -u || true)
UNIQ_COUNT=$(printf "%s\n" "$UNIQ_IDS" | sed '/^$/d' | wc -l | tr -d ' ')

ci::log "Unique advisories: ${UNIQ_COUNT}"
if [[ "$UNIQ_COUNT" -gt 0 ]]; then
  ci::log "Top advisory IDs (up to 50):"
  printf "%s\n" "$UNIQ_IDS" | head -50 | sed 's/^/ - /'
fi

FAIL_ON_VULNS_VAL="${FAIL_ON_VULNS:-}"
if [[ -n "$FAIL_ON_VULNS_VAL" ]]; then
  shopt -s nocasematch || true
  if [[ "$FAIL_ON_VULNS_VAL" =~ ^(1|true|yes|on)$ && "$UNIQ_COUNT" -gt 0 ]]; then
    ci::err "Vulnerabilities detected (${UNIQ_COUNT} unique advisories) and FAIL_ON_VULNS is enabled"
    exit 1
  fi
fi

ci::log "Summary complete for $INPUT_PATH"