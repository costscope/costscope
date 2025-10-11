---
title: Architecture - Security Model
description: Layered security controls: authN, RBAC, rate limiting, audit logging.
---

# Security Model

## Layers
1. Transport (TLS / ingress)
2. Authentication (JWT, API keys)
3. Authorization (RBAC)
4. Input Validation
5. Rate Limiting
6. Audit Logging

## RBAC
Role-policy engine with soft audit mode prior to enforcement.

## Audit Events
Structured log channel; correlation IDs propagate through stack.

## See also
- `api-layer.md`
- `core-services.md`
- `../security/supply-chain.md`
