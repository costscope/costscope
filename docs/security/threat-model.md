---
title: Threat Model
description: High-level STRIDE threat model and mitigations for core CostScope components.
last_reviewed: 2025-09-08
---

# Threat Model

This document outlines primary threats (STRIDE) across major surfaces (API, conversion pipeline, JobStore, supply chain) and maps them to current & planned mitigations. It is a lightweight, living artifact – revisit after material architecture or dependency changes.

## Scope

In-scope components:
* API layer (authentication, RBAC, request handling)
* Conversion pipeline (ingestion, normalization, Parquet output)
* JobStore (async job metadata persistence)
* CLI execution (local usage, scripting)
* Build & release pipeline (artifacts, SBOM, signatures)

Out-of-scope (covered elsewhere / future iteration): multi-tenant isolation specifics, detailed provider credential hardening guidelines.

## STRIDE Matrix

| Component | Spoofing | Tampering | Repudiation | Information Disclosure | Denial of Service | Elevation of Privilege | Key Mitigations |
|-----------|----------|----------|-------------|------------------------|-------------------|------------------------|-----------------|
| API Auth / RBAC | Stolen or forged JWT | N/A (token content) | Lack of audit trail | Token leakage | Excessive auth attempts | Weak role scoping | HS256/RS256 JWT, secret rotation guidance, audit log roadmap, rate limiting, minimal claims |
| Conversion Pipeline | Fake provider data source | Data row manipulation | Limited trace on batch failure | Sensitive cost usage leakage | Large file floods | Abuse of mapper flags | Input validation, invariants drift checks, streaming memory caps, file size limits (planned), observability spans |
| JobStore (future Bolt) | Unauthorized job injection | Mutation of job status/history | Missing write audit | Job metadata exposure | Long-running job starvation | Bypass retention policy | RBAC gating endpoints, separation of submit vs status verbs, future immutable event log |
| Build / Release | Malicious dependency | Binary tampering | Unattributed build | Build log leakage | Resource exhaustion (CI) | Gain signing capability | Go module sums, checksums, SBOM, reproducible builds, planned cosign attestation/policy |
| CLI Local Use | Alias of binary | Config tampering | Lack of local audit | Exposure of config with secrets | Resource hog via large inputs | Elevated config privileges | Checksum verification, principle of least privilege env vars, planned config scanner |

## Additional Mitigations & Backlog

| Area | Current | Planned / Backlog |
|------|---------|-------------------|
| Secrets Management | Env var secret loading | Encrypted secret store integration guidance |
| Policy Enforcement | Basic RBAC checks | Casbin integration, fine-grained policies |
| Supply Chain | SBOM, checksums, reproducible build doc | Cosign attestations, dependency diff gate |
| Observability | Metrics + partial tracing | Full trace coverage of conversion sub-phases |
| Data Validation | Invariants + parity | Extended schema evolution checks |

## Review Cadence

Review at least quarterly or when: significant dependency added, new data store introduced, or auth model changes.

## Change Log

Record substantive threat model updates here with date + summary.
