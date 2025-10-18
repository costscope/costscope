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

# Derive Go major.minor from go.mod for deterministic header when requested
GO_VERSION_FULL=$(go version | awk '{print $3}')
# Default minor string from tool output (e.g., go1.24.7)
GO_VERSION_MINOR="$GO_VERSION_FULL"
if [ -f go.mod ]; then
  # Parse `go` directive; expect forms like `go 1.24.6` or `go 1.24`
  if GO_MOD_VER=$(awk '/^go [0-9]+\.[0-9]+(\.[0-9]+)?/ {print $2; exit}' go.mod); then
    # Keep only major.minor
    GO_MM=$(echo "$GO_MOD_VER" | awk -F. '{print $1"."$2}')
    if [ -n "$GO_MM" ]; then
      GO_VERSION_MINOR="go${GO_MM}.x"
    fi
  fi
fi

# Deterministic mode (for drift guard): suppress volatile metadata so repeated
# regenerations without dependency changes yield identical NOTICE.
if [ "${NOTICE_DETERMINISTIC:-}" = "1" ]; then
  DATE_UTC="(deterministic)"
  GIT_COMMIT="(deterministic)"
  # Pin Go header line to major.minor.x derived from go.mod to reduce noisy drift
  GO_VERSION_FULL="$GO_VERSION_MINOR"
fi

# Collect module metadata: path|version|dir|indirect
MODULE_LINES=$(go list -m -f '{{ .Path }}|{{ .Version }}|{{ .Dir }}|{{ if .Indirect }}true{{ else }}false{{ end }}' all)

pick_license_in_dir() {
  # Pick the most appropriate license file in a directory with deterministic priority.
  # Priority: LICENSE (exact), LICENSE.* (txt, md), COPYING (exact), COPYING.*, UNLICENSE, UNLICENSE.*
  local d="$1"
  local f
  for f in "LICENSE" "LICENSE.txt" "LICENSE.md" "LICENSE.rst"; do
    if [ -f "$d/$f" ]; then echo "$d/$f"; return 0; fi
  done
  for f in "COPYING" "COPYING.txt" "COPYING.md"; do
    if [ -f "$d/$f" ]; then echo "$d/$f"; return 0; fi
  done
  for f in "UNLICENSE" "UNLICENSE.txt"; do
    if [ -f "$d/$f" ]; then echo "$d/$f"; return 0; fi
  done
  # As a last resort, any LICENSE.* file
  f=$(find "$d" -maxdepth 1 -type f -iname 'LICENSE.*' 2>/dev/null | sort | head -n1 || true)
  if [ -n "$f" ]; then echo "$f"; return 0; fi
  echo ""
}

find_license_file() {
  # Ascend until a LICENSE-like file is found (or root). Prefer LICENSE*/COPYING*; avoid NOTICE.
  local start="$1"; local cur="$start"; local cand="";
  while [ -n "$cur" ] && [ "$cur" != "/" ]; do
    cand=$(pick_license_in_dir "$cur")
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
  # Prefer LICENSE*/COPYING*/UNLICENSE* in the current dir with deterministic priority
  cand=$(pick_license_in_dir "$dir")
    if [ -z "$cand" ]; then
      cand=$(find_license_file "$dir")
    fi
    # Fallback for Go submodules whose license is only present in the parent module root
    # Example: cloud.google.com/go/logging@vX may inherit license from cloud.google.com/go@vY
    if [ -z "$cand" ]; then
      # Walk up 3 ancestors; at each step, look for a sibling directory named "<basename>@v*"
      local cur="$dir"; local i=0
      while [ $i -lt 3 ]; do
        local parent; parent=$(dirname "$cur")
        local base; base=$(basename "$parent")
        local grand; grand=$(dirname "$parent")
        # Look for the parent module root directories with version suffix
        if [ -d "$grand" ]; then
          local rootcand
          rootcand=$(find "$grand" -maxdepth 1 -type d -name "${base}@v*" 2>/dev/null | sort -V | tail -n1 || true)
          if [ -n "$rootcand" ]; then
            # Check for a license file under the found root
            cand=$(pick_license_in_dir "$rootcand")
            if [ -n "$cand" ]; then
              break
            fi
          fi
        fi
        cur="$parent"; i=$((i+1))
      done
    fi
    if [ -n "$cand" ]; then
      src="$cand"
  # Read many lines for broader pattern match (licenses often have long headers)
  local snippet
  snippet=$(head -n 2000 "$cand" | tr '\r' ' ')
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
      # As a last resort, search entire file for a few common identifiers (slower but reduces Unknown)
      if [ "$lic" = "Unknown" ]; then
        if grep -qiE 'SPDX-License-Identifier:[[:space:]]*Apache-2.0' "$cand"; then lic="Apache-2.0"; fi
      fi
      if [ "$lic" = "Unknown" ]; then
        if grep -qiE 'Apache License[[:space:]]+Version[[:space:]]+2' "$cand"; then lic="Apache-2.0"; fi
      fi
      if [ "$lic" = "Unknown" ]; then
        if grep -qiE 'MIT License|Permission is hereby granted' "$cand"; then lic="MIT"; fi
      fi
      if [ "$lic" = "Unknown" ]; then
        if grep -qiE 'BSD.*(Redistribution and use in source and binary forms)' "$cand"; then lic="BSD-3-Clause"; fi
      fi
    fi
  fi
  # If still Unknown and directory exists, scan auxiliary files (README, NOTICE, go.mod) for SPDX hints
  if [ "$lic" = "Unknown" ] && [ -n "$dir" ] && [ -d "$dir" ]; then
    local aux
    for aux in "README" "README.md" "readme.md" "NOTICE" "NOTICE.txt" "go.mod"; do
      if [ -f "$dir/$aux" ]; then
        if grep -qiE 'SPDX-License-Identifier:[[:space:]]*Apache-2.0' "$dir/$aux"; then lic="Apache-2.0"; src="$dir/$aux"; break; fi
        if grep -qiE 'SPDX-License-Identifier:[[:space:]]*MIT' "$dir/$aux"; then lic="MIT"; src="$dir/$aux"; break; fi
        if grep -qiE 'SPDX-License-Identifier:[[:space:]]*BSD-3-Clause' "$dir/$aux"; then lic="BSD-3-Clause"; src="$dir/$aux"; break; fi
        # Non-SPDX hint for Apache in NOTICE files
        if [ "$lic" = "Unknown" ] && grep -qiE 'Apache License[[:space:]]+Version[[:space:]]+2' "$dir/$aux"; then lic="Apache-2.0"; src="$dir/$aux"; break; fi
      fi
    done
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
  Go:        ${GO_VERSION_FULL}

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
  # Ensure module directory is materialized; go list may leave .Dir empty when not yet extracted
  if [ -z "$dir" ] || [ ! -d "$dir" ]; then
    # Try to resolve via go mod download -json (prefers cache, no network if present)
    if command -v jq >/dev/null 2>&1; then
      resolved_dir=$(go mod download -json "$path@$version" 2>/dev/null | jq -r '.Dir // empty' || true)
    else
      # Sed fallback to capture Dir path in JSON (best-effort)
      resolved_dir=$(go mod download -json "$path@$version" 2>/dev/null | sed -n 's/.*"Dir"[[:space:]]*:[[:space:]]*"\(.*\)".*/\1/p' | head -n1 || true)
    fi
    if [ -n "$resolved_dir" ] && [ -d "$resolved_dir" ]; then
      dir="$resolved_dir"
    fi
  fi
  # As a last resort, attempt to locate the module directory in GOMODCACHE by basename@version
  if [ -z "$dir" ] || [ ! -d "$dir" ]; then
    gomodcache=$(go env GOMODCACHE 2>/dev/null || echo "")
    if [ -z "$gomodcache" ]; then
      gopath=$(go env GOPATH 2>/dev/null || echo "")
      [ -n "$gopath" ] && gomodcache="$gopath/pkg/mod"
    fi
    if [ -n "$gomodcache" ] && [ -d "$gomodcache" ]; then
      tailname=$(basename "$path")
      # Try constructing the canonical cache path first: <gomodcache>/<parent>/<basename>@<version>
      parentpath=$(dirname "$path")
      # If module is at top-level (no dirname), dirname will return '.' which we should ignore
      if [ "$parentpath" = "." ]; then
        canddir="${gomodcache}/${tailname}@${version}"
      else
        canddir="${gomodcache}/${parentpath}/${tailname}@${version}"
      fi
      if [ ! -d "$canddir" ]; then
        # Fallback: search by basename@version anywhere under cache (slower but resilient)
        canddir=$(find "$gomodcache" -type d -name "${tailname}@${version}" -print -quit 2>/dev/null || true)
      fi
      if [ -n "$canddir" ] && [ -d "$canddir" ]; then
        dir="$canddir"
      fi
    fi
  fi
  result=$(detect_license "$dir")
  lic=${result%|*}; src=${result#*|}
  # Conservative overrides for modules with atypical layouts where heuristic may mis-detect in clean containers
  case "$path" in
    gopkg.in/yaml.v2)
      lic="MIT" ;;
    github.com/antlr4-go/antlr/v4)
      lic="BSD-3-Clause" ;;
    github.com/coreos/go-systemd|github.com/coreos/go-systemd/v22)
      lic="Apache-2.0" ;;
    github.com/prometheus/client_model|github.com/prometheus/common|github.com/prometheus/procfs)
      lic="Apache-2.0" ;;
    go.etcd.io/gofail)
      lic="Apache-2.0" ;;
    github.com/kr/pretty|github.com/kr/text)
      lic="MIT" ;;
  esac
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
