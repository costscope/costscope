#!/usr/bin/env bash
set -euo pipefail

# deadcode-baseline-guard.sh
# Compares current deadcode output with a stored baseline JSON and enforces:
#  - No >2 new symbols unless each is allowlisted with rationale.
#  - No removed symbols lingering in allowlist.
# Baseline format (JSON):
# {
#   "generated_at": "2025-08-25T12:34:56Z",
#   "symbols": ["Pkg.Func", ...]
# }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ALLOWLIST_FILE="${ROOT_DIR}/.deadcode-allowlist"
BASELINE_FILE="${ROOT_DIR}/.deadcode-baseline.json"
OUTPUT_FILE="${ROOT_DIR}/logs/deadcode_current.json"

mkdir -p "${ROOT_DIR}/logs"

if ! command -v deadcode >/dev/null 2>&1; then
  ci::warn "deadcode tool not found in PATH – attempting installation (go install github.com/tsenart/deadcode@latest)"
  if command -v go >/dev/null 2>&1; then
    # Attempt install (network required). Silence module download noise; show errors on failure.
    if GO111MODULE=on GOINSECURE= GONOSUMDB= GONOPROXY= go install github.com/tsenart/deadcode@latest 2>/dev/null; then
      if ! command -v deadcode >/dev/null 2>&1; then
        echo "deadcode install reported success but binary not found in GOPATH/bin (ensure GOPATH/bin in PATH)" >&2
        exit 2
      fi
      ci::log "deadcode tool installed successfully."
    else
      ci::die "automatic installation failed; please install manually: go install github.com/tsenart/deadcode@latest"
      exit 2
    fi
  else
    ci::die "Go toolchain not available to install deadcode."
    exit 2
  fi
fi

tmp_scan="$(mktemp)"; trap 'rm -f "$tmp_scan"' EXIT
deadcode ./... >"$tmp_scan" || true

parse_symbols() {
  # Extract only the exported symbol name (final identifier) so it aligns with allowlist lines (which do not include paths).
  # deadcode output example: internal/pkg/file.go:123: func unreachable FooBar ...
  # We grep the last exported-looking identifier (capitalized start). Some lines may not match; suppress non-zero grep exit.
  awk '{print $0}' "$1" | while IFS= read -r l; do
    sym=$( (echo "$l" | grep -Eo '[A-Z][A-Za-z0-9_]+' || true) | tail -1 )
    [[ -n "$sym" ]] && echo "$sym"
  done | sort -u
}

current_syms=$(parse_symbols "$tmp_scan")
jq_current=$(printf '%s\n' "$current_syms" | jq -R -s 'split("\n")|map(select(length>0))')
echo '{"generated_at":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","symbols":'"$jq_current"'}' > "$OUTPUT_FILE"

if [[ ! -f "$BASELINE_FILE" ]]; then
  ci::warn "Baseline missing ($BASELINE_FILE). Initialize by copying logs/deadcode_current.json."
  echo "INIT_REQUIRED" >&2
  exit 4
fi

baseline_syms=$(jq -r '.symbols[]' "$BASELINE_FILE" 2>/dev/null || true)
current_list=$(jq -r '.symbols[]' "$OUTPUT_FILE")

new_syms=$(comm -13 <(printf '%s\n' $baseline_syms | sort) <(printf '%s\n' $current_list | sort) | grep -v '^$' || true)
removed_syms=$(comm -23 <(printf '%s\n' $baseline_syms | sort) <(printf '%s\n' $current_list | sort) | grep -v '^$' || true)

# Check allowlist rationale presence for new symbols (bare symbol names)
missing_rationale=()
while IFS= read -r s; do
  [[ -z "$s" ]] && continue
  # symbol presence in allowlist with '# rationale:' substring on same line
  if ! grep -E "^$s(\b| ).*# rationale:" "$ALLOWLIST_FILE" >/dev/null 2>&1; then
    missing_rationale+=("$s")
  fi
done < <(printf '%s\n' "$new_syms")

count_new=$(printf '%s\n' "$new_syms" | awk 'NF' | wc -l | tr -d ' ')

status=0
if (( count_new > 2 )); then
  ci::warn " Deadcode baseline guard: too many new unreachable symbols ($count_new > 2)."
  status=1
fi
if ((${#missing_rationale[@]})); then
  ci::warn " New symbols without allowlist rationale:"
  printf ' - %s\n' "${missing_rationale[@]}" >&2
  status=1
fi
if [[ -n "$removed_syms" ]]; then
  while IFS= read -r rs; do
    [[ -z "$rs" ]] && continue
    if grep -E "^$rs" "$ALLOWLIST_FILE" >/dev/null 2>&1; then
      ci::warn " Symbol '$rs' removed from scan but still in allowlist."
      status=1
    fi
  done < <(printf '%s\n' "$removed_syms")
fi

if (( status == 0 )); then
  ci::log " Deadcode baseline guard passed (new=$count_new)."
fi
if (( status != 0 )) && [[ "${COSTSCOPE_DEADCODE_DEBUG:-}" == "1" ]]; then
  echo "--- DEBUG deadcode-baseline-guard ---" >&2
  echo "count_new=$count_new" >&2
  echo "new_syms=$(printf '%s' "$new_syms" | tr '\n' ' ')" >&2
  echo "removed_syms=$(printf '%s' "$removed_syms" | tr '\n' ' ')" >&2
  echo "missing_rationale=${missing_rationale[*]:-}" >&2
  echo "baseline_count=$(printf '%s\n' $baseline_syms | awk 'NF' | wc -l | tr -d ' ')" >&2
  echo "current_count=$(printf '%s\n' $current_list | awk 'NF' | wc -l | tr -d ' ')" >&2
  echo "-------------------------------------" >&2
fi
exit $status
