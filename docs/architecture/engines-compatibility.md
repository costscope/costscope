---
title: Engines Compatibility
description: Parquet output compatibility across DuckDB, Trino, Athena, Spark.
---
## Supported Engines
CostScope emits Parquet (Snappy by default) adhering to the FOCUS v1.2 schema; files are consumable by DuckDB, Trino, Athena, and Spark without transformation.

| Engine | Validation Method | Notes |
|--------|-------------------|-------|
| DuckDB | Local smoke & query tests | Used for ad-hoc analytics / developer workflows |
| Trino  | Optional external smoke test (env `TRINO_JDBC_URL`) | Query validation executed when env present |
| Athena | Optional external smoke test (env `ATHENA_CATALOG(see example below)ATHENA_DB`) | Skipped if vars absent |
| Spark  | Community‑driven (schema stable) | No bespoke code; rely on standard Parquet reader |

`internal/core/focus/writers/parquet_writer_external_smoke_test.go` conditionally writes a tiny file then external CI jobs can run engine-specific SQL.

## Schema Stability
Schema defined in `internal/core/focus/types/focus_v1_2.go`. Backwards compatibility rules:
1. Append-only for optional (nullable) fields is non-breaking
2. Modification of required field type/name requires major version bump (FOCUS spec change)
3. Field removal prohibited until next major spec revision

## Compression & Encoding
- Default codec: `snappy`
- Rotation disabled with `RotateSizeBytes=-1` in smoke tests to produce a single file for external validation.

## Timestamp Handling
All temporal fields stored as `TIMESTAMP_MILLIS` (UTC). Engines normalizing to micro/nano may up-convert; CostScope maintains millisecond precision source fidelity.

## Case & Naming
Snake_case column names align with FOCUS; engines with case folding (Athena/Trino) maintain readability; avoid quoting in generated SQL to retain portability.

## Partitioning Strategy (Future)
Current writer emits flat files. Planned enhancements:
- Optional directory partitioning by `provider_name=.../billing_period_start=YYYY-MM/`
- Manifest generation for Athena/Trino glue catalogs

## External Engine Query Examples
Trino (ad hoc):
```sql
SELECT service_name, SUM(effective_cost) total_cost
FROM focus_cost_data
WHERE charge_period_start >= DATE '2025-08-01'
GROUP BY service_name
ORDER BY total_cost DESC
LIMIT 10;
```

DuckDB (local):
```sql
SELECT DATE_TRUNC('month', charge_period_start) m, SUM(effective_cost) cost
FROM focus_cost_data
GROUP BY m ORDER BY m;
```

## Compatibility Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Schema drift | Central schema generator + tests referencing `GetFocusV12Schema()` |
| Engine timestamp mismatch | Standard TIMESTAMP_MILLIS + test reading back in DuckDB |
| Compression incompatibility | Stick to Snappy (widely supported) |
| Field addition breaking downstream | Only add nullable fields; update release notes |

Keep this file updated when new engine validation steps or partitioning ship.
---
title: Engines Compatibility
description: Parquet output compatibility across DuckDB, Trino, Athena, Spark.
---
## Supported Engines
CostScope emits Parquet (Snappy by default) adhering to the FOCUS v1.2 schema; files are consumable by DuckDB, Trino, Athena, and Spark without transformation.

| Engine | Validation Method | Notes |
|--------|-------------------|-------|
| DuckDB | Local smoke & query tests | Used for ad-hoc analytics / developer workflows |
| Trino  | Optional external smoke test (env `TRINO_JDBC_URL`) | Query validation executed when env present |
| Athena | Optional external smoke test (env `ATHENA_CATALOG(see example below)ATHENA_DB`) | Skipped if vars absent |
| Spark  | Community‑driven (schema stable) | No bespoke code; rely on standard Parquet reader |


```bash
,
```
`internal/core/focus/writers/parquet_writer_external_smoke_test.go` conditionally writes a tiny file then external CI jobs can run engine‑specific SQL.

## Schema Stability
Schema defined in `internal/core/focus/types/focus_v1_2.go`. Backwards compatibility rules:
1. Append-only for optional (nullable) fields is non-breaking
2. Modification of required field type/name requires major version bump (FOCUS spec change)
3. Field removal prohibited until next major spec revision

## Compression & Encoding
- Default codec: `snappy`
- Rotation disabled with `RotateSizeBytes=-1` in smoke tests to produce a single file for external validation.

## Timestamp Handling
All temporal fields stored as `TIMESTAMP_MILLIS` (UTC). Engines normalizing to micro/nano may up-convert; CostScope maintains millisecond precision source fidelity.

## Case & Naming
Snake_case column names align with FOCUS; engines with case folding (Athena/Trino) maintain readability; avoid quoting in generated SQL to retain portability.

## Partitioning Strategy (Future)
Current writer emits flat files. Planned enhancements:
- Optional directory partitioning by `provider_name=.../billing_period_start=YYYY-MM/`
- Manifest generation for Athena/Trino glue catalogs

## External Engine Query Examples
Trino (ad hoc):
```sql
SELECT service_name, SUM(effective_cost) total_cost
FROM focus_cost_data
WHERE charge_period_start >= DATE '2025-08-01'
GROUP BY service_name
ORDER BY total_cost DESC
LIMIT 10;
```

DuckDB (local):
```sql
SELECT DATE_TRUNC('month', charge_period_start) m, SUM(effective_cost) cost
FROM focus_cost_data
GROUP BY m ORDER BY m;
```

## Compatibility Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Schema drift | Central schema generator + tests referencing `GetFocusV12Schema()` |
| Engine timestamp mismatch | Standard TIMESTAMP_MILLIS + test reading back in DuckDB |
| Compression incompatibility | Stick to Snappy (widely supported) |
| Field addition breaking downstream | Only add nullable fields; update release notes |

Keep this file updated when new engine validation steps or partitioning ship.
