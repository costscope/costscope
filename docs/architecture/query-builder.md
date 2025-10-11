---
title: Query Builder Architecture
description: Core vs extended SQL builder surfaces, build tags, and telemetry.
---
## Overview
The FOCUS Query Builder provides a small, composable API for constructing analytical SQL over the canonical `focus_cost_dataqb_extended` build tag.

```bash
table. The surface is intentionally minimal in the core build to keep binary size and compile time low; advanced relational clauses are opt‑in behind the
```

## Design Goals
- Deterministic SQL emission (stable ordering of clauses)
- Low allocation profile (append-only slices reused per builder instance)
- Separation of core vs extended features via build tags
- Telemetry hooks without coupling to specific DB driver layer

## Core Surface (Always Available)
| Method | Purpose |
|--------|---------|
| `Select` | Add projection columns |
| `From` | Set base table |
| `Where` | Add filtered predicates (string fmt) |
| `GroupBy` | Grouping dimensions |
| `OrderBy` | Sort ordering |
| `Limit` / `Offset` | Pagination |
| Convenience filters (`FilterByProvider(see example below)FilterByDateRange`, etc.) | Encapsulate common predicates |
| `SelectCostMetrics` | Standard cost metric rollup set |


```bash
,
```
Extended operations (`JoinLeftJoinHavingWithCTE-tags qb_extended`.

```bash
) are compiled in only with
```

## Extended Build Tag
To enable:
```bash
go build -tags qb_extended ./...
```
Adds advanced relational constructs for multi-table scenarios and CTE pipelines without burdening the default build.

## Example (Core)
```go
qb := focus.NewFOCUSQueryBuilder().
	Select("service_name", "SUM(effective_cost) AS total_cost").
	From("focus_cost_data").
	Where("charge_period_start >= '%s'", start.Format("2006-01-02")).
	Where("charge_period_end <= '%s'", end.Format("2006-01-02")).
	GroupBy("service_name").
	OrderBy("total_cost", database.SortDesc).
	Limit(20)
sql := qb.(interface{ Build() (string, []interface{}) }).Build()
```

## Example (Extended Join)
```go
//go:build qb_extended
qb := focus.NewFOCUSQueryBuilder().
	WithCTE("top_services", "SELECT service_name, SUM(effective_cost) tc FROM focus_cost_data GROUP BY service_name").
	Select("t.service_name", "t.tc").
	From("top_services t").
	Having("SUM(t.tc) > %d", 1000)
```

## Telemetry
Builder usage contributes to aggregate metrics via the wrapping query executor layer (not directly inside builder) to avoid tight coupling. Areas tracked:
- Query build duration histogram (internal executor)
- Emitted SQL size (bytes) for anomaly detection

## Testing Strategy
- Core: table-driven tests on predicate assembly & time grouping
- Extended: build-tag gated tests validating JOIN/HAVING/CTE injection
- Smoke: duckdb & focus builders unified equivalence for cost metric selectors

## Rationale for Tag Split
Most deployments require only single-table aggregations; advanced operations add code paths and potential complexity. Gating keeps default binary lean and reduces surface for SQL injection mistakes.

## Future Enhancements
- Parameter binding abstraction (avoid fmt formatting) for multi-engine portability
- SQL lint step to detect anti-patterns (e.g., SELECT * with grouping)
- Optional AST export for query plan tooling

Keep aligned with `internal/database/focus/query_builder*.go`.
