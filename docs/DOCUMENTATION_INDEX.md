# Documentation index

This documentation index is maintained in-tree. Expand as needed.

This document consolidates the detailed content formerly in the root `README.md` to keep the top‑level file concise. It is the authoritative extended reference for advanced features, configuration, workflows, and operational guidance.

> For quick install + benefits, see the root `README.md`. For architecture see `architecture/overview.md`; for conversion specifics see `architecture/focus-conversion.md`.

---

## Table of Contents

1. FOCUS Specification & Core Features
2. Multi-Cloud Provider Support
3. Advanced Analytics & Experimental Features
4. Reporting & Metadata Persistence
5. Enterprise Feature Stub Generation
6. Real-time & Streaming Processing
7. Security (Auth, RBAC, Audit Mode)
8. Unified API Response Envelope
9. Aggregated Diff Insights (Comparison + Forecast)
10. Technical Capabilities Overview
11. Container Smoke Tests (CI)
12. Configuration Precedence & Resolution
13. Performance Benchmarks & Parity Guards
14. Data Parity + Invariants CI Guard
15. Testing & Quality Docs (Strategy, UAT, Config, Invariants)
16. Supply Chain & Verification (SBOM & Signatures)
17. Release Policy (SemVer & Process)
18. Contributing & Development Workflow
19. Compatibility & Support Matrix
20. Build Variants & Tags (duckdb, sqlite, enterprise, experimental)
21. Module Overview (Commands Summary)
22. SLO & Runbooks (see `runbooks/slo.md`)
23. Architecture Summary (Quick Inline View)
24. C4 Diagram Set (Context/Container/Component)
25. Glossary & Threat Model
26. Diagrams (source + generation)
27. Secure Defaults (Secrets Guidance)
28. Experimental Flags & Focus Engine
29. Commands Reference (Abbreviated)
30. Advanced Analytics / Extended Output Blocks
31. Reporting Output & Metadata Retention Examples
32. RBAC Audit Mode Rollout Guidance
33. Azure Discount Normalization Details
34. Performance / Bench Strategy & Thresholds
35. Supply Chain Verification Steps
36. Release Automation & Manual Hotfixes
37. Release Checklist & Promotion Pipeline

---

## 1. FOCUS Specification & Core Features

Full FOCUS v1.2 implementation, streaming conversion, validation, quality invariants (optional, lightweight) – see `architecture/focus-conversion.md` for deep dive.

## 2. Multi-Cloud Provider Support

AWS CUR / Azure Cost Management / GCP Billing (BigQuery export) → unified FOCUS. Unified opt‑in mapper for parity & migration.

## 3. Advanced Analytics & Experimental Features

The analytics-advanced build (enable via the experimental build tag) provides prototype forecast, anomaly, and optimization commands returning deterministic mock structures for interface validation. When disabled, the code emits metrics noting the feature gate.

Example: build with the experimental tag

```bash
go build -tags experimental ./...
```

## 4. Reporting & Metadata Persistence

Config keys (reports.output_dir, reports.metadata_store_path) control retention and output. Backends: file JSONL (default) or SQLite (when a DB path is provided). If an output flag is omitted, the tool will auto-generate a filename.

## 5. Enterprise Feature Stub Generation

Use the provided make helper to scaffold an enterprise stub. Example:

```bash
make gen-enterprise-stub FEATURE=streaming_engine PKG=internal/core/streaming TYPE=EnterpriseStreamingEngine
```

This creates a stub and an enterprise-tagged implementation scaffold. Stubs emit unified metrics and helpful errors.

## 6. Real-time & Streaming Processing

Streaming jobs include progress tracking, WebSocket updates, and a job manager for concurrency. Lightweight invariants may be streamed alongside conversion when enabled.

## 7. Security (Auth, RBAC, Audit Mode)

Authentication supports JWT and API keys. RBAC metrics include costscope_rbac_checks_total and costscope_rbac_audit_soft_denies_total. The audit rollout pattern is: enable audit mode, observe soft denies, adjust policies, then move to hard enforcement. See SECURITY.md for details.

## 8. Unified API Response Envelope

The standard JSON envelope is: { success, data, error{message,code}, meta{timestamp,request_id} }. Helper functions include response.AutoOK200, AutoCreated201, AutoNoContent204, AutoBadRequest, etc. The guard scripts block unsafe/old raw response patterns.

## 9. Aggregated Diff Insights

Example CLI usage for diffs and forecasts:

```bash
costscope diff baseline.parquet current.parquet --insights --forecast-periods 30
```

This command produces a combined diff, an executive summary, and an optional forecast block.

## 10. Technical Capabilities Overview

Supported formats: Parquet, CSV, JSON (with optional compression). Datastores: DuckDB for analytics and SQLite for metadata. Observability is via Prometheus metrics and OpenTelemetry tracing when configured.

## 11. Container Smoke Tests (CI)

CI workflows build standard and distroless images, validate /health/live and /metrics (expecting costscope\_ metrics), and run TLS variant checks.

## 12. Configuration Precedence & Resolution

Deterministic precedence is: explicit (flag/API) > YAML > ENV > default. The resolved field name is recorded as config_precedence_resolved. Helper functions used in the codebase include ResolveString, Bool, Int, Float, and Duration. Sensitive values are masked by key patterns when rendered in logs or traces.

## 13. Performance Benchmarks & Parity Guards

Run the perf benchmark target (make perf-bench) against a synthetic dataset to validate performance ratios. The CI guard expects unified vs fast path thresholds (for example, duration ratio < 1.15 and allocation ratio < 1.20). Baseline JSON is stored in the repository and optional environment variables may override thresholds during experimentation.

## 14. Data Parity + Invariants CI Guard

The data parity guard ensures unified mapper parity (aggregates + lightweight hash) and checks that invariants remain within configured tolerances. Distinct exit codes indicate parity vs invariants failures. Relevant artifacts are uploaded for human inspection when failures occur.

## 15. Testing & Quality Docs (Strategy, UAT, Config, Invariants)

See tests/strategy.md, tests/uat.md, tests/config-tests.md, and tests/data-quality-invariants.md for testing strategy and quality guidelines.

## 16. Supply Chain & Verification

Release artifacts include the binary, sbom.json, checksums.txt and checksums.txt.sig (signed). Verifying signatures and checksums is a required step during release validation. Optional SBOM analysis is performed with Syft; see section 31 for example commands.

## 17. Release Policy (SemVer & Process)

Conventional commits subset; changelog managed via `CHANGELOG.md`. Pre‑1.0 MINOR may break; ≥1.0 only MAJOR may break public APIs / CLI / FOCUS schema semantics. Hotfix pre‑release identifiers `-hotfix.N` allowed for security emergencies.

See also: `release/checklist.md` (gated pre-release validation) and `release/promotion.md` (multi-stage pipeline).

## 18. Contributing & Development Workflow

Run `make quality` before PR. Add tests, update docs when surface changes. API contract guard via OpenAPI baseline diff. Generated command builders & spec diff guard.

## 19. Compatibility & Support Matrix

Platforms: Linux (amd64/arm64), macOS (amd64/arm64), Windows via WSL. Go ≥1.21. Build flavors: slim (default), duckdb (`-tags duckdb`), sqlite (`-tags sqlite`), enterprise, experimental.

## 20. Build Variants & Tags

- `duckdb` – embedded DuckDB analytics
- `sqlite` – SQLite metadata store
- `enterprise` – enterprise features (streaming engine, pooling stubs replaced)
- `experimental` – advanced analytics prototype
  Tags are additive; combine as needed (e.g., go build -tags "duckdb sqlite enterprise" ./...).

## 21. Module Overview (Commands Summary)

Primary command groups: `focus`, `analytics`, `analytics-complex`, `reports`, `providers`, `streaming`, `production` / `prod-readiness`, `integration`, `config`, `api`.

## 22. SLO & Runbooks

Availability 99.5%, API p95 ≤800ms, conversion ACK p95 ≤5s, error ratio <1%. SLOs relocated to `runbooks/slo.md`. Runbooks in `docs/runbooks/` (latency, conversion degradation, auth failures).

## 36. Glossary & Threat Model

Short definitions and threat model are available in `docs/glossary.md` and `docs/security/threat-model.md`.

## 37. Diagrams (source + generation)

Diagrams are stored in `docs/assets/` as generated PNG/SVG artifacts. Prefer generating from source (Mermaid/PlantUML) in `docs/diagrams-src/` and committing the generated PNG only; if editing, add both source and generated outputs and document the generation command in `docs/DIAGRAMS.md`.

## 23. Architecture Summary (Quick Inline View)

See `architecture/overview.md` (contains high-level + C4 diagrams). Core layers: API → Core Modules → Core Services → Data Layer → Providers. (Root duplicate `framework_architecture.md` pending review for merge/removal.)

## 24. C4 Diagram Set (Context/Container/Component)

Consolidated into `architecture/overview.md` under section "C4 Diagrams" (system context, container, FOCUS conversion component). Former architecture/c4/ removed.

## 25. Secure Defaults (Secrets Guidance)

No default secrets. Use 32–48+ random bytes (Base64) for JWT, webhook, API key material. Rotate regularly; store in secret manager; audit access.

## 26. Experimental Flags & Focus Engine

`--use-focus-engine` adds extended analytics JSON block (`extended`) with anomalies, forecasts, trends, optimizations, executive summary, timing stats. Forecast horizon & phase timeout configurable via flags / YAML / ENV.

## 27. Commands Reference (Abbreviated)

Examples: see `CLI_HELP.md` (stub). Key flows: convert → validate → analyze → diff → reports generate → api enterprise.

## 28. Advanced Analytics / Extended Output Blocks

Extended block omitted unless focus engine or insights flags used. Contains arrays for phases + per-phase duration to aid latency attribution.

## 29. Reporting Output & Metadata Retention Examples

YAML:

```yaml
reports:
  output_dir: /opt/costscope/reports
  metadata_store_path: sqlite:///opt/costscope/reports_meta.db
  metadata_retention_max_records: 5000
  metadata_retention_max_age: 720h
```

Retention enforced opportunistically after each save; fallback to file JSONL if SQLite init fails.

## 30. RBAC Audit Mode Rollout Guidance

1. Enable audit: `COSTSCOPE_RBAC_AUDIT_MODE=1` → soft denies produce header `X-RBAC-Audit: deny` & metric increment.
2. Observe `costscope_rbac_audit_soft_denies_total` for unexpected access patterns.
3. Adjust roles/policies until soft denies near zero.
4. Disable audit variable → hard enforcement (403) & denied counts shift to `costscope_rbac_checks_total{allowed="false"}`.

## 31. Azure Discount Normalization Details

Substring (`discount`) in ChargeType/BillingType ⇒ ChargeCategory=Discount (even negative). Negative without token ⇒ Credit. Metrics: normalization total + skips (when env disables). Diagnostic disable: `COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION=1`.

## 32. Performance / Bench Strategy & Thresholds

Unif vs fast: duration ratio ≤1.15, alloc ratio ≤1.20 (defaults). Bench tool emits JSON + optional perf_engine micro-block if enabled. Failing ratios surface in CI.

## 33. Supply Chain Verification Steps

```bash
VER=v0.x.y
curl -LO https://github.com/costscope/costscope/releases/download/$VER/checksums.txt
curl -LO https://github.com/costscope/costscope/releases/download/$VER/checksums.txt.sig
COSIGN_VERSION=v2.2.4
curl -sSfL https://github.com/sigstore/cosign/releases/download/$COSIGN_VERSION/cosign-linux-amd64 -o cosign && chmod +x cosign
COSIGN_EXPERIMENTAL=1 ./cosign verify-blob --signature checksums.txt.sig checksums.txt
# Then verify binary sha256 line matches
sha256sum -c <(grep "costscope-$VER-linux-amd64" checksums.txt)
```

SBOM (`sbom.json`) – inspect with `jq` or ingest into security tooling.

## 34. Release Automation & Manual Hotfixes

Standard release: tag push triggers workflow (build → sign → SBOM → publish draft release). Manual hotfix: create `vX.Y.Z-hotfix.N` tag; pipeline produces artifacts; merge fix back to main. Breaking changes require major bump & updated OpenAPI baseline.

## 35. Release Checklist & Promotion Pipeline

Centralized release execution:

- `release/checklist.md` – human validation gates (security, integrity, performance, docs, tagging).
- `release/promotion.md` – automated stages (build → sign → sbom → smoke → stage → promote → tag) with extensibility hooks.

Adopt workflow: complete checklist (all boxes) → run promotion target (`make release-promo RELEASE_VERSION=X.Y.Z`) → verify SBOM & signatures → publish release notes.

---

Generated: consolidated from previous root README (date: 2025-08-26). Keep this file updated when adding/removing advanced features.
