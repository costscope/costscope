---
title: Architecture - Data Layer
description: Analytical storage (DuckDB), metadata state (SQLite), file artifacts and caching.
---

# Data Layer

## Stores
| Layer | Technology | Purpose |
|-------|------------|---------|
| Analytics | DuckDB | Columnar analytical queries & aggregations |
| Metadata | SQLite | Jobs, configuration snapshots, sessions |
| File Artifacts | Parquet/CSV | Exchange & durable storage |
| Cache | In-memory | Hot path normalization & lookups |

## Data Flow (Simplified)
`Input (CUR/Exports) -> Normalization -> FOCUS Parquet -> Query/Analytics (DuckDB) -> Reports / APIs`

## See also
- `../dev/performance-benchmarks.md`
- `api-layer.md`
- `core-services.md`
