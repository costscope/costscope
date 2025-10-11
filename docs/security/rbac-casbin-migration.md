---
title: RBAC Migration to Casbin
description: Phased rollout plan and testing strategy for Casbin-based RBAC.
---
## Rationale
Move from custom file-backed RBAC store to policy model enforced by Casbin for flexible model changes (role hierarchies, attribute conditions) without code churn.

## Phases
| Phase | Action | Exit Criteria |
|-------|--------|---------------|
| 0 | Baseline: file RBAC (current) | All required permissions covered |
| 1 | Introduce Casbin behind `--casbin` flag (enterprise serve) | Dual-path smoke tests pass |
| 2 | Shadow mode (evaluate both, compare decisions) | No mismatches for 7 days |
| 3 | Promote Casbin as default | Metrics stable, audit soft denies near zero |
| 4 | Remove previous path | Code removed, docs updated |

## Configuration
Model file: `configs/rbac_model.conf.example`
Policy file: `configs/rbac_policy.csv.example`
Enable:

```bash
costscope api serve --casbin
```

## Decision Parity Testing
Add integration test comparing `CheckPermission` vs Casbin enforcer for sample matrix (resources × actions × roles). Alert on drift.

## Metrics Impact
Existing counters remain; Casbin call wrapped so latency histogram reflects underlying evaluation cost.

## Migration Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Performance regression | Benchmark enforcer call; cache roles |
| Policy drift | CI validates policy schema; lint for orphan roles |
| Partial rollout confusion | Clear flag docs; deprecate after phase 3 |

Keep updated as phases progress.
