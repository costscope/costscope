# CI / Workflow pinning policy

This project pins critical GitHub Actions to commit SHAs to satisfy supply-chain security guardrails.

Policy

- Critical actions (build, publish, container build, signing) MUST be pinned to an exact commit SHA.
- Non-critical utility actions MAY use semver tags but should be reviewed and documented.
- When updating a SHA: add a short comment above the `uses:` line with the source URL and date.

How to refresh a pinned action

1. Inspect the action repository (for example, `docker/build-push-action`) and find a stable commit corresponding to a release; browse the repository on GitHub to pick the SHA.
2. Replace the tag with the commit SHA in the workflow and add a comment with the source reference.
3. Open a PR and run CI; include change justification in PR description.

Tools

- Use `rg "uses: .*@" .github/workflows -n` to locate uses.
- `./.github/workflows/workflow-audit.yml` contains helpers to detect unpinned critical actions.

Example: pinning a Go tool installed with `go install`

 - The `api-contract-guard.yml` workflow previously used `go install github.com/oasdiff/oasdiff@latest`.
 - We initially pinned it to the release tag `v1.11.7`, then updated to pin the exact commit referenced by that release for stronger reproducibility:
	 - `go install github.com/oasdiff/oasdiff@fc23f9bb1b54519f4f847e1724dbd0ab894e8ec8` # source: https://github.com/oasdiff/oasdiff/commit/fc23f9bb1b54519f4f847e1724dbd0ab894e8ec8 (pinned on 2025-09-28)
 - Rationale: For supply-chain sensitive steps (API contract checking) we prefer an exact commit SHA to avoid accidental changes from tag repoints or unexpected tag updates.

How to refresh a pinned Go tool

1. Check releases at the tool's GitHub releases page and pick the newest suitable tag (e.g., `v1.12.0`).
2. If you pin by tag: update the workflow line (replace `@vX.Y.Z`) and add a comment with the release URL and today's date.
3. If you pin by SHA (recommended for critical steps):
	- Identify the release tag you want to move to, visit the release page, and copy the release commit SHA (or the specific commit you want).
	- Replace the `@<old-sha>` with `@<new-sha>` and add a comment with the commit URL and date.
	- Example: `go install github.com/oasdiff/oasdiff@fc23f9b...`  # source: https://github.com/oasdiff/oasdiff/commit/fc23f9b... (2025-09-28)
4. Run CI (or `act`) to smoke-test. Include the change rationale in the PR.

Tip: when in doubt, prefer a signed release/verified commit and record the verification step in the PR description.
