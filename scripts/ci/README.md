# CI Scripts

This directory contains CI helper scripts used by GitHub Actions and local runs (including `nektos/act`). Most scripts expect Bash (>= 4) and rely on `set -euo pipefail` for safety. Common helpers live in `lib/common.sh`.

## Notable scripts

- `helm-deploy.sh`

  - Deploy or dry-run the Helm chart.
  - Inputs:
    - `KUBECONFIG_B64` (required for non-dry-run): Base64-encoded kubeconfig. Decoded into a secure tempfile and cleaned up automatically. Not required when `HELM_DRY_RUN=true` and no cluster access is desired.
    - `IMAGE_REPO` (required): Container image repository.
    - `IMAGE_TAG` (required): Container image tag.
    - `COSTSCOPE_JWT_SECRET` (optional): Secret value passed as env.
    - `HELM_DRY_RUN` (default `false`): When `true`, runs with `--dry-run --debug`.
  - Example:

    ```sh
    # Linux (GNU coreutils)
    KUBECONFIG_B64=$(base64 -w 0 ~/.kube/config) IMAGE_REPO=ghcr.io/you/costscope IMAGE_TAG=sha-123 HELM_DRY_RUN=true bash ./scripts/ci/helm-deploy.sh

    # macOS (BSD base64)
    KUBECONFIG_B64=$(base64 < ~/.kube/config | tr -d '\n') IMAGE_REPO=ghcr.io/you/costscope IMAGE_TAG=sha-123 HELM_DRY_RUN=true bash ./scripts/ci/helm-deploy.sh

    # Pure template/lint dry-run without a cluster (no kubeconfig required)
    IMAGE_REPO=ghcr.io/you/costscope IMAGE_TAG=sha-123 HELM_DRY_RUN=true bash ./scripts/ci/helm-deploy.sh
    ```

  - Notes:
    - The deploy script tolerates whitespace pasted into the secret, but it's best to store a clean, single-line base64 value.
    - If you accidentally paste the raw kubeconfig YAML into the secret, the script will detect it and proceed, but prefer encoding it as base64.
    - When `HELM_DRY_RUN=true` and no kubeconfig is provided, the script runs `helm lint` and `helm template` locally to validate values and rendering without talking to a cluster.

- `resolve-ci-image.sh`

  - Resolves a fully qualified CI tools image. Fails closed when `CI_IMAGE_TAG` is unset unless `RESOLVE_CI_IMAGE_ALLOW_LATEST=true` or running under `act`.
  - Inputs:
    - `CI_IMAGE_REPO` (optional): Defaults to `ghcr.io/<owner>/ci-base`.
    - `CI_IMAGE_TAG` (recommended): Tag or digest.
    - `RESOLVE_CI_IMAGE_ALLOW_LATEST` (default `false`): Allow `latest` fallback.
  - Output: `image=<repo:tag>` to `$GITHUB_OUTPUT`.

- `pin-ci-tools-resolve.sh` / `pin-ci-tools-update-var.sh`

  - Resolve a CI tools image reference and pin it to a repository Actions variable via GitHub API.
  - Skips API calls under `act`.

- `smoke-containers.sh`

  - Spins up the container image(s) and checks `/health/live` and `/metrics`.
  - Inputs:
    - `IMAGE_STD` (optional): Standard image reference.
    - `IMAGE_DISTROLESS` (optional): Distroless image reference.
    - `SMOKE_HTTP_PORT` (default `8080`), `SMOKE_TLS_PORT` (default `8443`): Host ports for mapping; useful to avoid collisions.
  - Notes: Uses `openssl` for TLS smoke if available; skips TLS when missing.

- `run-e2e-and-collect.sh`

  - Runs E2E tests (default `-tags duckdb`) and collects provider reports into `e2e-artifacts/`.
  - Produces merged `e2e-artifacts/summary.json` (requires `jq`; falls back to `{}` when absent).

- `run-shellcheck.sh`

  - Lints all `*.sh` files (uses `git ls-files` when available). Accepts extra ShellCheck options via `SHELLCHECK_OPTS`.

- `build-test-matrix.sh`
  - Drives test execution for `slim|sqlite|duckdb` variants; adapts behavior under `act` to reduce OOMs and provide diagnostics.

## Conventions

- Safety: Scripts should keep `set -euo pipefail`, quote variables, and clean up tempfiles with `trap`.
- act support: Use `ci::is_act` from `lib/common.sh` to tailor behavior for local runs.
- Outputs: When running in GitHub Actions, write outputs to `$GITHUB_OUTPUT` and summaries to `$GITHUB_STEP_SUMMARY`.

## Troubleshooting

- "latest" not allowed:
  - If `resolve-ci-image.sh` fails due to missing `CI_IMAGE_TAG`, either set the tag or export `RESOLVE_CI_IMAGE_ALLOW_LATEST=true` explicitly (local only).
- E2E summary is empty:
  - Ensure `jq` is installed in your environment or use the CI base image that includes it.
- Port in use during smoke tests:
  - Override `SMOKE_HTTP_PORT` / `SMOKE_TLS_PORT` with free ports.
