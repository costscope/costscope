CI notes — expected benign warnings

## Context

Some CI runs, self-hosted runners, and containerized test environments emit benign warnings that can create noise in workflow logs. Two common sources we see in this repository's workflows are:

- "sudo: command not found" or similar warnings when a step uses sudo inside a container image that doesn't provide it.
- host-resolution related warnings (DNS / host aliasing) coming from containerized networking differences between GitHub-hosted runners, local `act` runs, and devcontainer setups.

## Why this matters

Noise in CI logs can hide real errors and make triage slower. We prefer keeping logs quiet for expected benign messages while still surfacing real failures.

## What we changed

To reduce noise, workflow steps that previously called `sudo` unconditionally have been modified to:

- Check for availability of `sudo` (if present, use it; otherwise fall back to local installs or skip with a clear message).
- Attempt safe fallbacks when install targets are not writable (e.g. install binaries into the workspace `./bin` and add to PATH).

Files modified:

- `.github/workflows/security-gate.yml` — guarded apt / install actions and added workspace fallback for tools.
- `.github/workflows/cicd.yml` — guarded package installs in matrix jobs and jq install checks.
- `.github/workflows/coverage-guard-production.yml` — guarded make install step.

## Recommended follow-ups

- Where possible prefer using pre-provisioned runner images with required packages (jq, make, build-essential) to avoid ephemeral installs in workflow runs.
- For container-based steps prefer using container actions or `uses: docker/...` actions that include required tools, or run apt installs inside a step that uses `runs-on: ubuntu-latest` rather than inside a container without sudo.
- If warnings persist and are confirmed benign, standardize a brief comment in the workflow step explaining the expected warning so maintainers know it is intentional.

## How to handle similar future noise

- Add short `if command -v sudo` guards before using `sudo`.
- Use `|| true` with apt-get in places where install failures are non-fatal and expected in some environments.
- Document known benign warnings in `.github/ci/CI_NOTES.md` (this file).

If you want, I can:

- Add a small CI sanity step that greps workflow logs for the specific benign messages and comments their locations.
- Try to further reduce noise by installing tools via actions (e.g., `actions/setup-node`, `actions/cache`) where applicable.
