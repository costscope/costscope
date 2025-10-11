---
title: RBAC Tracing Schema
description: OpenTelemetry span attributes for RBAC permission checks.
---
## Span Overview
Every permission evaluation creates an internal span named `rbac.has_permission` (kind=INTERNAL) when tracing is enabled (OTEL exporter initialized). This provides latency breakdown and decision context for authorization paths.

## Attributes
| Attribute | Example | Description |
|-----------|---------|-------------|
| `rbac.role(see example below)analyst` | Role evaluated (may differ from authenticated principal if role mapping applied) |
| `rbac.resource(see example below)focus` | Target logical resource constant |
| `rbac.action(see example below)convert` | Requested action constant |
| `rbac.allowed(see example below)true` | Boolean decision result |


```bash
|
```
## Metrics Linkage
Span duration feeds histogram `costscope_rbac_check_latency_seconds(see example below)RBACCheckLatency(see example below)costscope_rbac_checks_total{allowed="allowed|denied"}(see example below)costscope_rbac_audit_soft_denies_total`.


```bash
(exposed via
```

```bash
). Decision outcome increments
```

```bash
and in audit mode (soft deny) also increments
```
## Audit Mode Flow
1. Middleware calls `CheckPermission` with context including active trace.
2. Span started → attributes set (role/resource/action).
3. Decision made; attribute `rbac.allowed` updated.
4. If denied and env `COSTSCOPE_RBAC_AUDIT_MODE=1X-RBAC-Audit: denyrbac_audit_soft_denies_total` incremented.

```bash
set → request allowed, header
```

```bash
returned, metric
```
5. Span ends (decision latency recorded).

## Sampling Guidance
Authorization spans are cheap; recommended to sample at ≥10% in production to assist debugging policy mismatches. For high-throughput conversion clusters, apply tail-sampling based on latency > P95 or denied outcome to focus storage.

## Troubleshooting
| Symptom | Investigation |
|---------|---------------|
| High RBAC latency | Inspect spans filtered by operation name; identify spikes >50ms (likely I/O or lock) |
| Unexpected denies | Query spans where `rbac.allowed=false` and correlate role to policy definitions |
| Missing spans | Ensure OTEL exporter initialized before API server start |

Keep updated if span name or attribute keys change.
