---
title: Performance Benchmarks & Profiling
description: Benchmark methodology, profiling workflow, and optimization targets.
last_reviewed: 2025-09-08
---

## Goals
Guard against performance regressions while migrating from the previous (fast) mapper to unified mapper and provide deterministic thresholds for CI.

## Key Ratios (Defaults)
| Metric | Env Override | Default Max |
|--------|--------------|-------------|
| Duration (unified / fast) | `PERF_BENCH_DURATION_MAX` | 1.15 |
| Allocations (unified / fast) | `PERF_BENCH_ALLOC_MAX` | 1.20 |

Failure occurs when observed ratio > threshold. Bench tool emits structured JSON for historical comparison.

## Primary Tool
`scripts/tools/perf-bench` (invoked via make targets) executes N iterations (default 5 unless overridden) on a dataset, computing mean duration & allocations for both paths.

## Datasets
| Source | Path | Notes |
|--------|------|-------|
| Demo AWS CUR | `demo/focus-conversion/demo-cur-data.csv` | Default quick sample |
| Synthetic | `tests/perf/aws-cur-synth.csv.gz(see example below)make perf-gen-synth` |


```bash
| Generated (20k rows) via
```
## Make Targets
| Target | Purpose |
|--------|---------|
| `perf-bench` | Regression guard on demo dataset (fails CI on regression) |
| `perf-short` | 3-iteration quick bench (local iteration) |
| `perf-bench-synth` | Bench on synthetic dataset (generates metrics file) |
| `perf-bench-full` | Extended shell script wrapper (previous vs unified) |
| `perf-bench-update-baseline` | Recompute baseline JSON (use sparingly) |
| `perf-gen-synth` | Generate synthetic dataset |
| `perf-parity(see example below)perf-short` + parity aggregates check |
| `parity-check` | Compare aggregate metrics between fast & unified (effective_cost, usage_quantity, records) |
| `parity-json(see example below)parity.json` with lite hash + aggregates; fails on mismatch |


```bash
| Runs
```

```bash
| Produce
```
## Typical Local Flow
```bash
make perf-gen-synth          # once
make perf-short              # quick sanity
make parity-check            # verify aggregates
make perf-bench              # full guard replicate CI
```

## Updating Thresholds
Adjust via environment when invoking target:
```bash
PERF_BENCH_DURATION_MAX=1.10 PERF_BENCH_ALLOC_MAX=1.15 make perf-bench
```
Persisting stricter thresholds requires documentation update + potential baseline regeneration.

## Baseline Regeneration
Use only after intentional performance change or dataset evolution:
```bash
make perf-bench-update-baseline
git add tests/perf/baseline_bench_results.json
```
Review diff; ensure improvement or acceptable shift.

## Output Artifacts
| File | Produced By | Description |
|------|-------------|-------------|
| `bench_results.json` | perf-* targets | Iteration stats + ratios |
| `perf_metrics.prom(see example below)perf-bench-synth` / full | Prometheus exposition of bench metrics |
| `tests/perf/baseline_bench_results.json` | update-baseline | Stored expected ranges (comparison) |


```bash
|
```
## CI Integration
`perf-benchparity-json` to ensure performance and correctness simultaneously.

```bash
target fails the pipeline if ratios exceed thresholds. Combine with
```

## Debugging Regressions
1. Reproduce locally with `perf-shortperf-bench`.

```bash
then
```
2. Inspect `bench_results.json` (look at per-iteration variance).
3. Run with `GODEBUG=allocfreetrace=1` (optional) or pprof heap/cpu via custom instrumentation (not default in guard).
4. Identify hotspots (mapper normalization, discount classification, join logic) and optimize.

## Micro-Optimization Guidelines
- Avoid unnecessary string allocations (reuse builders)
- Defer map allocations until needed; pre-size when known
- Stream row handling; prefer single-pass normalization
- Branch-predict frequent cases (simple charge classification) first

## Performance Engine Flag Interplay
`--use-focus-engine` may introduce additional phases; when benchmarking core mapper performance, omit the flag unless explicitly measuring overhead.

## Future Enhancements (Backlog)
- Capture p95 / p99 latency for iterations (currently mean)
- Optional CPU & heap profiles on regression
- Multi-provider mixed dataset synthetic generator

Keep this document updated when thresholds or tooling change.
