---
title: Architecture - Core Services
description: Foundational cross-cutting services: config, logging, scheduling, persistence, streaming.
---

# Core Services

## Service Catalog
| Service | Package | Function |
|---------|---------|----------|
| Configuration | `internal/core/config` | Layered precedence resolution |
| Logging & Audit | `internal/core/logging` | Structured logs & audit trail |
| Job Management | `internal/api/jobs` | Async submission + progress |
| Streaming | `internal/core/streaming` | Real-time conversion / updates |
| Persistence | `internal/core/persistence` | Storage abstraction |
| Reports | `internal/core/reports` | Export & formatting engine |

## Patterns
- Dependency injection via constructor funcs.
- Minimal global state (env capture isolated at startup).

## See also
- `api-layer.md`
- `data-layer.md`
- `../dev/performance-benchmarks.md`
