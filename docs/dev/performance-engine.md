---
title: Unified Performance Engine
description: Runtime-gated performance features consolidating prior build-tag variants.
---
## Purpose
Replace divergent build-tag controlled performance variants (fast mapper, experimental unified, debug alloc tracing) with a single runtime-gated engine offering parity guarantees plus optional extended instrumentation.

## Components
| Component | Responsibility |
|-----------|----------------|
| Mapper Core | Row normalization, currency/discount classification |
| Aggregator | Effective cost, usage quantity, category rollups |
| Parity Layer | Aggregate + lite hash comparison (when enabled) |
| Metrics Emitter | Duration / allocation gauges & counters |
| Phase Orchestrator | Sequential phase execution with deadlines |

## Gating Mechanisms
| Mechanism | Variable / Flag | Effect |
|-----------|-----------------|--------|
| Unified Mapper Enable | `COSTSCOPE_USE_UNIFIED_MAPPER=1` | Switch conversion to unified path |
| Focus Engine Extended | `--use-focus-engine` | Adds extended analytics JSON block |
| Parity Compare | `PARITY_CHECK=1` | Enables inline aggregate + lite hash diff logging |

## Metrics (Excerpt)
- `costscope_mapper_rows_total{path="fast|unified"}`
- `costscope_mapper_duration_seconds{path="fast|unified"}` (histogram)
- `costscope_mapper_alloc_bytes_total{path="fast|unified"}` (counter)
- `costscope_mapper_parity_mismatch_total` (counter)

## Parity Strategy
1. Fast path executes (authoritative baseline for now)
2. Unified path executes if enabled
3. Aggregates (effective_cost, usage_quantity, records) computed for both
4. Lite hash of first N (configurable) row keys compared
5. On mismatch: structured log + metric increment; exit only if guard script invoked (CI)

## Failure Modes & Mitigation
| Scenario | Symptom | Mitigation |
|----------|---------|-----------|
| Performance regression | Bench ratio > threshold | Optimize hotspot, tighten allocations |
| Parity mismatch | CI fail (exit 2) | Inspect parity.json & aggregated differences |
| Drift in invariants | CI fail (exit 3) | Update baseline if intentional or fix mapper |
| Excess allocations | High alloc ratio but pass duration | Pre-size slices/maps, reuse buffers |

## Migration Plan
| Phase | Action |
|-------|--------|
| 1 | Introduce unified mapper behind env flag (done) |
| 2 | Establish perf + parity guards (done) |
| 3 | Burn in unified in staging (audit metrics) |
| 4 | Default unify (fast path retained as fallback) |
| 5 | Remove the fast path after sustained stability |

## Instrumentation Hooks
Expose optional pprof endpoints only in API mode with `COSTSCOPE_ENABLE_PPROF=1time` + bench tooling.

```bash
(future). For batch CLI, encourage external
```

## Benchmark Interaction
Perf docs: `dev/performance-benchmarks.md`. When tuning, run:
```bash
make perf-short
make perf-parity
```

## Future Enhancements
- Adaptive batching (dynamic row grouping) to reduce allocations
- SIMD / vectorized numeric normalization (investigate via Go assembly or generics specialization)
- Memory arena pooling for transient row structs

Keep this updated as engine phases or gating change.
