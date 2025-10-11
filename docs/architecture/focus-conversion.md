---
title: FOCUS Data Conversion
description: Core functionality, features, and configuration for FOCUS v1.2 conversion.
---
## Overview
The FOCUS conversion pipeline ingests provider billing exports (AWS CUR, Azure Cost Management, GCP BigQuery export) and produces normalized Parquet conforming to FOCUS v1.2. Emphasis: correctness, deterministic transformations, and observability.

## Stages
| Stage | Description | Key Outputs |
|-------|-------------|-------------|
| Ingest | Stream CSV/Parquet source rows | Raw row counters, source file metadata |
| Normalize | Map provider-specific fields to FOCUS columns | Canonical `FocusRecord` structs |
| Discount & Credit Classification | Normalize discounts, credits, taxes | Adjusted `effective_cost`, charge categories |
| Enrichment | Optional tagging / inferred fields | Tag map population |
| Validation | Structural & semantic checks | Collect validation errors (optional) |
| Write | Parquet serialization + rotation | Output parquet shards + metrics |

## Schema
Central schema defined via `GetFocusV12Schema()`. Validation layer references same source for required/optional field enforcement ensuring parity.

## Provider Adapters
| Provider | Converter Path | Notes |
|----------|----------------|-------|
| AWS | `internal/core/focus/conversion/aws` | CUR, handles blended/unblended cost reconciliation |
| Azure | `internal/core/focus/conversion/azure(see example below)ChargeCategory` = Discount/Credit |
| GCP | `internal/core/focus/conversion/gcp` | BigQuery export, SKU price normalization |


```bash
| Normalizes discount rows into
```
## Key Environment Flags
| Flag | Purpose |
|------|---------|
| `COSTSCOPE_USE_UNIFIED_MAPPER` | Enable unified mapper for parity burn-in |
| `COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION` | Skip Azure discount category transform |
| `PARITY_CHECK` | Emit internal aggregate comparison logs |

## Rotation Strategy
Writer may rotate based on approximate size (`--rotate-size` / `ParquetOptions.RotateSizeBytes`). Large exports split into sequentially numbered files; parity/invariants pick latest when unspecified.

## Determinism
- No random ordering: stable loops over provider record sets
- Time truncated to millisecond precision to avoid serialization jitter
- Aggregate parity compares effective cost, usage quantity, record count

## Validation & Invariants
Invoke after conversion:
```bash
costscope focus validate focus.parquet --all --output validation.json
costscope invariants regenerate focus.parquet --output invariants_current.json --tolerance 0.01
```
CI guards drift against baseline invariants + parity metrics.

## Metrics (Excerpt)
- `costscope_conversion_rows_total{provider}`
- `costscope_conversion_duration_seconds` (histogram)
- `costscope_writer_rotations_total`
- `costscope_writer_bytes_written_total`
- `costscope_parity_mismatch_total` (when parity check active)

## Error Handling
Non-fatal row parse errors counted & optionally surfaced in validation report. Fatal schema mismatch aborts conversion with non-zero exit.

## Observability Hooks
Tracing spans (if OTEL configured): ingest → normalize → discount_classification → write. Span attributes include row counts and rotation sequence.

## Performance Considerations
- Streaming ingestion reduces peak memory footprint
- Discount classification uses inexpensive substring heuristics; toggle env to diagnose overhead
- Unified mapper overhead guarded by perf benchmarks (see dev docs)

## Future Enhancements
- Columnar intermediate representation to reduce struct allocations
- Adaptive rotation based on time + size hybrid trigger
- Multi-tenant isolation enforcement (TenantID) when feature activated

Document updates required when schema version bumps (FOCUS 1.3+).
