#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

# Ensure we execute from the repository root so relative paths resolved by
# git ls-files or find match actual files when run under containers/act.
ROOT_DIR="$(ci::repo_root "${SCRIPT_DIR}/../..")"
cd "${ROOT_DIR}"
ci::log "PWD=$(pwd)"

ci::require_cmd shellcheck

# Discover shell scripts: prefer git ls-files if available, otherwise fallback to find
declare -a files
if command -v git >/dev/null 2>&1; then
  mapfile -d '' -t files < <(git ls-files -z -- '*.sh')
else
  mapfile -d '' -t files < <(find . -type f -name '*.sh' -print0)
fi

if [[ ${#files[@]} -eq 0 ]]; then
  ci::log "No shell scripts found to lint."
  exit 0
fi

ci::log "Linting ${#files[@]} shell scripts with ShellCheck..."
shellcheck --version || true

# Determine minimum severity: when running under act, be less strict (only errors)
# Override with SHELLCHECK_MIN_SEVERITY if desired (error|warning|info|style)
MIN_SEV=${SHELLCHECK_MIN_SEVERITY:-}
if [[ -z "$MIN_SEV" ]]; then
  if ci::is_act; then
    MIN_SEV=error
  else
    MIN_SEV=warning
  fi
fi

# Extra options passthrough (space-separated string -> array)
SC_OPTS=()
if [[ -n "${SHELLCHECK_OPTS:-}" ]]; then
  # shellcheck disable=SC2206
  SC_OPTS=(${SHELLCHECK_OPTS})
fi

# Determine whether to follow sourced files. In GitHub Actions we follow (-x),
# but under act the shellcheck container may not resolve paths properly.
FOLLOW_ARGS=()
if ci::is_act; then
  FOLLOW_ARGS=()
else
  FOLLOW_ARGS=(-x)
fi

status=0
for f in "${files[@]}"; do
  if [[ ! -f "$f" ]]; then
    ci::warn "missing file (skipping): $f"
    status=1
    continue
  fi
  # Build a readable command preview
  preview=(shellcheck -S "${MIN_SEV}" ${FOLLOW_ARGS:+"${FOLLOW_ARGS[@]}"} "$f")
  if ((${#SC_OPTS[@]})); then
    preview+=("${SC_OPTS[@]}")
  fi
  ci::log "-> ${preview[*]}"

  # Invoke with proper array expansion
  if ((${#SC_OPTS[@]})); then
    if ! shellcheck -S "${MIN_SEV}" "${FOLLOW_ARGS[@]}" "${SC_OPTS[@]}" "$f"; then
      status=1
    fi
  else
    if ! shellcheck -S "${MIN_SEV}" "${FOLLOW_ARGS[@]}" "$f"; then
      status=1
    fi
  fi
done

if [[ $status -ne 0 ]]; then
  ci::warn "ShellCheck found issues."
  exit $status
fi

ci::log "ShellCheck passed."
