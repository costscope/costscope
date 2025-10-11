#!/bin/bash
# Shared pre-commit hook (can be sourced or copied). Ensures failures stop commit.
set -euo pipefail

echo "Running pre-commit checks (shared hook)..."

fail() { echo " $1" >&2; exit 1; }

# 1. Duplicates
echo " dupl..."; make duplicates || fail "duplicates failed"
# 2. Lint (full timeout)
echo " golangci-lint..."
# If running inside local act or golangci-lint is not installed, skip lint to avoid failing the local runner.
if command -v golangci-lint >/dev/null 2>&1; then
  # Prefer /usr/local/bin (where setup installs v2) over any distro v1
  export PATH="/usr/local/bin:$PATH"
  # Ensure the binary is v2 to match .golangci.yml (version: "2")
  if golangci-lint version 2>/dev/null | grep -q "version 1\."; then
    echo "Error: golangci-lint v1 detected but .golangci.yml requires version 2. Please install v2 (inside devcontainer, rerun .devcontainer/setup.sh)." >&2
    exit 1
  fi
  # Determine target scope: changed packages vs full repo to reduce load
  scope_args=(./...)
  if git rev-parse --git-dir >/dev/null 2>&1; then
    base_ref=${BASE_REF:-origin/main}
    changed_go=$(git diff --name-only --diff-filter=ACMRTUXB "$base_ref" -- '*.go' 2>/dev/null || true)
    if [ -n "$changed_go" ]; then
      pkg_dirs=$(echo "$changed_go" | xargs -r -n1 dirname | sort -u)
      filtered=""
      for d in $pkg_dirs; do
        if ls "$d"/*.go >/dev/null 2>&1; then filtered="$filtered ./${d#./}"; fi
      done
      count=$(echo "$filtered" | wc -w | tr -d ' ')
      if [ "${count:-0}" -gt 0 ] && [ "$count" -le "${GOLANGCI_MAX_TARGET_PKGS:-20}" ]; then
        # Use directory list instead of ./... to further narrow work
        # shellcheck disable=SC2206
        scope_args=($filtered)
        echo "   ↳ golangci-lint targeting changed packages ($count)"
      else
        echo "   ↳ golangci-lint full run (changed pkgs: $count)"
      fi
    fi
  fi
  # Cap concurrency further via env if set; default to 2. Allow override with GOLANGCI_CONCURRENCY.
  export GOLANGCI_LINT_RUN_CONCURRENCY=${GOLANGCI_CONCURRENCY:-2}
  golangci-lint run --timeout=10m "${scope_args[@]}" || fail "lint issues"
else
  echo " golangci-lint not found; skipping lint (install golangci-lint to enable this check)"
fi
# 3. Vulnerabilities (advisory only). Can be tuned / skipped to keep workflow fast.
if command -v govulncheck >/dev/null 2>&1; then
  if [ "${SKIP_GOVULNCHECK:-}" = "1" ]; then
    echo " govulncheck skipped (SKIP_GOVULNCHECK=1)"
  else
    echo " govulncheck (advisory)..."
    GVC_TIMEOUT_SECS=${GOVULNCHECK_TIMEOUT:-60}
    # Collect changed Go packages (vs main) if available to narrow scope
    pkgs_arg="./..."
    if git rev-parse --git-dir >/dev/null 2>&1; then
      base_ref=${BASE_REF:-origin/main}
      changed_go=$(git diff --name-only --diff-filter=ACMRTUXB "$base_ref" -- '*.go' 2>/dev/null || true)
      if [ -n "$changed_go" ]; then
        # Map files to unique module-relative dirs that contain go files
        pkg_dirs=$(echo "$changed_go" | xargs -r -n1 dirname | sort -u)
        # Filter out directories without go files (edge cases)
        filtered=""
        for d in $pkg_dirs; do
          if ls "$d"/*.go >/dev/null 2>&1; then filtered="$filtered ./${d#./}"; fi
        done
        # If reasonable number (<=15) use targeted list
        count=$(echo "$filtered" | wc -w | tr -d ' ')
        if [ "${count:-0}" -gt 0 ] && [ "$count" -le "${GOVULNCHECK_MAX_TARGET_PKGS:-15}" ]; then
          pkgs_arg="$filtered"
          echo "   ↳ targeting changed packages ($count)"
        else
          echo "   ↳ using full scan (changed pkgs: $count)"
        fi
      fi
    fi
    # Prefer faster source mode if available (older versions ignore flag)
    gvc_cmd=(govulncheck -mode=source $pkgs_arg)
    if command -v timeout >/dev/null 2>&1; then
      if ! timeout --preserve-status ${GVC_TIMEOUT_SECS}s "${gvc_cmd[@]}" >/dev/null; then
        ec=$?
        if [ $ec -eq 124 ] || [ $ec -eq 137 ]; then
          echo "(advisory) govulncheck timed out after ${GVC_TIMEOUT_SECS}s"
        else
          echo "(advisory) govulncheck exited non-zero (code=$ec)"
        fi
      fi
    else
      # Fallback without timeout
      if ! "${gvc_cmd[@]}" >/dev/null; then
        echo "(advisory) govulncheck exited non-zero"
      fi
    fi
  fi
fi
# 4. Tests (race)
echo " tests..."; go test -race ./... || fail "tests failed"
# 5. Build
echo " build..."; go build ./... || fail "build failed"

echo " All pre-commit checks passed."
