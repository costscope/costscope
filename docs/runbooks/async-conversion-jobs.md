---
title: Asynchronous Conversion Jobs
description: Lifecycle, API/CLI surface, metrics, alert rules, and Grafana panels for async FOCUS conversion jobs.
last_reviewed: 2025-09-08
---

# Asynchronous Conversion Jobs

This document describes the asynchronous FOCUS conversion job lifecycle: submission, observation, cancellation, and history retrieval. It also defines emitted Prometheus metrics and recommends alert rules & Grafana panels.

## Overview

Long‑running conversions (large CUR / usage exports) can be executed asynchronously. A job transitions through:

(see example below)


```bash
pending → running → (success | failed | cancelled)
```
Each job captures: id, provider, input_path, output_path, submitted_at, started_at, finished_at, status, error (if any), and basic counters (records_processed if available).

## API Endpoints (JSON envelope)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/focus/convert/async` | Submit a new conversion job |
| GET | `/api/v1/focus/jobs` | List active + recent jobs |
| GET | `/api/v1/focus/jobs/{id}` | Get job status & progress |
| DELETE | `/api/v1/focus/jobs/{id}` | Cancel a pending/running job (idempotent) |
| GET | `/api/v1/focus/jobs/history?limit=N` | Completed/failed/cancelled history (persisted) |

Example submit request:

```json
POST /api/v1/focus/convert/async
{
	"provider": "aws",
	"input_path": "/data/aws-cur.csv.gz",
	"output_path": "costscope-data/focus/aws-focus.parquet",
	"options": {"streaming": true, "invariants": true}
}
```

Response:

```json
{
	"success": true,
	"data": {"job_id": "job_1756031648188023000", "status": "pending"},
	"meta": {"timestamp": "2025-08-24T10:34:08Z"}
}
```

Status response (success example):

```json
{
	"success": true,
	"data": {
		"job_id": "job_1756031648188023000",
		"provider": "aws",
		"status": "success",
		"submitted_at": "2025-08-24T10:34:08Z",
		"started_at": "2025-08-24T10:34:08Z",
		"finished_at": "2025-08-24T10:34:15Z",
		"records_processed": 1045231,
		"output_path": "costscope-data/focus/aws-focus.parquet"
	},
	"meta": {"timestamp": "2025-08-24T10:34:15Z"}
}
```

## CLI Commands

| Command | Description |
|---------|-------------|
| (see example below) | Submit async job |
| (see example below) | List active + recent jobs |
| (see example below) | Show job status |
| (see example below) | Cancel running job |
| (see example below) | Show recent terminal jobs |


```bash
costscope focus convert --provider aws --input cur.csv.gz --output focus.parquet --submit-only
```

```bash
costscope focus jobs list
```

```bash
costscope focus jobs status JOB_ID
```

```bash
costscope focus jobs cancel JOB_ID
```

```bash
costscope focus jobs history --limit 20
```
Submit output includes `job_id=<id>` for scripting.

## Persistence

A pluggable `JobStore` persists terminal job metadata. Current implementation:

* In-memory (default, process‑lifetime only)

Removed (cleanup 2025-08, see CHANGELOG): JSON append-file store (`JSONAppendFileJobStore`). It was a low‑overhead temporary prototype but added maintenance cost and duplicated upcoming durable backends.

Planned (backlog): SQLite / BoltDB (or badger) durable stores for multi-process persistence, plus optional object storage export. These will land behind interfaces already defined in `internal/core/focus/conversion/store` without changing the CLI/API surface.

Migration note: existing deployments relying on the removed append-file path should transition to periodic export of  output if archiving is required until a durable store ships.

```bash
focus jobs history
```

## Prometheus Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `costscope_conversion_jobs_submitted_total` | Counter | – | Jobs submitted |
| `costscope_conversion_active_jobs` | Gauge | – | Currently running jobs |
| `costscope_conversion_jobs_completed_total(see example below)success(see example below)failed(see example below)cancelled`) |
| `costscope_conversion_job_duration_seconds` | Histogram | outcome | Wall time per job (observed when terminal) |


```bash
| Counter | outcome | Terminal jobs by outcome (
```

```bash
,
```
Additional existing converter/mapper metrics remain available for deeper analysis.

### Metric Semantics

* `active_jobs` increments at submission, decrements exactly once on terminal state (even if cancelled mid-run).
* Duration histogram is recorded only for jobs that started (pending but never started due to early cancellation can be excluded depending on implementation; current path records duration for all with a start timestamp).

## Recommended Recording Rules

```yaml
groups:
	- name: conversion_jobs.rules
		interval: 30s
		rules:
			- record: job:conversion:fail_rate_5m
				expr: |
					sum(rate(costscope_conversion_jobs_completed_total{outcome="failed"}[5m]))
						/
					sum(rate(costscope_conversion_jobs_completed_total[5m]))
			- record: job:conversion:cancel_rate_5m
				expr: |
					sum(rate(costscope_conversion_jobs_completed_total{outcome="cancelled"}[5m]))
						/
					sum(rate(costscope_conversion_jobs_completed_total[5m]))
			- record: job:conversion:avg_duration_5m
				expr: |
					sum(rate(costscope_conversion_job_duration_seconds_sum[5m]))
						/
					sum(rate(costscope_conversion_job_duration_seconds_count[5m]))
```

## Alert Rules

```yaml
groups:
	- name: conversion_jobs.alerts
		interval: 1m
		rules:
			- alert: ConversionJobHighFailureRate
				expr: job:conversion:fail_rate_5m > 0.10 and sum(rate(costscope_conversion_jobs_completed_total[5m])) > 0
				for: 10m
				labels:
					severity: warning
				annotations:
					summary: "Conversion job failure rate >10%"
					description: ">10% of conversion jobs failed in the last 5m. Investigate input data or mapper regressions."

			- alert: ConversionJobHighCancelRate
				expr: job:conversion:cancel_rate_5m > 0.25 and sum(rate(costscope_conversion_jobs_completed_total[5m])) > 5
				for: 10m
				labels:
					severity: info
				annotations:
					summary: "High cancellation rate (>25%)"
					description: "Users are cancelling many jobs; may indicate incorrect parameters or throughput issues."

			- alert: ConversionJobStuck
				expr: costscope_conversion_active_jobs > 0 and on() (time() - max by() (timestamp(costscope_conversion_jobs_submitted_total))) > 1800
				for: 5m
				labels:
					severity: critical
				annotations:
					summary: "Potentially stuck conversion job(s)"
					description: "Active jobs present with no new submissions for >30m; investigate hung workers or I/O stalls."

			- alert: ConversionJobSlowDuration
				expr: job:conversion:avg_duration_5m > 600
				for: 15m
				labels:
					severity: warning
				annotations:
					summary: "Average conversion duration >10m"
					description: "Sustained slowdown; consider scaling workers or analyzing I/O."
```

## Grafana Panel Suggestions

1. Bar / Time-series: Success vs Failed vs Cancelled (stacked) using `increase(costscope_conversion_jobs_completed_total[1h])` grouped by outcome.
2. SingleStat / Gauge: Current Active Jobs `costscope_conversion_active_jobs`.
3. Table: Recent jobs (via API → panel JSON with transformation) including status & duration.
4. Heatmap: Duration distribution using histogram buckets.

Example panel query (success + failure rate overlay):

```promql
sum(rate(costscope_conversion_jobs_completed_total{outcome="success"}[5m]))
/
sum(rate(costscope_conversion_jobs_completed_total[5m]))
```

## Operational Runbook (Summary)

1. Spike in failure rate → inspect logs filtered by  and sample input paths; validate schema differences.

```bash
conversion job failed
```
2. High cancel rate → correlate with API latency / queue backlog; check user feedback.
3. Stuck jobs → acquire goroutine dump (if enabled), verify file system / object storage availability.
4. Sustained duration increase → profile unified vs fast mapper path; compare perf-bench historical ratios.

## Backlog / Future Enhancements

* Pagination & status filtering on list/history.
* Structured progress events (percentage / records) exposed through WebSocket channel.
* Persisted error categories for richer failure cause dashboards.
* Durable store (SQLite/Bolt) & retention policy (age or count-based trimming).

---
