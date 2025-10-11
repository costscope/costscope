---
title: Release Checklist
description: Structured steps and integrity gates for preparing and publishing a CostScope versioned release.
---

# Release Checklist

<!-- Canonical release checklist; previous root file retained temporarily until global review -->

## Preconditions
- [ ] CHANGELOG.md updated with 1.0.0 section (no TBD)
- [ ] No uncommitted changes (`git status` clean)
- [ ] CI green on main (tests, lint, security, perf)

## Security & Compliance
- [ ] `gitleaks` run (0 leaks)
- [ ] `gosec` (no HIGH severity) summary attached
- [ ] `govulncheck` (no reachable vulns)
- [ ] `trivy fs` (CRITICAL=0 HIGH=0)
- [ ] SBOM (`sbom-syft.json`) generated (`make sbom-syft`) & verified (`make sbom-verify`)
- [ ] NOTICE regenerated & diff-free
- [ ] Communicate any updates to `SECURITY.md` (PGP key status, SLA updates) in release notes and to the security contact; confirm that the release checklist references the current `SECURITY.md` status

## Build Integrity
- [ ] Reproducible build (`./scripts/repro-check.sh`): two builds (fixed SOURCE_DATE_EPOCH) → identical sha256
- [ ] Version stamping via ldflags (Version, Commit, BuildDate, Go)
- [ ] `costscope version` outputs all fields

### Container image metadata
- [ ] Dockerfile / Dockerfile.release embed version metadata (via ldflags or build-arg labels) and are verifiable (e.g. `docker build` with --build-arg / `docker image inspect` or running `costscope version` in the container)

## API Stability
- [ ] OpenAPI diff vs baseline (no breaking removals)
- [ ] New endpoints have backward-compatible additive changes only

## Functionality Smoke
- [ ] AWS small conversion (fast + unified) parity OK
- [ ] Azure & GCP conversion invariants pass (no violations)
- [ ] API health endpoints: /health /health/live /health/ready = 200

## Performance Guards
- [ ] Unified vs fast mapper benchmark: wall_time ≤ +20%, allocs ≤ +25%

## RBAC & Config
- [ ] RBAC baseline tests pass
- [ ] Config precedence resolver tests ≥90% coverage

## Docs
- [ ] `docs/MIGRATION_0.x_to_1.0.md` present
- [ ] README links Release & Migration docs
- [ ] Version matrix (supported Go versions) updated

## Tag & Publish
- [ ] Create git tag v1.0.0
- [ ] GitHub Release draft: notes, SBOM, checksums, binaries
- [ ] Attach NOTICE, CHANGELOG excerpt
- [ ] Publish release
- [ ] Include `docs/SECURITY_RELEASE_NOTE.md` (PGP status + target timelines) in the GitHub Release notes and confirm security contact notified

## Post-Release
- [ ] Update main branch baseline OpenAPI copies if additive changes
- [ ] Announce (website/blog/social)

---
## See also
- `promotion.md`
- `../SECURITY.md`
- `../support/faq.md`
- `../architecture/overview.md`
# Release Checklist (v1.0.0 GA)

See full checklist: security, reproducible builds, API contract guard, smoke tests and tagging. Use make targets to run gates.
