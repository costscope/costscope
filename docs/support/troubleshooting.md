---
title: Troubleshooting Guide
description: Canonical troubleshooting reference for conversion, auth, performance, and diagnostics.
---

# Troubleshooting Guide

This is the canonical troubleshooting document (all previous duplicates consolidated).

## Conversion Issues

### 1. Error: "Missing required FOCUS field ..." / validation failures
**Cause:** Source billing record missing or mis-mapped field.
**Fix:** Re-run with `--verbose` to log mapping. Ensure using latest version and correct provider flag. If AWS CUR manifest, pass the manifest not individual partition CSVs.

### 2. Error: "unsupported timestamp format" during GCP/Azure conversion
**Cause:** Non-standard timestamp variant. Parser is tolerant but may fail on custom exports.
**Fix:** Verify raw file; ensure no Excel re-save altered formatting. Open issue attaching a redacted sample row.

### 3. Memory spikes (OOM kill) converting large file without streaming
**Cause:** Non-streaming path loads large chunks.
**Fix:** Use `--streaming--chunk-size` (e.g. 25000) and workers (4–8). Monitor RSS; keep within available memory.

```bash
plus tune
```

### 4. Very slow conversion throughput
**Cause:** Network filesystem I/O or insufficient concurrency.
**Fix:** Copy input locally first. Increase `--workers--parquet-compression zstdsnappy` faster.

```bash
. Use
```

```bash
only if CPU headroom exists; otherwise
```

### 5. Parquet rotation produced multiple files unexpectedly
**Cause:** Default rotation size triggered (512MB) or interval set.
**Fix:** Disable rotation with  for a single file deterministic output.

```bash
--rotate-size -1
```

## Authentication & Authorization

### 6. 401 Unauthorized on protected endpoint
**Checklist:**
- JWT secret exported ((see example below) ≥ 32)

```bash
echo $COSTSCOPE_JWT_SECRET | wc -c
```
- Token not expired; system clock correct (NTP)
- Header formatted:

```bash
Authorization: Bearer <token>
```
- Role has permission (check RBAC config if Casbin enabled)

### 7. 403 Forbidden though token valid
**Cause:** Role/permission mismatch.
**Fix:** Inspect RBAC policies (`configs/rbac_policy.csv.example`). Enable debug logging temporarily if compiled with it. Ensure tenant header present when multi-tenancy flag enabled.

### 8. API key auth failing
**Cause:** Key not registered or header name mismatch.
**Fix:** Confirm creation endpoint response stored correctly; header must match documented format (e.g. `X-API-Key`). Verify no leading/trailing whitespace.

## Performance & Resource Usage

### 9. High latency spikes in API requests (> p95 SLO)
**Cause:** Concurrent heavy conversions saturating CPU.
**Fix:** Stagger jobs, reduce worker count per job, or scale horizontally. Check metrics: CPU, `costscope_conversion_active_jobs` gauge.

### 10. Increased converter duration after enabling unified mapper
**Cause:** Unified mapper slower by design (parity mode).
**Fix:** Disable unified mapper (remove flag/env) unless explicitly validating. Investigate if ratio > 1.20 vs baseline.

```bash
make perf-parity
```

### 11. Excessive disk usage from rotated Parquet
**Cause:** Rotation size too small or old files uncollected.
**Fix:** Increase `--rotate-size` or disable. Implement lifecycle policies (S3/GCS) or periodic cleanup script.

## General Debug Workflow
1. Re-run with `--verbose` (or increase log level via config).
2. Capture error log snippet + command invocation.
3. Check `/metrics` for correlated counters / spikes.
4. Run  on produced Parquet to isolate data issues.

```bash
costscope validate
```
5. If reproducible, open issue referencing version ().

```bash
costscope version
```

## Health & Diagnostics Commands
-- Basic health:

```bash
curl -s localhost:8080/health
```

-- Metrics sample:

```bash
curl -s localhost:8080/metrics | head -100
```

-- Version:

```bash
costscope version
```

-- Perf bench:

```bash
make perf-parity
```

## When to Open an Issue
Open an issue if:
- Data mapping deviates from FOCUS spec
- Unified vs alternate mapper aggregates differ
- Conversion crash (panic) with valid input
- Contract guard blocks expected non-breaking spec addition

Provide: version, command, minimal sample rows (sanitized), and logs.

---
For environment-specific VS Code or language server issues see `vscode-troubleshooting.md`.

## See also
- `support.md`
- `faq.md`
- `vscode-troubleshooting.md`

> Missing scenario? Open an issue with failing command and minimal anonymized sample.