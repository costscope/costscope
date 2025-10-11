---
title: Code Health Guide
description: Reproducing complexity, duplication, and lint checks plus integration command registry details.
---
## Signals
| Area | Tool | Target |
|------|------|--------|
| Formatting | (see example below) / `golangci-lint` subset | Zero diffs |
| Static analysis | (see example below), `staticcheck` | No blocking issues |
| Security lint | `gosec` (report json) | Reviewed highs |
| Secrets scan | `gitleaks` | Zero leaks |
| Duplication | (future) gocyclo / dupl | Budget established |


```bash
go fmt
```

```bash
go vet
```
## Make Targets
 aggregates fmt-check, vet, staticcheck (non-fatal), build-slim.

```bash
make quality
```

## Adding Lint Rules
1. Evaluate signal quality (noise vs value)
2. Start in advisory mode (non-fatal)
3. Escalate to blocking after cleanup

## Integration Command Registry Notes
Generated command builders must be drift-checked via ; missing commit triggers CI failure.

```bash
make gen-commands-drift
```

## Reviewing Large PRs
- Prefer logical commits (feature, refactor, test)
- Keep generated code isolated for easier review
- Ensure docs & OpenAPI baselines updated

## Test Hygiene
Fast tests <5s; isolate longer perf/integration under explicit targets to keep default suite lean.

## Backlog
- Automated dependency license classification
- Complexity budget dashboard

Update as new quality gates appear.
