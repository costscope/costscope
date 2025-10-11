#!/usr/bin/env sh
set -eu

# Discover shell scripts: prefer git ls-files if available, otherwise fallback to find
if command -v git >/dev/null 2>&1; then
  files=$(git ls-files | grep -E '\\.(sh)$' || true)
else
  # Use POSIX find; exclude vendor and dist directories if present
  files=$(find . -type f -name '*.sh' ! -path './vendor/*' ! -path './dist/*' || true)
fi

if [ -z "${files}" ]; then
  echo "No shell scripts found to lint."
  exit 0
fi

count=$(printf "%s\n" "${files}" | sed '/^$/d' | wc -l | tr -d ' ')
echo "Linting ${count} shell scripts with ShellCheck..."
shellcheck --version || true

# Determine minimum severity: when running under act, be less strict (only errors)
# Override with SHELLCHECK_MIN_SEVERITY if desired (error|warning|info|style)
MIN_SEV=${SHELLCHECK_MIN_SEVERITY:-}
if [ -z "$MIN_SEV" ]; then
  if [ "${IS_ACT:-}" = "true" ] || [ "${ACT:-}" = "true" ]; then
    MIN_SEV=error
  else
    MIN_SEV=warning
  fi
fi

status=0
for f in ${files}; do
  echo "-> shellcheck -S ${MIN_SEV} -x ${f}"
  shellcheck -S "${MIN_SEV}" -x "${f}" || status=1
done

if [ "${status}" -ne 0 ]; then
  echo "ShellCheck found issues."
  exit ${status}
fi

echo "ShellCheck passed."
