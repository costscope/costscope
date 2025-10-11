---
title: Compliance Notes
description: SOC2 and GDPR control mapping and logging guidance.
---
## SOC2 Focus Areas
| Domain | Control Theme | Implementation Notes |
|--------|---------------|----------------------|
| Security | Access control | RBAC + audit mode soft denies for staged rollout |
| Availability | Monitoring | Prometheus metrics + health endpoints + SLO doc |
| Processing Integrity | Data accuracy | Invariants + parity guards in CI |
| Confidentiality | Secrets handling | External secret stores, no inline defaults |
| Privacy | Data minimization | Only billing usage fields persisted (no PII) |

## GDPR Considerations
No direct personal data processed; ensure logs avoid embedding customer PII. If multi-tenant tag evolves to include emails/names, revisit DPIA.

## Logging Guidance
- Structured JSON (when logger configured), avoid sensitive values
- Include request_id for correlation
- Redact secrets by key pattern at config load

## Data Retention
Bench / invariants artifacts not shipped to production by default. If exported for audits, apply retention (≤90 days) and access control.

## Future Controls
- Signed provenance (SLSA) for build pipeline
- OPA policies for config validation
- Automated dependency license classification gating

Keep updated when compliance scope changes.
