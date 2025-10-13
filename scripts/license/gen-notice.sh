#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# gen-notice.sh
# Generates a deterministic NOTICE file enumerating all (direct + indirect) Go modules
# with best‑effort license detection. If a license cannot be heuristically identified
# it is reported as "Unknown". Detection intentionally avoids network calls; only
# local module directories under GOPATH / module cache are scanned.
#
# Usage:
#   bash scripts/license/gen-notice.sh [OUTPUT_FILE]
# Default OUTPUT_FILE = NOTICE
#
# Drift guard (CI): make notice-drift executes this script to a temp file and diffs.
# Adding / updating dependencies (go.mod / go.sum) without regenerating NOTICE will fail.

OUTPUT_FILE="${1:-NOTICE}"
PROJECT_MODULE=$(go list -m -f '{{ .Path }}')
PROJECT_VERSION=$(go list -m -f '{{ .Version }}' 2>/dev/null || echo "")
DATE_UTC=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse --short=12 HEAD 2>/dev/null || echo "unknown")

# Deterministic mode (for drift guard): suppress volatile metadata so repeated
# regenerations without dependency changes yield identical NOTICE.
if [ "${NOTICE_DETERMINISTIC:-}" = "1" ]; then
  DATE_UTC="(deterministic)"
  GIT_COMMIT="(deterministic)"
fi

# Collect module metadata: path|version|dir|indirect
MODULE_LINES=$(go list -m -f '{{ .Path }}|{{ .Version }}|{{ .Dir }}|{{ if .Indirect }}true{{ else }}false{{ end }}' all)

find_license_file() {
  # Ascend until a LICENSE-like file is found (or root).
  local start="$1"; local cur="$start"; local cand="";
  while [ -n "$cur" ] && [ "$cur" != "/" ]; do
    cand=$(find "$cur" -maxdepth 1 -type f \( -iname 'LICENSE' -o -iname 'LICENSE.*' -o -iname 'COPYING*' -o -iname 'NOTICE' \) 2>/dev/null | head -n1 || true)
    if [ -n "$cand" ]; then echo "$cand"; return 0; fi
    cur=$(dirname "$cur")
  done
  echo "" # none
}

detect_license() {
  local dir="$1"
  local lic="Unknown"; local src=""
  if [ -n "$dir" ] && [ -d "$dir" ]; then
    local cand
    # First try direct directory; if none, ascend.
    cand=$(find "$dir" -maxdepth 1 -type f \( -iname 'LICENSE' -o -iname 'LICENSE.*' -o -iname 'COPYING*' -o -iname 'NOTICE' \) 2>/dev/null | head -n1 || true)
    if [ -z "$cand" ]; then
      cand=$(find_license_file "$dir")
    fi
    if [ -n "$cand" ]; then
      src="$cand"
      # Read more lines for broader pattern match (short licenses may be >40 lines apart)
      local snippet
      snippet=$(head -n 120 "$cand" | tr '\r' ' ')
      # SPDX detection (first occurrence preserved; handles multi-license OR expressions)
      local spdx
      spdx=$(grep -E 'SPDX-License-Identifier:' "$cand" 2>/dev/null | head -n1 | sed -E 's/.*SPDX-License-Identifier:\s*//' | tr -d '\r' || true)
      if [ -n "$spdx" ]; then
        lic="$spdx"
      else
        shopt -s nocasematch || true
        if [[ $snippet == *"Apache License"* && $snippet == *"Version 2"* ]]; then lic="Apache-2.0"; fi
        # MIT alternative: many files omit title but have permission clause
        if [[ $lic == "Unknown" && $snippet == *"Permission is hereby granted"* && $snippet == *"MIT"* ]]; then lic="MIT"; fi
        if [[ $lic == "Unknown" && $snippet == *"Permission is hereby granted"* && $snippet == *"without restriction"* ]]; then lic="MIT"; fi
        if [[ $lic == "Unknown" && ( $snippet == *"BSD 3"* || ( $snippet == *"Redistribution and use in source and binary forms"* && $snippet == *"WITH OR WITHOUT MODIFICATION"* ) ) ]]; then lic="BSD-3-Clause"; fi
        # BSD-2 (no advertising clause reference)
        if [[ $lic == "Unknown" && $snippet == *"Redistribution and use in source and binary forms"* && $snippet != *"Neither the name of"* && $snippet != *"the names of its contributors"* ]]; then lic="BSD-2-Clause"; fi
        if [[ $lic == "Unknown" && $snippet == *"Mozilla Public License"* && $snippet == *"2.0"* ]]; then lic="MPL-2.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"GNU GENERAL PUBLIC LICENSE"* && $snippet == *"Version 3"* ]]; then lic="GPL-3.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"GNU GENERAL PUBLIC LICENSE"* && $snippet == *"Version 2"* ]]; then lic="GPL-2.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"Affero General Public License"* && $snippet == *"Version 3"* ]]; then lic="AGPL-3.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"GNU LESSER GENERAL PUBLIC LICENSE"* && $snippet == *"Version 2"* ]]; then lic="LGPL-2.1"; fi
        if [[ $lic == "Unknown" && $snippet == *"GNU LESSER GENERAL PUBLIC LICENSE"* && $snippet == *"Version 3"* ]]; then lic="LGPL-3.0"; fi
        if [[ $lic == "Unknown" && ( $snippet == *"ISC License"* || ( $snippet == *"Permission to use, copy, modify, and distribute"* && $snippet == *"without fee"* ) ) ]]; then lic="ISC"; fi
        if [[ $lic == "Unknown" && $snippet == *"Boost Software License"* && $snippet == *"1.0"* ]]; then lic="BSL-1.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"BOOST SOFTWARE LICENSE"* ]]; then lic="BSL-1.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"Eclipse Public License"* && $snippet == *"2.0"* ]]; then lic="EPL-2.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"THE ARTISTIC LICENSE"* && $snippet == *"Version 2.0"* ]]; then lic="Artistic-2.0"; fi
        if [[ $lic == "Unknown" && $snippet == *"Creative Commons"* && $snippet == *"Attribution"* ]]; then lic="CC-BY"; fi
        if [[ $lic == "Unknown" && $snippet == *"UNLICENSE"* || $snippet == *"This is free and unencumbered software released into the public domain"* ]]; then lic="Unlicense"; fi
      fi
    fi
  fi
  # Export via nameref (bash) simulated by echo assignments (caller evals) — simpler: print two values
  printf '%s|%s' "$lic" "$src"
}

TMP=$(mktemp)
{
  cat <<EOS
NOTICE
=======

This NOTICE file lists third‑party Go modules incorporated (directly or transitively)
into the CostScope distribution. License identifiers are detected via lightweight
heuristics scanning common license files within each module directory. Detection
is best‑effort and non-authoritative; refer to upstream artifacts for full terms.

Generation metadata:
  Generated: ${DATE_UTC}
  Commit:    ${GIT_COMMIT}
  Go:        $(go version | awk '{print $3}')

Format:
  MODULE_PATH  VERSION  LICENSE

EOS
  if [ "${NOTICE_VERBOSE:-}" = "1" ]; then
    printf '%-70s %-16s %-15s %s\n' "MODULE" "VERSION" "LICENSE" "SOURCE"
    printf '%-70s %-16s %-15s %s\n' "------" "-------" "-------" "------"
  else
    printf '%-70s %-16s %s\n' "MODULE" "VERSION" "LICENSE"
    printf '%-70s %-16s %s\n' "------" "-------" "-------"
  fi
} > "$TMP"

# Process modules (skip std library placeholders and duplicates). Sort deterministically.
echo "$MODULE_LINES" | while IFS='|' read -r path version dir indirect; do
  # Skip the main module entry for local project (list only external deps)
  if [ "$path" = "$PROJECT_MODULE" ]; then continue; fi
  # Skip empty version (should not happen for modules)
  if [ -z "$version" ]; then continue; fi
  result=$(detect_license "$dir")
  lic=${result%|*}; src=${result#*|}
  if [ "${NOTICE_VERBOSE:-}" = "1" ]; then
    [ -z "$src" ] && src="-"
    printf '%-70s %-16s %-15s %s\n' "$path" "$version" "$lic" "$src" >> "$TMP"
  else
    printf '%-70s %-16s %s\n' "$path" "$version" "$lic" >> "$TMP"
  fi
done

# Sort dependency lines (exclude header lines starting with 'MODULE' or '------')
HEADER_LINES=$(grep -n '^MODULE' "$TMP" | cut -d: -f1 | head -n1 || echo 0)
if [ "$HEADER_LINES" -gt 0 ]; then
  # Extract header (first HEADER_LINES+1 lines inc separator) and body separately
  head -n $((HEADER_LINES+1)) "$TMP" > "$TMP.header"
  tail -n +$((HEADER_LINES+2)) "$TMP" | sort -k1,1 -f > "$TMP.body"
  cat "$TMP.header" "$TMP.body" > "$TMP.sorted"
  mv "$TMP.sorted" "$TMP"
fi

cat <<'EOS' >> "$TMP"

Notes:
  * "Unknown" entries indicate the heuristic matcher did not conclusively map a license.
    This can occur if the module embeds a non-standard license filename, has only references
    to a parent licensing file, or the pattern set here is incomplete. Review upstream repo.
  * Dual / multi-licensed modules may appear with a single detected identifier; consult the
    original license text for full options and obligations.
  * This file is regenerated via scripts/license/gen-notice.sh and guarded in CI. Edit manually
    only if adding clarifying commentary – dependency lines will be overwritten on regeneration.

EOS

mv "$TMP" "$OUTPUT_FILE"
echo "NOTICE written to $OUTPUT_FILE (dependencies: $(grep -Ev '^(NOTICE|=|^$|MODULE|------|Notes:)' "$OUTPUT_FILE" | wc -l | tr -d ' '))"
