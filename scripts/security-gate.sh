#!/usr/bin/env bash
# Security gating script (M2) – aggregates static analysis & vulnerability scans.
# Tools:
#  1. gitleaks (secrets)
#  2. gosec (code security) – fail on HIGH
#  3. govulncheck (reachable vulns) – fail on reachable
#  4. trivy filesystem scan – fail on CRITICAL/HIGH vulns
# Output: SECURITY_REPORT.md (table) + individual JSON reports.
# Exit code: 0 if all pass, 1 otherwise.

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
REPORT_MD="${ROOT_DIR}/SECURITY_REPORT.md"

GITLEAKS_JSON="gitleaks-report.json"
GOSEC_JSON="gosec-report.json"
GOVULN_JSON="govulncheck-report.json"
TRIVY_JSON="trivy-report.json"

# Configurable thresholds (env override)
GOSEC_FAIL_LEVEL="${GOSEC_FAIL_LEVEL:-HIGH}"          # severity to fail on (only HIGH supported in logic now)
TRIVY_SEVERITIES="${TRIVY_SEVERITIES:-CRITICAL,HIGH}" # severities to gate
GITLEAKS_ALLOW_FINDINGS="${GITLEAKS_ALLOW_FINDINGS:-0}" # if 1, do not fail on leaks (still report)
FAIL_ON_MISSING_TOOLS="${FAIL_ON_MISSING_TOOLS:-1}"     # if 1, missing tool => fail

has_cmd() { command -v "$1" >/dev/null 2>&1; }

require() {
  for c in "$@"; do
    if ! has_cmd "$c"; then
      echo "[security-gate] Missing required tool: $c" >&2
      MISSING=1
    fi
  done
  if [ "${MISSING:-0}" = 1 ]; then
    echo "Install missing tools before running this script." >&2
    exit 127
  fi
}

require jq

STATUS_ANY_FAIL=0

row() { # tool status findings notes
  printf '| %s | %s | %s | %s |\n' "$1" "$2" "$3" "$4" >>"$REPORT_MD"
}

echo '# Security Gate Report' >"$REPORT_MD"
echo >>"$REPORT_MD"
echo "Generated: $(date -u '+%Y-%m-%dT%H:%M:%SZ')" >>"$REPORT_MD"
echo >>"$REPORT_MD"
echo '| Tool | Status | Findings | Notes |' >>"$REPORT_MD"
echo '|------|--------|----------|-------|' >>"$REPORT_MD"

############################################
# 1. GITLEAKS
############################################
if has_cmd gitleaks; then
  echo '[1/4] Running gitleaks...'
  # --no-git scans current directory as plain files (no git history)
  if gitleaks detect --no-git -r "$GITLEAKS_JSON" -f json >/dev/null 2>&1; then
    :
  else
    # gitleaks exits non-zero on findings; still capture report
    echo '[gitleaks] Completed with potential findings.'
  fi
  GITLEAKS_FINDINGS=$(jq 'length' "$GITLEAKS_JSON" 2>/dev/null || echo 0)
  if [ "$GITLEAKS_FINDINGS" -gt 0 ]; then
    if [ "$GITLEAKS_ALLOW_FINDINGS" = "1" ]; then
      row 'gitleaks' 'WARN' "$GITLEAKS_FINDINGS" 'Findings ignored by policy'
    else
      row 'gitleaks' 'FAIL' "$GITLEAKS_FINDINGS" 'Secrets detected'
      STATUS_ANY_FAIL=1
    fi
  else
    row 'gitleaks' 'PASS' 0 'No leaks'
  fi
else
  if [ "$FAIL_ON_MISSING_TOOLS" = "1" ]; then
    row 'gitleaks' 'FAIL' '-' 'Tool missing'
    STATUS_ANY_FAIL=1
  else
    row 'gitleaks' 'SKIP' '-' 'Tool not installed'
  fi
fi

############################################
# 2. GOSEC
############################################
if has_cmd gosec; then
  echo '[2/4] Running gosec...'
  # -fmt json for machine parsing; allow list of packages ./...
  if ! gosec -fmt json -out "$GOSEC_JSON" ./... >/dev/null 2>&1; then
    # gosec returns non-zero sometimes on issues; we still parse
    echo '[gosec] Completed with potential issues.'
  fi
  GOSEC_HIGH=$(jq '[.Issues[]? | select(.severity == "HIGH")] | length' "$GOSEC_JSON" 2>/dev/null || echo 0)
  GOSEC_TOTAL=$(jq '.Issues | length' "$GOSEC_JSON" 2>/dev/null || echo 0)
  if [ "$GOSEC_FAIL_LEVEL" = "HIGH" ] && [ "$GOSEC_HIGH" -gt 0 ]; then
    row 'gosec' 'FAIL' "$GOSEC_HIGH (high / $GOSEC_TOTAL total)" 'HIGH severity issues'
    STATUS_ANY_FAIL=1
  else
    row 'gosec' 'PASS' "$GOSEC_TOTAL" 'No HIGH severity'
  fi
else
  if [ "$FAIL_ON_MISSING_TOOLS" = "1" ]; then
    row 'gosec' 'FAIL' '-' 'Tool missing'
    STATUS_ANY_FAIL=1
  else
    row 'gosec' 'SKIP' '-' 'Tool not installed'
  fi
fi

############################################
# 3. GOVULNCHECK
############################################
if has_cmd govulncheck; then
  echo '[3/4] Running govulncheck...'
  # JSON mode for reliable parsing
  if ! govulncheck -json ./... >"$GOVULN_JSON" 2>/dev/null; then
    echo '[govulncheck] Finished (non-zero exit).'
  fi
  # Reachable vulns heuristic: objects with finding + isCalled true OR callStacks length > 0
  GOVULN_REACHABLE=$(jq '[.[] | select(.finding.vuln != null) | select((.finding.isCalled == true) or (.finding.trace? | length > 0) or (.finding.vuln.callStacks? | length > 0)) ] | length' "$GOVULN_JSON" 2>/dev/null || echo 0)
  GOVULN_TOTAL=$(jq '[.[] | select(.finding.vuln != null)] | length' "$GOVULN_JSON" 2>/dev/null || echo 0)
  if [ "$GOVULN_TOTAL" -eq 0 ]; then
    row 'govulncheck' 'PASS' 0 'No vulnerabilities reported'
  elif [ "$GOVULN_REACHABLE" -gt 0 ]; then
    row 'govulncheck' 'FAIL' "$GOVULN_REACHABLE (reachable / $GOVULN_TOTAL total)" 'Reachable vulnerabilities'
    STATUS_ANY_FAIL=1
  else
    row 'govulncheck' 'PASS' "$GOVULN_TOTAL" 'No reachable vulns'
  fi
else
  if [ "$FAIL_ON_MISSING_TOOLS" = "1" ]; then
    row 'govulncheck' 'FAIL' '-' 'Tool missing'
    STATUS_ANY_FAIL=1
  else
    row 'govulncheck' 'SKIP' '-' 'Tool not installed'
  fi
fi

############################################
# 4. TRIVY
############################################
if has_cmd trivy; then
  echo '[4/4] Running trivy filesystem scan...'
  TRIVY_CMD=(trivy fs --security-checks vuln --severity "${TRIVY_SEVERITIES}" --format json -o "$TRIVY_JSON" .)
  # We want exit code semantics (exit 1 on findings) but still parse JSON even if it fails.
  if ! "${TRIVY_CMD[@]}" >/dev/null 2>&1; then
    echo '[trivy] Scan completed (potential findings).'
  fi
  # Count HIGH + CRITICAL from JSON severity array
  TRIVY_JQ_SEV=$(echo "$TRIVY_SEVERITIES" | sed 's/,/" or .Severity == "/g')
  TRIVY_COUNT=$(jq "[.Results[]?.Vulnerabilities[]? | select(.Severity == \"${TRIVY_SEVERITIES%%,*}\" or .Severity == \"${TRIVY_SEVERITIES#*,}\")] | length" "$TRIVY_JSON" 2>/dev/null || echo 0)
  if [ "$TRIVY_COUNT" -gt 0 ]; then
    row 'trivy' 'FAIL' "$TRIVY_COUNT" "$TRIVY_SEVERITIES vulns"
    STATUS_ANY_FAIL=1
  else
    row 'trivy' 'PASS' 0 "No ${TRIVY_SEVERITIES}"
  fi
else
  if [ "$FAIL_ON_MISSING_TOOLS" = "1" ]; then
    row 'trivy' 'FAIL' '-' 'Tool missing'
    STATUS_ANY_FAIL=1
  else
    row 'trivy' 'SKIP' '-' 'Tool not installed'
  fi
fi

echo >>"$REPORT_MD"
if [ "$STATUS_ANY_FAIL" -eq 0 ]; then
  echo 'Overall Status: PASS' >>"$REPORT_MD"
  echo '[security-gate] SUCCESS: all checks passed.'
else
  echo 'Overall Status: FAIL' >>"$REPORT_MD"
  echo '[security-gate] FAILURE: one or more checks failed.' >&2
fi

echo >>"$REPORT_MD"
echo 'Artifacts:' >>"$REPORT_MD"
echo "- $GITLEAKS_JSON" >>"$REPORT_MD"
echo "- $GOSEC_JSON" >>"$REPORT_MD"
echo "- $GOVULN_JSON" >>"$REPORT_MD"
echo "- $TRIVY_JSON" >>"$REPORT_MD"

exit "$STATUS_ANY_FAIL"
