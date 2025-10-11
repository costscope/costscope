Production coverage guard
=========================

Summary
-------
This project enforces a lightweight drift guard for the `internal/core/production` package to prevent accidental coverage regressions.

What changed
------------
- A GitHub Actions workflow was added: `.github/workflows/coverage-guard-production.ymlmake coverage-guard-production`.

```bash
(runs on push and pull_request). It executes
```
- A Make target `coverage-guard-production` runs the package tests with coverage and fails if coverage drops more than the allowed delta below the baseline.
- The per-package minimum in `coverage.minlocal/costscope/internal/core/production92`.

```bash
for
```

```bash
has been raised to
```

Baseline and policy
-------------------
- Baseline used by the guard (Makefile): 92.0
- Allowed drop (grace): 2.0 percentage points
- Effective minimum enforced by the guard: baseline - allowed = 90.0
- `coverage.minlocal/costscope/internal/core/production=92`

```bash
entry (package enforcement):
```

How the guard works (quick)
---------------------------
1. The Make target runs `go test -coverprofile=... ./internal/core/production` and writes a coverage profile.
2. It parses the profile with  and extracts the total coverage percent.

```bash
go tool cover -func=...
```
3. If the measured coverage is less than , the target exits non-zero and CI will fail the job.

```bash
baseline - allowed
```

Run locally
-----------
To run the guard locally (fast check):

```bash
make coverage-guard-production
```

CI
--
The repository contains a workflow that runs the guard on push and pull_request to `main`/`master`.

Escalation and rollback
-----------------------
- If CI fails because of coverage regression, check the failing job logs for the reported `currentmin[coverage-guard] current=... baseline=... min=...`.

```bash
vs
```

```bash
values. The Makefile prints:
```
- If the regression is legitimate (tests removed or behavior changed), either add tests to restore coverage or, when justified, open a PR that adjusts the baseline/min after team discussion.
- To temporarily bypass (not recommended), a maintainer may revert the workflow or adjust baseline; prefer adding tests and fixing the cause.

Contacts / Ownership
--------------------
- Coverage guard introduced by: engineering/QA (see project commit log).
- If coverage failures occur repeatedly, tag the platform/QA team Slack channel or raise an issue titled .

```bash
coverage-guard: production regression
```

Next steps
----------
- Monitor CI for 1–2 PR cycles. If stable, consider raising baseline further (e.g., 93+) in a follow-up PR.
- Optionally add guard to branch protection checks once stable.

Questions? Open a PR or issue referencing this doc.
