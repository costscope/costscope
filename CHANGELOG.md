# Changelog

All notable changes to this project are captured here. This file documents the initial public release. The project follows Semantic Versioning.

## Unreleased

### Breaking

- OpenAPI: explicitly declared required path parameter `id` for existing by-id endpoints, which is detected as a breaking change by oasdiff:
  - Public: `GET /api/v1/focus/conversions/{id}`
  - Enterprise: `GET /focus/conversions/{id}`, `DELETE /focus/conversions/{id}`
    This aligns the specification with actual behavior. Clients calling these by-id routes must ensure the `id` segment is provided in the path template. No change is required for list endpoints.

### Migration

- This change updates the contract to be explicit about required path parameters. If you are gating CI with API contract checks, set `ALLOW_API_DIFF=1` for the acknowledging PR, and reference this CHANGELOG entry in the PR description.

## 0.1.0 - 2025-09-07

### Added

- `costscope` CLI with core commands (version, diff, convert, analyze) and machine-readable output.
- FOCUS validation command (schema, quality, performance, anomalies; single file & batch modes; machine-readable output).
- Multi-cloud converters for AWS, Azure and GCP (CSV/JSON, optional gz).
- Streaming conversion pipeline for large datasets with a worker pool and memory controls.
- Invariants & drift detection (baseline JSON, tolerance, report generation, fail-on-drift flags across convert / validate).
- Initial Enterprise API scaffolding (API key auth, basic RBAC & rate limiting, permissive JWT placeholder, stub OpenAPI docs endpoint).
- Unified performance engine (parallel execution, memory management, caching) and CI bench harness.
- Metrics and tracing (Prometheus + OpenTelemetry) and deterministic memory-limit tests.
- Helm chart for Kubernetes deployment (deployment, service, monitoring, network, ingress, rule templates).
- Reports export and artifact generation (parquet/JSON outputs, initial comparison foundations).
- Security & supply-chain tooling (static analysis, vulnerability + secret + container scans, SBOM generation, checksum signing/verification).
- Release promotion pipeline (build → security gate → sign → SBOM → smoke → stage/promote tagging → release notes → checksums/signature).

### Notes

This is the initial public release. The `Added` list reflects the current functionality; future releases will include grouped entries for changes and fixes as needed.

<!-- End of changelog -->
