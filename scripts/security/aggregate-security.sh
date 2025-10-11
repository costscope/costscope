#!/usr/bin/env bash
set -euo pipefail
# Aggregate security scan outputs into a single Markdown summary.
# Policy: Fail build if any HIGH/CRITICAL (CVSS>=7) vulnerabilities or HIGH severity gosec issues or secrets found.
# Medium severities -> warnings (non-blocking) but listed.

OUTPUT_MD="docs/security/security-summary.md"
OUTPUT_JSON="docs/security/security-summary.json"

mkdir -p "$(dirname "$OUTPUT_MD")"

section() { echo -e "\n## $1\n" >> "$OUTPUT_MD"; }

risk_fail=0
warn_count=0
> "$OUTPUT_MD"
echo "# Security Scan Summary" >> "$OUTPUT_MD"
echo "Generated: $(date -u '+%Y-%m-%dT%H:%M:%SZ')" >> "$OUTPUT_MD"

# GOVULNCHECK
if [[ -f govulncheck.json ]]; then
  section "Go Vulnerabilities (govulncheck)"
  high=$(jq '[.vulnerabilities[]? | .osv.severity[]? | select(.score>=7) ] | length' govulncheck.json 2>/dev/null || echo 0)
  med=$(jq '[.vulnerabilities[]? | .osv.severity[]? | select(.score>=4 and .score<7) ] | length' govulncheck.json 2>/dev/null || echo 0)
  echo "High (>=7): $high  Medium (4-6.9): $med" >> "$OUTPUT_MD"
  if [ "$high" -gt 0 ] 2>/dev/null; then risk_fail=1; fi
  if [ "$med" -gt 0 ] 2>/dev/null; then warn_count=$((warn_count+med)); fi
  jq -r '.vulnerabilities[]? | select(.osv.severity[]?.score>=7) | "- HIGH: \(.osv.id) \(.osv.summary)"' govulncheck.json >> "$OUTPUT_MD" || true
fi

# GOSEC
if [[ -f gosec.json ]]; then
  section "Static Analysis (gosec)"
  hi=$(jq '[.Issues[] | select(.severity=="HIGH")] | length' gosec.json 2>/dev/null || echo 0)
  med=$(jq '[.Issues[] | select(.severity=="MEDIUM")] | length' gosec.json 2>/dev/null || echo 0)
  echo "High: $hi  Medium: $med" >> "$OUTPUT_MD"
  if [ "$hi" -gt 0 ] 2>/dev/null; then risk_fail=1; fi
  if [ "$med" -gt 0 ] 2>/dev/null; then warn_count=$((warn_count+med)); fi
  jq -r '.Issues[] | select(.severity=="HIGH") | "- HIGH: [\(.rule_id)] \(.details) (\(.file):\(.line))"' gosec.json >> "$OUTPUT_MD" || true
fi

# Secrets (gitleaks) – treat any finding as HIGH
if [[ -f gitleaks-report.json ]]; then
  section "Secrets Scan (gitleaks)"
  leaks=$(jq '.findings | length' gitleaks-report.json 2>/dev/null || echo 0)
  echo "Findings: $leaks" >> "$OUTPUT_MD"
  if [ "$leaks" -gt 0 ] 2>/dev/null; then risk_fail=1; fi
  jq -r '.findings[]? | "- SECRET: \(.Description) file=\(.File) line=\(.StartLine)"' gitleaks-report.json >> "$OUTPUT_MD" || true
fi

# Trivy FS + Image
summarize_trivy() {
  local file=$1 label=$2
  if [[ -f $file ]]; then
    section "Vulnerabilities ($label)"
    high=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity=="HIGH" or .Severity=="CRITICAL")] | length' "$file" 2>/dev/null || echo 0)
    med=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity=="MEDIUM")] | length' "$file" 2>/dev/null || echo 0)
    echo "High/Critical: $high  Medium: $med" >> "$OUTPUT_MD"
  if [ "$high" -gt 0 ] 2>/dev/null; then risk_fail=1; fi
  if [ "$med" -gt 0 ] 2>/dev/null; then warn_count=$((warn_count+med)); fi
    jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity=="HIGH" or .Severity=="CRITICAL") | "- \(.Severity): \(.VulnerabilityID) \(.PkgName)@\(.InstalledVersion) -> \(.Title)"' "$file" >> "$OUTPUT_MD" || true
  fi
}

summarize_trivy trivy-fs.json "Trivy FS"
summarize_trivy trivy-image.json "Trivy Image"

# Grype (optional)
if [[ -f grype.json ]]; then
  section "SBOM Vulnerabilities (grype)"
  high=$(jq '[.matches[]? | select(.vulnerability.severity=="High" or .vulnerability.severity=="Critical")] | length' grype.json 2>/dev/null || echo 0)
  med=$(jq '[.matches[]? | select(.vulnerability.severity=="Medium")] | length' grype.json 2>/dev/null || echo 0)
  echo "High/Critical: $high  Medium: $med" >> "$OUTPUT_MD"
  if [ "$high" -gt 0 ] 2>/dev/null; then risk_fail=1; fi
  if [ "$med" -gt 0 ] 2>/dev/null; then warn_count=$((warn_count+med)); fi
fi

section "Policy Result"
if [ "$risk_fail" -gt 0 ] 2>/dev/null; then
  echo "**STATUS: FAIL** (High/Critical issues present)" >> "$OUTPUT_MD"
else
  echo "**STATUS: PASS** (No High/Critical issues)" >> "$OUTPUT_MD"
fi

echo "Warnings (medium): $warn_count" >> "$OUTPUT_MD"

# Provide machine readable summary
jq -n --arg status "$( [ "$risk_fail" -eq 0 ] && echo PASS || echo FAIL )" --argjson warnings ${warn_count:-0} '{status:$status, warnings:$warnings}' > "$OUTPUT_JSON"

if [ "$risk_fail" -gt 0 ] 2>/dev/null; then
  echo "Security gate failed (see $OUTPUT_MD)" >&2
  exit 1
fi

echo "Security gate passed (see $OUTPUT_MD)"
