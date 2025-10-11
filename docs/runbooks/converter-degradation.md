---
title: Converter Throughput / Error Degradation
description: Runbook for identifying and mitigating converter throughput slowdown or elevated error ratios.
last_reviewed: 2025-09-08
---

# Converter Throughput / Error Degradation

## 1. Detection
Alerts:
- Converter Error Ratio >0.5% for 15m.
- Conversion Ack p95 >5s for 15m & >2x 7d moving baseline.
- Unified vs unified-experimental duration ratio >1.15 (guard rail) during opt-in tests.

## 2. Initial Triage
1. Identify provider(s) affected (`provider` label in duration histogram).
2. Check `costscope_converter_errors_totalcostscope_converter_records_total`.

```bash
delta vs
```
3. Inspect recent code changes in `internal/core/focus/conversion/`.

## 3. Diagnosis
| Suspect | Evidence | Action |
|---------|----------|--------|
| Input format change (provider) | Spike in parse errors, logs referencing missing column | Capture sample row; add defensive mapping |
| Performance regression (allocation) | Increased duration but normal errors | Compare perf bench results; run (see example below) |


```bash
make perf-parity
```
```bash
make perf-parity
```
| Unified mapper overhead | Unified path ratio > threshold | Temporarily disable unified mapper (flag/env) |
| I/O bottleneck (storage) | High wait / slow parquet rotation | Check rotation span durations & disk metrics |
| Memory pressure (chunk size) | Elevated GC & pauses | Reduce `--chunk-size` temporarily |
| Hot tagging / normalization path | CPU profile shows string ops heavy | Cache canonicalization map |

## 4. Mitigation
- Roll back to prior binary if regression introduced in last release.
- Temporarily reduce concurrency (workers) if causing thrash; or increase if under-utilizing CPU.
- Disable unified mapper by removing flag/env if slowing pipeline.
- Patch quick fix for missing field with default fallback; schedule full spec update.
- If error isolated to single provider: pause that provider's conversions and communicate partial outage.

## 5. Verification
Criteria:
1. Error ratio <0.5% for 3 consecutive 5m windows.
2. Ack p95 ≤5s returning to baseline ±10%.
3. Unified vs experimental duration ratio ≤1.15 again (if enabled).
4. Re-run the local parity + perf check to validate parity; consult `docs/dev/performance-benchmarks.md` for changing thresholds.

```bash
make perf-parity
```

## 6. Post-Incident
- Add regression test or dataset to perf bench synthetic corpus.
- Extend mapper unit tests for new field scenario.
- Document any provider format change in CHANGELOG & internal notes.

## 7. Prevention
- Add anomaly detection on `costscope_converter_duration_seconds` trend.
- Increase sampling of input schema validation early.
- Evaluate adaptive chunk size logic (future enhancement).

## 8. Escalation
If >2h sustained degradation or backlog > SLA threshold: escalate to Data Processing Lead + SRE.
