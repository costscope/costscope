---
title: Glossary
description: Core terminology used across CostScope documentation and codebase.
last_reviewed: 2025-09-08
---

# Glossary

| Term | Definition |
|------|------------|
| FOCUS | FinOps Open Cost and Usage Specification – standardized cost & usage dataset schema adopted by CostScope as the canonical output format. |
| Converter (Mapper) | Streaming component that reads provider billing export rows and emits normalized FOCUS records (handles discount attribution, normalization, tagging). |
| Unified Mapper | Next-generation mapper path consolidating the previous fast mapper + experimental enrichments under a single code path with parity & performance guardrails. |
| Fast Mapper | Previous high‑performance baseline conversion path used for regression and parity comparison against the unified mapper. |
| Parity (Data Parity) | Guarantee that key aggregates (e.g., effective_cost, usage_quantity, records) match between fast and unified paths within strict bounds; enforced via (see example below). |
| Performance Bench (Perf Bench) | Benchmark suite comparing duration & allocations between unified and fast mapper paths with configurable ratio thresholds (default 1.15 duration / 1.20 alloc). |
| Invariants | Deterministic data quality checks (record counts, cost/usage sums, distribution sanity) executed during conversion to catch drift early. |
| Invariants Drift | Condition where current conversion's invariant metrics fall outside historical or expected ranges, triggering investigation. |
| SLO | Service Level Objective – agreed target for a reliability indicator (latency, availability, error rate) defining error budget consumption. |
| Error Budget | Allowed unreliability window (1 - SLO target) consumed by incidents or degradations; drives alerting & release gating decisions. |
| RBAC | Role-Based Access Control – authorization model; currently simple internal model with roadmap toward Casbin-based policy enforcement. |
| Casbin Migration | Planned evolution from simple RBAC towards a policy engine (Casbin) enabling richer role/tenant conditions and centralized evaluation. |
| JobStore | Persistence layer (current or planned) for asynchronous conversion job metadata (status, timing, outcome). |
| Parquet Rotation | Strategy to split output into multiple Parquet files by size or time for downstream query efficiency and bounded memory footprint. |
| Supply Chain Guard | Collection of build & release security measures (SBOM, checksums, reproducible builds, signature verification). |
| Contract Guard | CI mechanism failing the pipeline on breaking OpenAPI diffs to maintain API stability. |
| Placeholder Guard | Make target preventing merge if migration placeholder markers remain in docs. |
| Link Check | CI validation ensuring local markdown links resolve to existing files. |
| Perf Parity | Combined assurance of performance ratio thresholds and data parity invariants before accepting a change touching conversion paths. |


```bash
make parity-check
```
Add new terms here when introducing new subsystems or governance concepts. Keep concise; link to deeper docs where appropriate.
