#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

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

# Extra options passthrough
SC_OPTS=()
if [[ -n "${SHELLCHECK_OPTS:-}" ]]; then
  # shellcheck disable=SC2206
  SC_OPTS=(${SHELLCHECK_OPTS})
fi

status=0
for f in "${files[@]}"; do
  ci::log "-> shellcheck -S ${MIN_SEV} -x ${f} ${SC_OPTS[*]:-}"
  if ! shellcheck -S "${MIN_SEV}" -x "${SC_OPTS[@]:-}" "$f"; then
    status=1
  fi
done

if [[ $status -ne 0 ]]; then
  ci::warn "ShellCheck found issues."
  exit $status
fi

ci::log "ShellCheck passed."
