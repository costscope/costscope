---
title: Security & Supply Chain Gate
description: SBOM, vulnerability, secret, SAST, container scan and signing pipeline integration.
---

# Security & Supply Chain Gate

SBOM generation, govulncheck, gosec, gitleaks, trivy and cosign; CI gates and make targets for security enforcement.

## Core Stages
| Stage | Tool | Purpose |
|-------|------|---------|
| sbom | Syft | Produce CycloneDX SBOM (`sbom-syft.json`) |
| vuln (code) | govulncheck | Reachable Go vulnerability scan |
| vuln (deps) | trivy fs | Dependency & binary vuln scan |
| sast | gosec | Static code analysis (security smells) |
| secrets | gitleaks | Detect committed secrets |
| sign | cosign | Keyless (Sigstore) signing of checksums / images |

## Make Targets (Reference)
| Target | Action |
|--------|--------|
| (see example below) | Generate SBOM |
| (see example below) | Validate SBOM presence / format |
| (see example below) | Aggregate security scanning (pipeline hook) |
| (see example below) | Includes signing & SBOM gating |


```bash
make sbom-syft
```

```bash
make sbom-verify
```

```bash
make security-scan
```

```bash
make release-promo
```
```bash
make sbom-syft
make sbom-verify
make security-scan
make release-promo
```

## Verification Flow
1. Generate SBOM early (ensures stability of dependency graph).
2. Run secret + SAST + vuln scans in parallel.
3. Fail fast on HIGH/CRITICAL findings unless explicitly waived (policy doc TBD).
4. After build, sign artifact digests (checksums + container image).
5. Publish SBOM + signatures with release.

## Future Enhancements
- SLSA provenance emission.
- SBOM diff gating (fail on unexpected new high-risk components).
- Policy-as-code integration for waiver lifecycle.
- License allow/deny policy evaluation (OPA/Rego) with build fail on disallowed.
- Secrets scan gating in pre-commit (lightweight subset for speed).
- Vulnerability age SLA (flag stale high severity > N days).

## Additional Gate Components (Merged)
From previous security-supply-chain overview:
- SBOM diff insights (new components report)
- Parallel execution of SBOM + secret + vuln + SAST scans (fail-fast)
- Provenance attestation (cosign attest) planned toggle

## CI Workflow (Reference Outline)
1. Checkout & tool install (opa, trivy, cosign, gitleaks)
2. SBOM generation → `sbom-syft.json`
3. (Planned) License policy evaluation
4. govulncheck (reachable vulns)
5. gosec (HIGH severity gating)
6. gitleaks secrets scan
7. Build binary
8. Trivy filesystem/image scan
9. Sign artifacts (checksums, image) + (planned) provenance attest
10. Upload artifacts & reports

## Fail Criteria (Consolidated)
- Any HIGH/CRITICAL vulnerability (govuln/trivy) unless explicit waiver
- Any HIGH gosec issue
- Any secret detected (no waivers)
- Policy violation (license or custom) once enabled

---
_Merged unique content from previous `security-supply-chain.md` (removed)._

## See also
- `reproducible-builds.md`
- `../release/checklist.md`
- `../release/promotion.md`
- `../architecture/overview.md`

