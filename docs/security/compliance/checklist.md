---
title: Compliance Verification Checklist
description: Release-time checklist covering data/privacy, security, operations, and documentation sign-off.
---
## Sections
| Area | Items |
|------|-------|
| Security | Secrets scan clean, RBAC audit soft denies reviewed |
| Data | Invariants guard pass, parity pass, schema unchanged (or version bump) |
| Performance | `perf-bench` ratios ≤ thresholds, baseline updated if intended |
| Supply Chain | SBOM generated, signatures (checksums.txt.sig) verified |
| Docs | Release notes drafted, upgrade/migration docs updated |
| Compliance | SOC2/GDPR log retention & access review complete |

## Sample Checklist (Pre-Release)
```bash
make quality
```
- [ ] OpenAPI baseline diff clean
```bash
make perf-bench
```
```bash
make data-parity-guard
```
- [ ] Placeholder migration complete (guard  passes)

```bash
make docs-placeholder-guard
```
- [ ] SECURITY.md reviewed
- [ ] Release version tagged (SemVer rules)
- [ ] CHANGELOG.md entry added

## Evidence Capture
Store artifacts: `bench_results.jsoninvariants.jsonsbom.json`, signature verification log, parity.json, OpenAPI diff output.

```bash
,
```

## Post-Release Validation
- Container smoke run on tag
- Metrics surfacing (no elevated error rates)
- RBAC denied/audit ratios unchanged

Keep synchronized with release automation evolution.
