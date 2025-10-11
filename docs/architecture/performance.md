---
title: Architecture - Performance & Scalability
description: Memory, concurrency, caching, and scaling strategies.
---

# Performance & Scalability

## Strategies
- Streaming conversion to reduce peak memory.
- Concurrency via goroutine orchestration & bounded worker pools.
- Caching for normalization & classification decisions.
- Columnar analytics (DuckDB) + Parquet for IO efficiency.

## Benchmarks
Thresholds enforced via perf guard (unified vs fast mapper ratios).

## Scalability Patterns
Horizontal scaling (stateless API + distributed workers) with persistence externalization.

## See also
- `data-layer.md`
- `core-services.md`
- `api-layer.md`
