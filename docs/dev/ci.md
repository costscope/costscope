# CI, Security, and Release Scripts

This repository centralizes CI logic into small, testable scripts. This keeps GitHub Actions YAML readable and makes local reproduction easy.

- Location

  - CI helpers: `scripts/ci/`
    - build/test matrix, E2E summary, smoke containers
    - distroless build from existing image: `build-distroless-from-image.sh`
  - Security helpers: `scripts/security/`
    - govulncheck, gosec, gitleaks, trivy fs/image, aggregate
  - Release helpers: `scripts/release/`
    - derive version, install tools, changelog, sbom/sign/attest, summary

- Local runs

  - E2E summarize: `bash scripts/ci/e2e-summarize.sh`
  - Build/test matrix locally (example): `bash scripts/ci/build-test-matrix.sh duckdb`
  - Security scans:
    - `bash scripts/security/run-govulncheck.sh`
    - `bash scripts/security/run-gosec.sh`
    - `bash scripts/security/run-gitleaks.sh`
    - `bash scripts/security/run-trivy-fs.sh`
    - `TRIVY_IMAGE=costscope:ci bash scripts/security/run-trivy-image.sh`
    - Aggregate/gate results: `bash scripts/security/aggregate-security.sh`
  - Smoke containers:
    - Build images per workflow or locally, then:
    - `IMAGE_STD=ghcr.io/<org>/costscope:smoke-<sha> IMAGE_DISTROLESS=... bash scripts/ci/smoke-containers.sh`
    - Build distroless from existing image: `bash scripts/ci/build-distroless-from-image.sh <src_image> <dst_image>`
  - Release helpers:
    - `GITCLIFF_VERSION=v0.10.0 bash scripts/release/install-git-cliff.sh`
  - `SYFT_VERSION=v1.33.0 bash scripts/release/install-syft.sh` # adjust as needed; prefer latest compatible release
    - `bash scripts/release/generate-changelog.sh vX.Y.Z`

- Version pinning (Variant B)

  - Pin tool versions in workflow `env:` and let scripts read them (e.g., `GO_VERSION`, `GOVULNCHECK_VERSION`, `SYFT_VERSION`, `GITCLIFF_VERSION`, `COSIGN_VERSION`).
  - Examples:
    - In CI/CD: `GO_VERSION: "1.24.x"`, `GOVULNCHECK_VERSION: latest`
  - In Release/Supply-chain: `SYFT_VERSION: v1.33.0`, `COSIGN_VERSION: v2.2.4`
  - Prefer action SHAs for third-party actions (docker/\*, trivy-action, cosign-installer); keep our own logic in scripts.

- Conventions
  - All scripts use `set -euo pipefail` and return non-zero for failures.
  - Invoke scripts via `bash ./scripts/...` in workflows to avoid exec-bit surprises on fresh checkouts.
  - No secrets are logged; avoid printing tokens or credentials.
