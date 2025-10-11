---
title: API Latency Degradation
description: Runbook for investigating and mitigating API latency regressions.
last_reviewed: 2025-09-08
---

# API Latency Degradation

## 1. Detection
Triggers:
- Warning: p95 latency > 800ms for 3 of last 5 minutes.
- Critical: p95 latency > 1500ms for 2 of last 3 minutes OR fast error budget burn (>2% in 1h).
Dashboards (TODO JSON): "API Overview" → panels: p50/p95/p99 latency, saturation, error budget.

## 2. Initial Triage (≤5 min)
1. Confirm alert validity: Compare p95 vs p50 (wide gap implies tail issue).
2. Check error rate – if 5xx > normal may be cascading failures.
3. Identify affected routes ().

```bash
topk(5, increase(costscope_http_request_duration_seconds_count[5m]))
```

## 3. Diagnosis Deep Dive
| Hypothesis | Checks | Commands / Queries |
|-----------|--------|--------------------|
| Hot route regression | Per-route p95 vs baseline | PromQL: `histogram_quantile(0.95, sum(rate(costscope_http_request_duration_seconds_bucket{route="/api/v1/focus/convert"}[5m])) by (le))` |
| DB / DuckDB saturation | Look for increased conversion duration & active jobs | `costscope_conversion_active_jobs` gauge, system metrics (CPU/IO) |
| GC pressure | Check Go GC cycles & pause (if exported) | `go_memstats_gc_sys_bytes(see example below)go_gc_duration_seconds` |
| External dependency slowness | Trace spans with elevated duration | Query tracing backend (filter latency > 1000ms) |
| Thundering herd / traffic spike | Request rate surge | `rate(costscope_http_requests_total[1m])` vs 1h baseline |
| Lock contention / new release | Recent deploy? | Check deployment timeline / release notes |


```bash
,
```
## 4. Mitigation
- If only one route impacted (e.g. conversion submit) & queueing: scale horizontally (add pod / instance) or increase worker pool limit (temporary).
- Enable adaptive concurrency (if feature flag present) to shed load gracefully.
- Roll back recent deployment if latency began immediately post-release.
- If CPU saturation >85% sustained: scale up CPU or reduce parallel job concurrency.
- If GC pressure: increase GOGC (temporary) or reduce allocation hotspots (see profiles).

## 5. Verification
Success criteria (all):
1. p95 latency back < 800ms for 3 consecutive 5m intervals.
2. Error budget burn rate normal (<0.2% / h).
3. Verify related performance guards: see `docs/dev/performance-benchmarks.md` for thresholds.

```bash
make perf-parity
```
3. No elevated 5xx or conversion ack slowdown.

## 6. Post-Incident
- Capture root cause & remediation in incident doc.
- Add test / benchmark to prevent regression (e.g. perf-bench threshold tweak).
- Update capacity planning assumptions.

## 7. Prevention Opportunities
- Autoscaling rule tied to p95 & saturation.
- Pre-deploy synthetic latency canary.
- Tighten perf bench CI guard if regression pattern repeated.

## 8. Escalation
If unresolved after 30 minutes or critical burn continues: page Platform Lead & SRE secondary.
