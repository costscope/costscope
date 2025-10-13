#!/usr/bin/env bash
# check-allowlist-rationale.sh
#
# Purpose:
#   Enforce that every non-empty, non-comment line in an allowlist file includes an inline
#   justification comment containing the token "# rationale:" followed by a non-empty reason.
#
# Why:
#   1. Governance / auditability – institutional memory for each exception.
#   2. Dead code / drift prevention – allows periodic pruning of obsolete allowlist entries.
#   3. Security posture – discourage silent expansion of allowlists.
#
# Usage:
#   scripts/tools/check-allowlist-rationale.sh [--format text|json] [--min-len N] [--strict-case] <file> [<file>...]
#
# Exit Codes:
#   0 = success (no violations)
#   1 = one or more violations found
#   2 = usage / internal error
#
# Rules (encoded below):
#   R1: Every substantive line MUST contain the substring "# rationale:" (default case‑insensitive unless --strict-case).
#   R2: The rationale text (trimmed) AFTER the colon must be non-empty and at least --min-len characters (default 5).
#   R3: Lines that are blank or start with '#' (after optional leading whitespace) are ignored.
#   R4: Trailing whitespace is ignored.
#   R5: Multiple rationale tokens on one line: the first is evaluated; others are allowed but discouraged.
#   R6: If a suppression token "# allowlist-ignore-rationale" appears, the line is skipped (explicit waiver – include sparingly).
#
# Suggested Conventions:
#   <entry><space># rationale: concise reason (<ticket/id/owner>)
#   Example: my.domain.ClassName # rationale: legacy provider SDK still required (JIRA-123, owner:platform)
#
# Example Violations (text format):
#   path/allowlist.txt:12: missing rationale token (R1)
#   path/allowlist.txt:27: empty rationale text (R2)
#   path/allowlist.txt:33: rationale too short (<5 chars) (R2) -> "fix"
#
# JSON format output structure:
# {
#   "files": [
#     {
#       "file": "path/allowlist.txt",
#       "violations": [
#         {"line":12, "rule":"R1", "message":"missing rationale token"},
#         {"line":27, "rule":"R2", "message":"empty rationale text"}
#       ]
#     }
#   ],
#   "summary": {"files":1, "violations":2}
# }

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

FORMAT="text"
MIN_LEN=5
STRICT_CASE=0

err() { echo "[check-allowlist-rationale] ERROR: $*" >&2; }
usage() {
  cat <<EOF
Usage: $0 [--format text|json] [--min-len N] [--strict-case] <file> [<file>...]
Checks that each allowlist line includes an inline '# rationale: <reason>' justification.

Exit codes: 0 (ok), 1 (violations), 2 (usage/internal error)
EOF
}

if [[ $# -lt 1 ]]; then
  usage; exit 2
fi

FILES=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --format)
      FORMAT="${2:-}"; shift 2 || { usage; exit 2; } ;;
    --min-len)
      MIN_LEN="${2:-}"; shift 2 || { usage; exit 2; } ;;
    --strict-case)
      STRICT_CASE=1; shift ;;
    -h|--help)
      usage; exit 0 ;;
    --)
      shift; while [[ $# -gt 0 ]]; do FILES+=("$1"); shift; done ;;
    *)
      FILES+=("$1"); shift ;;
  esac
done

if [[ ${#FILES[@]} -eq 0 ]]; then
  usage; exit 2
fi

case "$FORMAT" in
  text|json) ;; 
  *) err "invalid --format '$FORMAT'"; exit 2 ;;
esac

if ! [[ "$MIN_LEN" =~ ^[0-9]+$ ]]; then
  err "--min-len must be integer"; exit 2
fi

declare -A FILE_VIOL_COUNT=()
TOTAL_VIOL=0
JSON_FILES=()

for f in "${FILES[@]}"; do
  if [[ ! -f "$f" ]]; then
    err "file not found: $f"; exit 2
  fi
  mapfile -t LINES < "$f"
  VIOLS=()
  for i in "${!LINES[@]}"; do
    lineno=$((i+1))
    line="${LINES[$i]}"
    trimmed="${line%$'\r'}"  # strip CR
    # Skip blank & comment-only lines
    if [[ -z "${trimmed//[[:space:]]/}" ]]; then
      continue
    fi
    if [[ $trimmed =~ ^[[:space:]]*# ]]; then
      continue
    fi
    if [[ $trimmed == *"# allowlist-ignore-rationale"* ]]; then
      continue
    fi
    search_line="$trimmed"
    token_pattern="# rationale:"
    if [[ $STRICT_CASE -eq 0 ]]; then
      # lower-case both for search
      search_line="${search_line,,}"
    fi
    if [[ $search_line != *"# rationale:"* ]]; then
      msg="missing rationale token"
      VIOLS+=("$lineno|R1|$msg")
      continue
    fi
    # Extract original (case-sensitive) token segment
    # Use awk to split on '# rationale:' ignoring preceding content (first occurrence)
  rationale_segment="$(echo "$trimmed" | sed 's/^[^#]*# rationale:/# rationale:/I')"
    rationale_text="$(echo "$rationale_segment" | sed -E 's/.*# rationale:[[:space:]]*(.*)$/\1/I')"
    rationale_text_trimmed="$(echo "$rationale_text" | sed -E 's/[[:space:]]+$//')"
    if [[ -z "$rationale_text_trimmed" ]]; then
      VIOLS+=("$lineno|R2|empty rationale text")
      continue
    fi
    if (( ${#rationale_text_trimmed} < MIN_LEN )); then
      VIOLS+=("$lineno|R2|rationale too short (<$MIN_LEN chars) -> '${rationale_text_trimmed}'")
      continue
    fi
  done
  if [[ ${#VIOLS[@]} -gt 0 ]]; then
    FILE_VIOL_COUNT["$f"]=${#VIOLS[@]}
    TOTAL_VIOL=$((TOTAL_VIOL + ${#VIOLS[@]}))
    if [[ $FORMAT == "text" ]]; then
      for v in "${VIOLS[@]}"; do
        IFS='|' read -r l rule msg <<< "$v"
        echo "$f:$l: $msg ($rule)"
      done
    else
      # Build JSON array entries per file
  file_json="{\"file\":\"$f\",\"violations\":["
      first=1
      for v in "${VIOLS[@]}"; do
        IFS='|' read -r l rule msg <<< "$v"
  # Escape embedded double quotes for JSON safety
  esc_msg=${msg//\"/\\\"}
  esc_msg=${esc_msg//"/\\"}
        if [[ $first -eq 0 ]]; then file_json+=" ,"; fi
        file_json+="{\"line\":$l,\"rule\":\"$rule\",\"message\":\"$esc_msg\"}"
        first=0
      done
  file_json+="]}"
      JSON_FILES+=("$file_json")
    fi
  fi
done

if [[ $FORMAT == "json" ]]; then
  echo -n '{"files":['
  for i in "${!JSON_FILES[@]}"; do
    [[ $i -gt 0 ]] && echo -n ','
    echo -n "${JSON_FILES[$i]}"
  done
  echo -n '],"summary":{'
  echo -n "\"files\":${#JSON_FILES[@]},\"violations\":$TOTAL_VIOL" 
  echo '}}'
fi

if [[ $TOTAL_VIOL -gt 0 ]]; then
  exit 1
fi
exit 0
