#!/usr/bin/env bash
set -euo pipefail

# contract-diff.sh - Compare generated OpenAPI specs against baseline snapshots.
# Exit 0 if no breaking changes, >0 otherwise.
# Breaking rules (initial simple heuristic):
#  - Removed path, method, response code, or schema property = breaking (fail)
#  - Added optional field = allowed (pass)
#  - Changed type of existing field = breaking (fail)
#
# Tool preference: uses oasdiff (if installed) else falls back to a minimal grep heuristic.
# Install oasdiff (align with CI pinned commit):
#   go install github.com/oasdiff/oasdiff@fc23f9bb1b54519f4f847e1724dbd0ab894e8ec8
# See .github/workflows/api-contract-guard.yml for the current pinned revision.

RED="\033[31m"; GREEN="\033[32m"; YELLOW="\033[33m"; RESET="\033[0m"
REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
BASE_DIR="$REPO_ROOT/api"
GEN_PUBLIC="$REPO_ROOT/internal/api/docs/openapi.yaml"
GEN_ENT="$REPO_ROOT/internal/api/docs/enterprise-openapi.yaml"
BASE_PUBLIC="$BASE_DIR/openapi.v1.json"
BASE_ENT="$BASE_DIR/openapi.enterprise.v1.json"

fail_msgs=()

# Helper: verify CHANGELOG note when override used
check_changelog_note() {
  local label=$1
  local changelog="$REPO_ROOT/CHANGELOG.md"
  if [ ! -f "$changelog" ]; then
    echo -e "${RED}[contract] ALLOW_API_DIFF=1 but CHANGELOG.md missing (required)${RESET}";
    fail_msgs+=("$label missing CHANGELOG.md for breaking override");
    return 1
  fi
  if grep -qiE 'openapi|api contract|api breaking' "$changelog"; then
    echo -e "${YELLOW}[contract] Breaking change in $label acknowledged in CHANGELOG (override)${RESET}";
    return 0
  fi
  echo -e "${RED}[contract] ALLOW_API_DIFF=1 but no CHANGELOG line mentioning OpenAPI/API contract${RESET}";
  fail_msgs+=("$label missing CHANGELOG note for breaking change override");
  return 1
}

check_file() {
  local baseline=$1
  local generated=$2
  local label=$3
  if [ ! -f "$generated" ]; then
    echo -e "${RED}[contract] Missing generated spec: $generated${RESET}"; fail_msgs+=("$label missing generated spec"); return
  fi
  if ! command -v oasdiff >/dev/null 2>&1; then
    echo -e "${YELLOW}[contract] oasdiff not found; using lightweight heuristic for $label${RESET}";
    # Extract baseline paths (JSON key names at depth 2 under "paths")
    if command -v jq >/dev/null 2>&1; then
      jq -r '.paths | keys[]' "$baseline" 2>/dev/null || true | while read -r p; do
        [ -z "$p" ] && continue
        if ! grep -q "$p" "$generated"; then
          echo -e "${RED}[contract] Removed path $p in $label${RESET}"; fail_msgs+=("$label removed path $p")
        fi
      done
    else
      grep -o '"/[^"]*"' "$baseline" | tr -d '"' | sort -u | while read -r p; do
        [ -z "$p" ] && continue
        if ! grep -q "$p" "$generated"; then
          echo -e "${RED}[contract] Removed path $p in $label${RESET}"; fail_msgs+=("$label removed path $p")
        fi
      done
      # Heuristic schema property removal detection (jq required)
      if command -v jq >/dev/null 2>&1; then
        jq -r '.components.schemas // {} | to_entries[] | .key as $schema | (.value.properties // {} | keys[] | "\($schema):\(.)")' "$baseline" 2>/dev/null | while read -r sp; do
          schema_name=${sp%%:*}
          prop_name=${sp#*:}
          # Look for occurrence of property name under same schema block in generated spec (simple grep fallback)
          if ! grep -A10 -F "$schema_name" "$generated" | grep -q "^[[:space:]]*$prop_name:"; then
            echo -e "${RED}[contract] Removed schema property $schema_name.$prop_name in $label${RESET}"; fail_msgs+=("$label removed property $schema_name.$prop_name")
          fi
        done
      fi
    fi
  else
    echo "[contract] Running oasdiff (breaking) for $label";
    if ! oasdiff breaking "$baseline" "$generated" --fail-on ERR >/dev/null 2>&1; then
      echo -e "${RED}[contract] Breaking changes detected in $label${RESET}";
      # Show diff details
      oasdiff breaking "$baseline" "$generated" || true
      if [ "${ALLOW_API_DIFF:-}" = "1" ]; then
        check_changelog_note "$label" || true
      else
        echo -e "${YELLOW}[contract] Set ALLOW_API_DIFF=1 and add CHANGELOG note (OpenAPI/API contract) to override.${RESET}";
        fail_msgs+=("$label breaking changes")
      fi
    else
      echo -e "${GREEN}[contract] No breaking changes in $label${RESET}";
    fi
  fi
}

check_file "$BASE_PUBLIC" "$GEN_PUBLIC" "public"
check_file "$BASE_ENT" "$GEN_ENT" "enterprise"

if [ ${#fail_msgs[@]} -gt 0 ]; then
  echo -e "${RED} API contract guard failed:${RESET}";
  printf ' - %s\n' "${fail_msgs[@]}"
  exit 1
fi
echo -e "${GREEN} API contract guard passed${RESET}"
