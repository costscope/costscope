# Security Policy

We take the security of CostScope and its users seriously. Thank you for helping keep the ecosystem safe.

## Reporting a Vulnerability

PRIVATE disclosure only – do not create a public issue for sensitive security reports.

1. Email: dev404ai@gmail.com
2. Subject: "[SECURITY] <short summary>"
3. Include:
   - Affected components / paths
   - Version / commit hash tested
   - Reproduction steps (minimal PoC)
   - Impact assessment (confidentiality / integrity / availability)
   - Suggested remediation (if any)
4. (Optional) Attach POC exploit as a password‑protected archive and provide the password in a separate email.

We will acknowledge receipt within **72 hours** (usually much faster). If you have not received a reply by then, please re‑send or ping via a non‑public channel.

## Coordinated Disclosure
- We strive to triage within 5 business days and provide an initial remediation plan or status.
- Fixes are backported to supported release lines where feasible.
- We aim to release a patch within 30 days for high severity issues (faster if actively exploited).
- Public disclosure occurs after a fix & release are available, or by mutual agreement.

## PGP / Encryption
A project PGP public key is not yet published.

We plan to publish an official project PGP public key to allow encrypted vulnerability reports. Current status:

- Responsible owner: `dev404ai@gmail.com` (security team)
- Current status: key generation / approval in progress (may require organizational sign-off)
- ETA for publication: TBD — contact the security team to request an expected publication date or an out-of-band key exchange

When published we will include:

- the armored public key block and a link to a public keyserver (where applicable)
- the key fingerprint (SHA256 and long key ID)
- brief instructions for encrypting vulnerability reports to the key

Until the official key is published, plain-text email to `dev404ai@gmail.com` is accepted for private disclosures. If you require encrypted submissions sooner, contact the security team via the same address and we will arrange an out-of-band key exchange.

## Scope
In‑scope: core application code, CLI, API server, configuration handling, security middleware, RBAC, conversion pipelines. Out‑of‑scope: third‑party dependencies (report upstream), demo data, test fixtures, build scripts (unless leading to supply chain risk).

## Exclusions
Reports solely about: missing security headers on non‑production demo instances, rate limiting thresholds, or best‑practice suggestions without a demonstrable vulnerability may be closed as informational.

## Safe Harbor
Good‑faith security research conforming to this policy will not initiate legal action. Avoid privacy violations, service degradation, or data exfiltration beyond minimally necessary proof.

## Handling Sensitive Data in Reports
Please redact or mask customer / tenant identifiers. Provide synthetic examples where possible.

## Credit
We credit researchers in release notes unless anonymity is requested. Provide the name/handle for attribution.

## Contact Updates
If the security contact changes, we will update this file in the primary branch and note it in the CHANGELOG.

Thank you for responsibly disclosing vulnerabilities and helping improve CostScope's security posture.

## Active Supply Chain & Security Scanners

The automated security program runs multiple scanners locally (via `make security` / `make security-aggregate`) and in CI (`security.yml`). High / Critical issues fail the build; Medium issues generate warnings.

| Category | Tool | Purpose | Fail Threshold |
|----------|------|---------|----------------|
| Code Vulnerabilities | `govulncheck` | Go module vulnerability intelligence (Go advisory DB / OSV) | CVSS >= 7 (score) |
| Static Analysis | `gosec` | Source code security weaknesses | Severity HIGH |
| Secrets Detection | `gitleaks` | Hard‑coded secrets & credentials | Any finding |
| Dependency / Image Vulns | `trivy` (fs, image) | OS & library CVEs in source tree & container image | HIGH/CRITICAL |
| SBOM Generation | `syft` | CycloneDX SBOM (attached to releases) | n/a |
| SBOM Vuln Scan (optional) | `grype` | Secondary scan of SBOM components | HIGH/CRITICAL (non-blocking if tool absent) |
| Signing / Integrity | `cosign` | Container & artifact signing / provenance | Verification must pass when signatures present |

Artifacts produced:
- `sbom.json` (CycloneDX Syft)
- `docs/security/security-summary.md` & `docs/security/security-summary.json` (aggregated results)
- `govulncheck.json`, `gosec.json`, `trivy-fs.json`, `trivy-image.json`, `grype.json` (when applicable)

Policy summary: High / Critical findings block merges; Medium findings are advisory. Secrets always block.

## Target timelines (non-binding) for vulnerability handling

The project publishes the following target timelines to set expectations for external reporters and maintainers. These are non‑binding targets (not legal or contractual SLAs) and are provided to improve coordination and transparency.

- Acknowledgement: aim to acknowledge receipt within 72 hours for all reports
- Triage: aim to perform initial triage and severity classification within 5 business days
- Medium severity: target initial response and triage within 72 hours; target mitigation plan or patch within 30 calendar days where feasible
- High / Critical severity: target mitigation plan within 7 calendar days and a prioritized patch schedule when practicable
- Communication: provide status updates at least weekly for active reports until resolved

Notes & limitations:

- These are target timelines only and may be extended for complex issues, third-party coordination, or organization-level approvals (for example publishing encryption keys or coordinating cross-team fixes). Where feasible we will publish an expected resolution ETA after triage.
- If you believe timelines are not being met, escalate by replying to the original security email with "ESCALATE" in the subject or contact project maintainers via the private channel.

See `scripts/security/aggregate-security.sh` and `.github/workflows/security.yml` for implementation.

## Local Parity With CI Gate

To mirror the CI security gate locally (same tools, thresholds, and aggregate logic), run:

```
make security-gate
```

This installs required tools if missing, generates/reuses `sbom.json`, runs govulncheck, gosec, optional gitleaks, Trivy FS and image scans (image scan skipped if Docker isn’t available), optional Grype SBOM scan, and then aggregates results. The summary is written to:
- `docs/security/security-summary.md`
- `docs/security/security-summary.json`

Exit code is non‑zero when High/Critical issues or secrets are detected, matching CI behavior.
