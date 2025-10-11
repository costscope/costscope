---
title: Architecture - API Layer
description: External interface, middleware stack, authentication and real-time channels.
---

# API Layer

Focus: request handling, authentication, authorization, streaming progress updates, job lifecycle visibility.

## Responsibilities
- REST endpoints (OpenAPI-defined)
- WebSocket session & push notifications
- Middleware: auth, RBAC, rate limiting, tracing, logging
- Async job submission & status polling

## Key Packages
- `internal/api/handlers` – endpoint implementations
- `internal/api/middleware` – cross-cutting concerns
- `internal/api/websocket` – realtime interface
- `internal/api/jobs` – async orchestration hook layer

## Data Contracts
Stable response envelope: .

```bash
{ success, data, error, meta{request_id,timestamp} }
```

## See also
- `core-services.md`
- `data-layer.md`
- `security-model.md`
