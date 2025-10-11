---
title: Logging & Metrics
description: Standard logging policy, metrics catalog, exemplar queries, and observability integration guidance.
---

# Logging & Metrics Standardization

This document supersedes the previous combined logging & metrics draft. The earlier combined draft was removed after consolidation.

## Overview
* Structured JSON lines (UTC) with correlation IDs.
* Strict PII redaction patterns (password|secret|token|api_key|authorization|cookie|email|ssn).
* Prometheus metrics exposition at `/metrics/debug/cache-stats` when enabled.

```bash
and optional debug endpoint
```
* Consistent metric naming: `costscope_<domain>_<metric>` using buckets for latency histograms.

## Logging Policy
- Format: JSON line per event, fields:  + domain keys.

```bash
ts, level, service, msg, request_id, trace_id, span_id
```
- Levels: debug, info, warn, error, fatal (panic reserved internally).
- Truncation: entries above `COSTSCOPE_LOG_MAX_BYTES` truncated with suffix.
- Redaction: matching sensitive key substrings replaced with `***`.
- Correlation: middleware injects IDs; propagate W3C Trace Context headers.

Example:
```
{"ts":"2025-08-12T10:00:00Z","level":"info","service":"costscope","msg":"conversion completed","request_id":"a1b2","trace_id":"6f1b...","span_id":"8a2c...","provider":"aws","records":123456}
```

## HTTP Middleware Summary
| Middleware | Purpose |
|------------|---------|
| RequestID | Ensures `X-Request-ID` + context key |
| Tracing | Inject / extract trace + span ids |
| Prometheus | Instrument request count, latency histograms |

## Metrics Catalog (Selected)
| Metric | Type | Key Labels | Notes |
|--------|------|-----------|-------|
| costscope_converter_records_total | Counter | provider,status | Converted records count |
| costscope_converter_duration_seconds_bucket | Histogram | provider | Conversion latency distribution |
| costscope_unified_mapper_rows_total | Counter | provider,path | Unified vs fast path volume |
| costscope_http_requests_total | Counter | method,path,status | API traffic |
| costscope_http_request_duration_seconds_bucket | Histogram | method,path,status | Request latency |
| costscope_normalization_cache_hits_total | Counter | type,provider | Enum / region normalization cache hits |
| costscope_classifier_decisions_total | Counter | provider,path,decision | Charge classification attribution |
| costscope_enterprise_feature_invocations_total | Counter | feature,allowed | Enterprise gate usage |
| costscope_enterprise_feature_errors_total | Counter | feature,error_kind | Error modes for gated features |
| costscope_health_readiness | Gauge | (none) | Readiness state (1/0) |

### Normalization Cache Metrics
Expose efficiency & size. Enable periodic refresh with .

```bash
--enable-cache-stats --cache-metrics-refresh-interval 30s
```

### Example Prometheus Queries
RPS:
```
sum(rate(costscope_http_requests_total[5m]))
```
P95 latency by route:
```
histogram_quantile(0.95, sum(rate(costscope_http_request_duration_seconds_bucket[5m])) by (le,path))
```
Converter throughput:
```
sum(rate(costscope_converter_records_total{status="ok"}[5m])) by (provider)
```
Classifier decisions distribution:
```
sum(rate(costscope_classifier_decisions_total[5m])) by (provider,decision)
```

### Debug Endpoint: /debug/cache-stats
Authenticated (admin role) JSON snapshot of normalization caches. Register only when explicitly enabled.

## Tracing & OTLP
Set `OTEL_EXPORTER_OTLP_ENDPOINTOTEL_SERVICE_NAME` to override default service label.

```bash
to export traces; include correlation IDs in spans. Optional
```

## Alerting Examples
Azure discount normalization spike:
```
sum(rate(costscope_azure_discount_normalizations_total{provider="azure"}[5m])) >
 (sum(rate(costscope_azure_discount_normalizations_total{provider="azure"}[1h])) * 2.5)
```
Provider registry fallback (should be zero):
```
increase(costscope_providers_registry_fallback_total[10m]) > 0
```

## Observability Integration Checklist
- [ ] `/metrics` scraped (Prometheus or OTEL collector).
- [ ] Log retention & centralization configured.
- [ ] Trace sampling policy defined (tail or head).
- [ ] Key SLO panels: latency p95, error rate, conversion throughput, cache hit ratio.
- [ ] Alerts tuned (noisy metrics threshold stabilized).

## See also
- `monitoring/monitoring-overview.md`
- `monitoring/dashboards.md`
- `../security/`
- `../architecture/overview.md`
- `production-deployment.md`

> This is the canonical logging & metrics reference.
# Logging & Metrics Standardization

JSON lines logs, UTC timestamps, PII redaction, request_id/trace_id correlation, Prometheus metrics catalogue and example queries. Expose /metrics and optional /debug/cache-stats.
