# Contributing Guide

Thank you for contributing to CostScope! This document explains how to set up your environment, propose changes, and get them merged following project quality, security, and release standards.

## Quick Start Checklist

- Fork & clone the repo
- Create a well‑named branch (see Branch Naming) off `main`
- Make focused, incremental commits (see Commit Style)
- Add/adjust tests & docs
- Run quality gates locally (`make quality`)
- Open a PR using the template; link any related issues
- Respond to review feedback promptly

## Branch Naming

Use a lower‑case, slash‑separated prefix describing the change type:

```
feature/<short-slug>
fix/<issue-number>-<slug>
docs/<slug>
perf/<slug>
refactor/<slug>
chore/<slug>
security/<slug>
release/<version>            # used by maintainers
```

Examples:

- `feature/unified-mapper-cache`
- `fix/342-null-pointer-streaming`
- `security/sanitize-provider-input`

## Commit Style

We follow a pragmatic Conventional Commits subset for clarity and changelog automation:

```
<type>(<optional scope>): <short imperative summary>

Body (optional) – explain WHAT & WHY, not HOW.

Footer (optional):
BREAKING CHANGE: <description>
Refs: #123
```

Allowed types: `feat`, `fix`, `docs`, `perf`, `refactor`, `test`, `chore`, `build`, `ci`, `security`.
Multi‑part work should be squashed before merge unless each commit is independently valuable.

Examples:

```
feat(conversion): add Azure unified mapper parity metrics
fix(streaming): guard nil writer during rotation error path (#412)
security(api): enforce tenant header for non-admin roles
```

## Workflow

1. Sync local `main`: `git fetch origin && git checkout main && git pull`.
2. Branch: `git checkout -b feature/<slug>`.
3. Implement changes with tests/docs.
4. Run quality gates: `make quality` (minimum: `make test lint duplicates`).
5. If you changed a command spec (`command_spec.yaml`) run `make gen-commands` and commit updated `zz_generated_*` files. CI job `CLI Command Drift Guard` (`cli-drift`) will fail with message:
   `Run make gen-commands and commit updated zz_generated_* files` when drift is detected.
6. Ensure no secrets / credentials are added (`git grep` for accidental keys, run `make security`).
7. Open PR with concise description & rationale.
8. Address review feedback; rebase if requested.

## Local Tooling & Commands

- Build: `make build`
- Tests: `make test` / coverage `make test-coverage`
- Lint / static analysis: `make lint`
- Security scan: `make security`
- Duplicates gate: `make duplicates-gate`
- Complexity gate: `make complexity`
- Performance (critical areas): `make perf-bench-synth` or `make perf-parity`

Aim for all gates GREEN before opening a PR. CI enforces a subset; failing locally first saves iteration time.

## Declarative Integration CLI (Registrar)

Integration commands (webhook, dashboard, connections) are defined in a single specs file:

- `cmd/modules/integration/integration_action_specs.go`

Add a new command:

1. Append an `ActionSpec` (copy a nearby example).
2. Implement any new runtime logic in the appropriate manager (dashboard/webhooks/etc). Keep handlers small and pure when possible.
3. Run `go test ./...`. If the snapshot test fails, inspect diff and intentionally update the snapshot (follow test helper instructions) if the change is expected.
4. Re-run tests until green.

## Duplication Gate

We enforce a duplication baseline using `dupl` (token threshold 50). The gate runs in CI and locally with:

```
make duplicates-gate
```

If you _reduce_ duplication, also lower `DUPL_MAX_GROUPS` in the `Makefile` to the new clone group count (never raise casually).

## Commit Hygiene

See Commit Style above. Keep commits scoped & reversible. Large refactors: call out in PR description with migration notes.

## Tests

- Table-driven unit tests preferred.
- Add at least 1 happy path + 1 failure/edge case for new logic.
- Keep test fixtures small & deterministic.

## Performance

If touching conversion or unified mapper logic, run perf bench:

```
make perf-bench-synth
```

Investigate if ratios regress near thresholds.

## Security

Never commit secrets. Follow `docs/SECRETS_MANAGEMENT.md`.
If fixing a vulnerability:

- Prefer a dedicated `security/<slug>` branch
- Avoid revealing exploit details in public issue before patch release
- Coordinate with maintainers (see `SECURITY.md` for private email)

## Documentation

Update relevant docs (README sections, architecture, code health) when behavior or workflows change.

## GitHub Actions Pinning Policy

For supply-chain security, all GitHub Actions in `.github/workflows` must be pinned to immutable commit SHAs (40 hex chars) for critical actions. Non-critical actions may use SemVer tags but are encouraged to pin.

- Critical actions (must pin to commit SHA):
  - actions/checkout, actions/setup-go, actions/cache
  - actions/upload-artifact, actions/download-artifact
  - docker/setup-buildx-action, docker/login-action, docker/metadata-action, docker/build-push-action, docker/setup-qemu-action
  - aquasecurity/trivy-action, sigstore/cosign-installer, zaproxy/action-baseline, azure/setup-helm
  - gitleaks/gitleaks-action, securego/gosec, actions/github-script, marocchino/sticky-pull-request-comment, softprops/action-gh-release
- Enforcement: the `Workflow Audit` workflow fails if critical actions aren’t pinned to SHAs, or if risky refs like `@master`, `@main`, `@latest` are used, or if `curl | sh` installers appear (outside of the audit itself).
- Pinning helper: run `bash scripts/tools/pin-actions.sh` to rewrite `owner/repo@tag` to `@<commit-sha>`. Use `--dry-run` to preview.
- Optional automation: trigger the `Pin GitHub Actions` workflow to run the script and open a branch with changes.

### API Response Helpers Policy

All handlers must use the standardized envelope helpers (e.g. `response.AutoOK200`, `AutoBadRequest`, `AutoBadRequestCode`, `AutoCreated201`, `AutoNotFound404`, `AutoNoContent204`). Raw `c.JSON(400|201|404|204, ...)` calls in handlers are blocked by `scripts/tools/check-api-response-helpers.sh`. 204 responses MUST NOT include a body. When adding a new 4xx variant prefer `AutoBadRequestCode` with an existing code; introduce new codes sparingly and document them in `docs/api/index.md`.

## Release Process (Maintainers)

Releases use semantic versioning (MAJOR.MINOR.PATCH).

1. Ensure `main` is green (tests, perf guards, security scan).
2. Draft changelog section (group by feat/fix/perf/docs/security/breaking).
3. Run: `make release-promo RELEASE_VERSION=X.Y.Z` (build → sign → SBOM → smoke → tag).
4. Verify artifacts & container images published.
5. Publish GitHub Release with generated notes + highlights.
6. Announce (website / social / mailing list as applicable).

Patch releases: only fixes & security. Minor: backward compatible features. Major: breaking changes (must include migration notes and `BREAKING CHANGE:` footer in commits).

## Communication Channels

- Issues: bug reports & feature requests
- Discussions (if enabled): Q&A, design proposals
- Security: private email (see `SECURITY.md`)
- PR comments: implementation review

## Data Parity & Invariants Guard

Every PR that may affect FOCUS conversion or mapping correctness is protected by the Data Parity CI Guard (see `.github/workflows/data-parity-guard.yml`). It ensures:

1. Aggregate parity (effective_cost, usage_quantity, record_count) between the optimized fast path and the experimental unified mapper.
2. Stable high‑level dataset quality invariants (cost/usage sums, distributions, negative usage counters, distinct counts) relative to a checked‑in baseline derived from `tests/perf/aws-cur-synth.csv.gz`.

Workflow Steps (current refactored flow):

1. Build slim binary (fast/unified conversions) and both DuckDB binaries: optimized + debug (the guard prefers debug for invariants to avoid intermittent segfaults observed in the optimized build's inline invariants path).
2. Convert synthetic dataset twice:
   - Fast path → `focus_fast.parquet` (large rotate size; may still emit `focus_fast-YYYYMMDD-HHMM-001.parquet`).
   - Unified mapper (env `COSTSCOPE_USE_UNIFIED_MAPPER=1`) → `focus_unified.parquet`.
3. Run parity checker (`scripts/tools/parity-check`) → `parity.json` (aggregates + lite hash). Fail on any mismatch beyond `PARITY_TOLERANCE`.
4. Invariants guard (regenerate + diff):
   - Regenerate current invariants from the latest `focus_fast*.parquet` via `costscope invariants regenerate` → `invariants_current.json`.
   - Diff against baseline with `costscope invariants diff` → `invariants.json` (contains violations array if drift).
   - Fail when violations exceed `INVARIANTS_TOLERANCE`.
5. Upload artifacts: `parity.json`, `invariants.json`, `invariants_current.json`, `focus_fast.parquet`, `focus_unified.parquet`, `invariants_engine.txt` (records which binary handled invariants to track fallback frequency).

Exit Codes:

- 0 = success
- 2 = parity mismatch
- 3 = invariants drift violation

Environment Variables:

- `PARITY_TOLERANCE` (float, default `1e-9`) – relative tolerance for aggregates & record count.
- `INVARIANTS_TOLERANCE` (float, default `0.01`) – relative (sums) / absolute percentage‑point (distributions) drift allowance.
- `INVARIANTS_BASELINE` – override path to the baseline JSON.

Make Targets:

- `make parity-json` – fast/unified conversions + parity.json.
- `make invariants-guard` – regenerate + diff invariants (debug DuckDB preferred, optimized fallback).
- `make data-parity-guard` – combined sequence.
- `make invariants-update-baseline` – convert fast path, auto-detect rotated parquet, regenerate baseline with fallback.

Baseline Update Procedure (only when intentional changes accepted):

```bash
make build-slim build-optimized-duckdb
make invariants-update-baseline INVARIANTS_TOLERANCE=0.01
git add tests/fixtures/quality/baseline_synth_invariants.json
git commit -m "chore: update invariants baseline (synth dataset)"
```

Guidelines:

- Keep `PARITY_TOLERANCE` extremely tight (`1e-9`). Even small drift can signal a mapping regression.
- Provide rationale in the PR description for any baseline update (expected semantic change vs. bug fix).
- Avoid increasing `INVARIANTS_TOLERANCE` unless strictly necessary; prefer fixing root causes.
- If the optimized DuckDB build stabilizes, we may revert to inline invariants; until then the regenerate+diff path is canonical.

## Questions

Open a draft PR early or file an issue for architectural discussions.

Happy hacking!

## Modular Makefile Layout

The monolithic Makefile has been refactored into a modular layout for maintainability.

Structure:

```
Makefile              # Slim delegator (includes + meta aliases only)
mk/common.mk          # Shared variables, versions, global thresholds
mk/build.mk           # Build variants (slim, release, optimized, duckdb, enterprise)
mk/test.mk            # Test & coverage targets (package, runtime subset, pkg enforcement)
mk/perf.mk            # Benchmarks, parity & invariants guards, perf baselines
mk/gen.mk             # Code generation (commands, providers, integration docs), LOC guard
mk/quality.mk         # Lint, duplicates, deadcode, static analysis, notice, flag guards
mk/security.mk        # SBOM, vuln, secrets, signing, provenance, OPA, aggregated gates
mk/release.mk         # Release orchestration (promotion, checksums, signature)
mk/docker.mk          # Docker build/run targets
mk/tools.mk           # Dev tools build, profiling (cpu/mem), size analysis
mk/dev.mk             # Dev convenience (dev cycle, fmt, hooks, env, cleanup, backups)
```

Guidelines:

1. Add new targets to the logically closest mk/\*.mk (never expand root `Makefile`).
2. Each user‑facing target MUST have a help description comment: `target: ## One line summary`.
3. Internal / experimental targets: prefix with `_` OR omit the help comment so they stay hidden from `make help`.
4. Keep variable definitions centralized in `mk/common.mk` unless scoped; prefer uppercase names and document non‑obvious defaults.
5. When moving or renaming targets, search for CI references (`grep -R "make <target>" .github/ scripts/`). Update workflows atomically.
6. Guard logic (duplication, build flag, hotpath deps) belongs in `mk/quality.mk`; avoid scattering policy checks.
7. Supply chain & security (SBOM, signing, provenance, vuln scans) belong only in `mk/security.mk`.
8. Keep the root delegator minimal—only includes + meta aliases (`all`, `help`, `quality`, `tests`, `includes-check`).
9. Do not bypass pre‑commit / CI hooks (`--no-verify` not allowed) when restructuring mk files.
10. If a new mk file is added, ensure root `Makefile` includes it in the same PR.

Help Output:
`make help` aggregates all help comments from every included mk file via `$(MAKEFILE_LIST)`; adding a help comment is enough for discoverability—no extra registration needed.

Rationale:

- Reduces merge conflicts in large PRs.
- Clear separation of concerns → lowers cognitive load.
- Easier targeted reviews (e.g. security changes isolated in `mk/security.mk`).
- Faster drift detection in CI (build flag guard and duplication gates remain focused).

Migration Notes:
Previous scripts referencing old monolithic line numbers must be updated to pattern-based searches across `mk/*.mk`.

Adding a New Area:

1. Create `mk/<area>.mk`.
2. Populate `.PHONY` declarations + targets with help comments.
3. Add `-include mk/<area>.mk` line (keep alphabetical or logical grouping) to root `Makefile`.
4. Run `make help` to validate visibility, then `make quality`.

Questions: open a PR or discussion referencing this section.

## Allowlist Rationale Policy (.deadcode-allowlist)

We maintain a `.deadcode-allowlist` for symbols intentionally exempt from dead‑code removal. To prevent silent drift every non‑comment, non‑blank line MUST include an inline justification comment with the token `# rationale:`:

```
MyUnusedType # rationale: required by plugin reflection until v1 adapter removed (ISSUE-421, owner:platform)
```

Rules (enforced by `make allowlist-lint`):

1. R1: Substantive line must contain `# rationale:` (case-insensitive).
2. R2: Text after the colon must be non‑empty and at least 5 chars (configurable via `--min-len`).
3. R3: Blank lines / lines starting with `#` or `//` ignored.
4. R4: Optional waiver marker `# allowlist-ignore-rationale` skips a line (use sparingly, still explain upstream in PR description).
5. R5: First token only evaluated if multiple `# rationale:` occurrences.

CI Integration:

- Added target: `make allowlist-lint` (can be inserted before `make quality`).
- Fails the build if any violation is present (exit code 1).

Example failure output:

```

```

Allowlist rationale lint failed. Missing rationale comments for symbols:

- PreviousParser
- OldAdapter
  Each line must include '# rationale: <reason>' after the symbol.

Remediation Checklist:

- Confirm the symbol is still needed; if not—delete instead of annotating.
- Add a concise rationale (<120 chars) incl. ticket or owner.
- Prefer actionable phrasing ("until migration X", "needed for Y parity test").

Governance Goal: keep the allowlist minimal, auditable, and time‑bounded. Periodically review entries whose rationale contains temporal qualifiers ("until", target version) and remove outdated ones.

## Lint Suppression Policy (`//nolint:unused`)

We intentionally keep a very small number of `//nolint:unused` directives to preserve forward‑looking API shapes (stubs behind build tags, phased rollouts, parity helpers). To avoid accumulation of silent dead code the following policy applies:

Allowed cases (add a short rationale comment immediately after the directive):

1. Build‑tag gated stubs (e.g. enterprise / experimental) required for interface stability.
2. Transitional feature scaffolding with an approved issue or roadmap reference.
3. Parity/testing helpers that will be invoked only from scripts until promotion.

Disallowed cases:

- Generic TODOs without an associated issue.
- Previous code kept "just in case" without a clear near-term activation path.

Formatting requirement (CI enforced):

```
//nolint:unused // <concise reason (≤120 chars) + optional issue ref like #123>
```

Checklist when adding a new suppression:

- Confirm it does not mask an actual reachable unused symbol (prefer deleting instead).
- Provide rationale + (issue|roadmap) reference.
- Add/Update tests if the stub influences resolution logic.
- Open/Link the tracking issue that will remove or activate the code.

Removal: Once the symbol is referenced in normal build paths or deemed unnecessary, remove both the code and the directive in the same PR.

Guard: `make lint-nolint-guard` fails if any `//nolint:unused` lacks the trailing rationale comment.
