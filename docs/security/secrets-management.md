---
title: Secrets Management & Rotation
description: Principles and rotation playbook for secrets and credentials.
---
## Principles
- Least privilege: only required scope for each component.
- No secrets in repo (enforced via `gitleaks`).
- Short TTL & rotation automation preferred to static long‑lived keys.
- Deterministic secret loading order: ENV → runtime injection → config (no plaintext defaults).

## Secret Types
| Type | Examples | Rotation Target |
|------|----------|-----------------|
| API Keys | External SaaS, webhooks | 30–90d |
| JWT Signing | Auth tokens | 60d (rolling overlap) |
| DB Credentials | Metadata store (SQLite usually none) | On compromise / password change |
| Encryption Keys | Future at-rest features | 90d |

## Storage Backends (Recommended)
| Backend | Use Case | Notes |
|---------|---------|------|
| AWS Secrets Manager / GCP Secret Manager / Azure Key Vault | Managed secret lifecycle | Native rotation hooks |
| Kubernetes Secrets + sealed-secrets | Cluster deployment | Encrypt at rest, RBAC gate |
| Vault | Enterprise centralized policy | Dynamic credentials potential |

## Environment Variable Loading
Critical variables (examples):
| Var | Purpose |
|-----|--------|
| `COSTSCOPE_RBAC_AUDIT_MODE` | Soft-deny audit passthrough |
| `COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION` | Diagnostic toggle |
| Future: `COSTSCOPE_JWT_SECRET` | JWT signing key (32+ random bytes) |

## Rotation Playbook
1. Generate new secret (e.g., ).

```bash
openssl rand -base64 48
```
2. Inject into secret manager as `newcurrent` valid).

```bash
version (keep
```
3. Deploy application reading both (grace period window).
4. Invalidate old version; confirm metrics/traces show only new usage.
5. Remove old version after retention window.

## Incident Response
If leak suspected: revoke or delete secret, re-issue, rotate dependent tokens, review access logs & RBAC audit span samples.

## Tooling
- `gitleaksgitleaks-report.json`)

```bash
CI job (report in
```
- Optional future: commit hook scanning staged diffs.

Keep updated when auth modes or secret variables expand.
