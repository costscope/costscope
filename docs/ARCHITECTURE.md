# CostScope Architecture

This document provides a concise, high‑level view of CostScope’s architecture to orient contributors and operators. It complements the deeper docs under `docs/architecture/` and API details under `docs/api/`.

- Tech stack: Go 1.24.x, Cobra CLI, Gin HTTP API, DuckDB + SQLite integration, Casbin RBAC, Prometheus metrics, OpenTelemetry tracing.
- Packaging: Static Linux binaries and container images (multi-arch) built via CI.
- Config: Layered YAML with environment overlays (development/staging/production/docker/optimization).

## Core Components

- CLI (Cobra): entrypoints for analysis, conversion, validation, optimization, and admin tasks.
- API (Gin): HTTP server exposing read‑only analytics and operational endpoints; optional WebSocket streaming.
- Analytics Engine: parsing, normalization, aggregation, and optimization of cost/usage datasets.
- Storage: primarily DuckDB/SQLite files; abstractions keep DB choice localized.
- Security: Casbin RBAC policies, least‑privilege defaults; input validation at boundaries.
- Observability: Prometheus metrics (bounded cardinality), OpenTelemetry tracing; structured logging.

## Data Flow (Overview)

1. Ingest: cost/usage sources are read (files/streams/CLI input).
2. Normalize: data mapped into unified internal models.
3. Analyze: metrics, aggregations, and optimization passes.
4. Serve: results rendered via CLI, files (reports), and/or API endpoints.

## Process Model

- CLI commands execute synchronously with cancellable contexts.
- API server runs independently with background workers for long‑running tasks.
- All long operations accept `context.Context` for cancellation and deadlines.

## Extensibility

- Providers: add new input providers behind small interfaces in `internal/`.
- Policies/RBAC: update Casbin model/policy files under `configs/` with tests.
- Metrics: export via existing Prometheus registry; avoid high‑cardinality labels.

## Diagrams

See `docs/DIAGRAMS.md` and sources under `docs/diagrams-src/`. For framework‑level details, see `docs/framework_architecture.md`.

## References

- API Overview: `docs/api/index.md`
- Security Model: `docs/security/`
- Runbooks: `docs/runbooks/`
- Glossary: `docs/glossary.md`

This page is intentionally brief. Expand sections as the architecture evolves; prefer why‑focused notes near non‑obvious decisions.
