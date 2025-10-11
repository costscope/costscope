---
title: Monitoring Architecture
status: draft-split
description: Observability stack, health endpoints, and telemetry architecture for CostScope.
---

# Monitoring & Observability Architecture

Split from `overview.md` for focused evolution and deeper dashboards / alerting content.

## Observability Stack
```
┌──────────────────────────────────────────────────────────┐
│                   Observability                          │
├──────────────────────────────────────────────────────────┤
│ [METRICS] Prometheus + custom business metrics           │
│ [LOGGING] Structured JSON (ingest to ELK / Loki / Splunk)│
│ [TRACING] OpenTelemetry exporters                        │
│ [ALERTING] Rules => PagerDuty / Teams / Email            │
└──────────────────────────────────────────────────────────┘
```

### Metrics Categories
- System: CPU, memory, goroutines, GC pauses.
- Ingestion: rows processed/sec, provider latency, error counts.
- Jobs: queue depth, active workers, retry rate, duration percentiles.
- API: request rate, p95 latency, error ratio, auth failures.
- Storage: DuckDB query duration, file IO wait, cache hit ratio.

### Logging Strategy
| Layer          | Format  | Rationale                         |
|----------------|---------|-----------------------------------|
| API            | JSON    | Structured ingestion & filtering  |
| Jobs/Workers   | JSON    | Correlation via job_id            |
| Security/Audit | JSON    | Immutable compliance trail        |
| Startup        | Key=Val | Human quick diagnostics           |

### Tracing Scope
- Request spans wrap handler execution.
- Child spans for DB queries, provider calls, stream batches.
- Correlate job execution spans with WebSocket progress events.

### Health Endpoints
| Endpoint          | Purpose            | Typical Probe |
|-------------------|--------------------|---------------|
| /health           | Basic online       | 30s           |
| /health/live      | Liveness           | 15s           |
| /health/ready     | Readiness gating   | 10s           |
| /metrics          | Metrics scrape     | 15s           |

### Alerting Guidelines
| Domain     | Condition                             | Initial Threshold |
|------------|----------------------------------------|-------------------|
| API        | 5xx error ratio                        | >2% 5m            |
| Ingestion  | Failed rows per batch                  | >0.5%             |
| Jobs       | Retry rate                             | >10% 10m          |
| Storage    | Query p95 latency                      | >2s               |
| Security   | Auth failures                          | >25 / 5m          |
| System     | Memory usage                           | >85%              |

### Telemetry Export
| Channel       | Use Case               | Tooling              |
|---------------|------------------------|----------------------|
| Prometheus    | Metrics scraping       | Native scrape        |
| OpenTelemetry | Traces + future logs   | OTLP exporters       |
| WebSocket     | Live job progress      | Internal broadcast   |
| File/S3       | Batch report artifacts | Async jobs           |

## Incremental Adoption
1. Start with metrics + liveness.
2. Add structured logging & correlation IDs.
3. Introduce tracing for top N slow endpoints.
4. Add alert rules after baseline established.

## See also
- `../ops/monitoring/monitoring-overview.md`
- `../ops/logging.md`
- `deployment.md`
- `../dev/performance-benchmarks.md`

---
_Status: initial extracted split. Extend with dashboards & exemplar queries later._
