---
title: Release Promotion Pipeline
description: Automated multi-stage pipeline for building, validating, and publishing a release.
---

# Release Promotion Pipeline

<!-- Canonical promotion pipeline doc; previous root file retained temporarily until global review -->

## Stages
1. build – Production binary build (`build-production`).
2. sign – Cosign sign image if `cosign` present (optional).
3. sbom – CycloneDX SBOM generation (`sbom`).
4. smoke – Functional smoke (`scripts/release/smoke.sh`):
   - focus convert demo CSV → Parquet
   - validate produced Parquet
   - start API and poll `/health/ready` + `/health`
5. stage – Tag container `costscope:stage-<version>`.
6. promote – Tag container `costscope:<version>`.
7. prod – Annotated git tag `v<version>` + release notes + checksums.

Security gate (aggregated: SBOM, govulncheck, gosec, secrets) must pass before build proceeds.

## Usage
```bash
make release-promo RELEASE_VERSION=1.4.0
```
Environment overrides:
- IMAGE – override docker image name (default `costscope:latest`).
- SMOKE_API_PORT – override API port (default 18080).
- SMOKE_START_TIMEOUT – seconds to wait for readiness (default 20).

## Artifacts
Generated in repository root:
- `bin/costscope-production`
- `sbom.json`
- `checksums.txt` (sha256)
- `release_notes.md`
- Container image tags: `latest`, `stage-<version>`, `<version>`

## Release Notes Generation
`make release-notes RELEASE_VERSION=X.Y.Z` or implicit in `release-promo`.
Script diffs commits since previous tag (chronologically latest). Template sections for summary, upgrade notes.

## Extending
Add additional gates before promotion by appending to `release-promo` target (e.g., perf benchmarks, policy eval).

## Exit Codes
- Non-zero: pipeline failure (step logged with emoji prefix). Troubleshoot and re-run after fix.

## Future Enhancements
- Push tags & images (CI context) when env `PUSH=1`.
- SLSA provenance auto-generation.
- Multi-arch image build + SBOM diff gating.

---
## See also
- `checklist.md`
- `../SECURITY.md`
- `../support/faq.md`
- `../architecture/overview.md`
# Release Promotion Pipeline

End-to-end automated promotion pipeline with build -> sign -> sbom -> smoke -> promote phases. Use `make release-promo RELEASE_VERSION=x.y.z`.
