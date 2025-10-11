# Release Scope Validation – v0.1.0 (Internal Draft)

Status: Draft (not yet merged into CHANGELOG)
Date: 2025-09-10

## Summary
This document captures validated feature scope, gaps vs published CHANGELOG 0.1.0, and classification of TODO items prior to finalizing the public changelog amendment.

## Feature Status Matrix

| Feature | Status | Notes / Action |
|---------|--------|----------------|
| CLI convert | ok | Implemented (`focus convert` + streaming) |
| CLI validate | gap | Present in code; missing from CHANGELOG |
| Invariants & drift (convert/validate flags) | gap | Implemented flags/tests; add to CHANGELOG |
| Multi-cloud converters (AWS/Azure/GCP) | ok | Implemented |
| Streaming pipeline | ok | Worker pool + memory controls present |
| Enterprise API (JWT/API key/RBAC/rate limit) | clarify | JWT validation stub only; either harden or soften wording |
| OpenAPI docs loading | gap | Handler returns stub; real file load & Swagger UI pending |
| Swagger UI | gap | Not implemented; mention as future work |
| Unified performance engine | clarify | Core parallel/caching present; optimizer/ML TODO |
| Metrics & tracing | ok | Prometheus + OpenTelemetry integrated |
| Reports export | ok | Basic export present |
| Comparison advanced export/scheduling | defer | TODO placeholders; not GA |
| Security tooling (scans, gate) | gap | Exists (Make targets, scripts); not listed |
| Helm chart | gap | Chart exists under `charts/costscope/`; add mention |
| SBOM generation | ok | `sbom` target (Syft) |
| Checksums/signing | ok | `checksums`, `sign-checksums`, cosign keyless |
| Release promotion pipeline | clarify | Present (`release-promo`); only partially implied in CHANGELOG |

## Blocking vs Defer TODO Classification
(Top blocking-first; others grouped for backlog issue creation.)

### Proposed Blocking (address OR adjust wording before final changelog amendment)
1. JWT validation stub (`internal/api/middleware/enterprise.go`) – implement minimal signature/issuer check OR change wording to “initial enterprise API scaffolding”.
2. OpenAPI spec stub (`internal/api/handlers/docs.go`) – load real `api/openapi.v1.json` file OR adjust wording to “initial documentation endpoint stub”.
3. Missing CHANGELOG entries (validate, invariants, helm chart, security tooling) – add section or bullet list.

### Defer (create issues, not blocking release if transparently documented)
- Swagger UI serving (`docs.go`).
- Analysis comparison/export/scheduling TODO trio (group into one issue).
- Comprehensive FOCUS validation (`focus_v1_2.go`).
- Parquet/CSV/HTML export placeholders in comparison engine.
- Plugin scanning (`internal/framework/plugin.go`).
- Optimizer / ML analytics wiring & circular import resolution (duckdb engine + enhanced service param use TODOs).
- WebSocket coordination TODO.

## Recommended Actions Before Finalizing
- Decide: implement quick JWT validation (use HS256 + env secret / placeholder) OR revise wording. (If implement, re-run security scan.)
- Implement simple OpenAPI file loader (optional) otherwise soften wording.
- Amend CHANGELOG once decisions above are locked.
- Create umbrella issues:
  * "Security Hardening 0.1.x" (JWT + docs load + Swagger UI).
  * "Analysis Enhancements".
  * "FOCUS Validation Deep Checks".
  * "Optimizer & ML Roadmap".
  * "Export Format Extensions".

## Draft CHANGELOG Amendments (Proposed)
Add bullets under 0.1.0 “Added” (after approval):
- FOCUS validation command (schema, quality, performance, anomalies, batch mode).
- Invariants & drift detection (baseline comparison, tolerance, fail-on-drift, JSON report).
- Helm chart for Kubernetes deployment.
- Security & supply-chain tooling (static/vuln/secret/container scans, SBOM, signing pipeline).
- Release promotion pipeline (secure build → sign → SBOM → smoke → stage/promote).

If wording softened:
- Replace “Enterprise API (JWT/API key auth, RBAC, rate limiting) with OpenAPI docs.” with “Initial Enterprise API scaffolding (API key auth, RBAC, rate limiting, placeholder JWT check, stub OpenAPI endpoint).”

## Verification Notes
- All features marked ok located via code grep and Make targets.
- No `todo.md` found; this document supersedes missing plan artifact for audit.

## Next Steps
1. Team review & choose implement vs soften for JWT/OpenAPI.
2. Create issues for defer items.
3. Single controlled CHANGELOG update referencing this scope validation.

(End of internal scope draft)
