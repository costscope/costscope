---
title: Authentication Failures Spike
description: Runbook for diagnosing and remediating spikes in authentication failures (401/403).
last_reviewed: 2025-09-08
---

# Authentication Failures Spike

## 1. Detection
Alerts:
- Auth Failures Spike: >50 failures / min for 5m & >5x 24h baseline.
- Elevated 401/403 ratio while 5xx stable (suggest config / credential issue).

## 2. Initial Triage
1. Confirm spike is not a load test (check deployment / test schedule).
2. Separate 401 (unauthenticated) vs 403 (unauthorized) trends.
3. Examine recent config change (JWT secret rotation, RBAC policy).

## 3. Diagnosis
| Hypothesis | Indicators | Steps |
|-----------|-----------|-------|
| Expired / rotated JWT secret not propagated | All tokens invalid suddenly | Verify env var/secret version across pods |
| Clock skew | Tokens failing `nbf` / `exp` early | Check NTP sync & system time |
| Brute force / attack | High unique IPs; 401 > usual | Rate limit metrics / WAF logs |
| Policy / RBAC misconfig | 403 climbs; 401 normal | Diff current vs last known-good policy file |
| Token signing algorithm mismatch | Signature errors in logs | Confirm alg header matches config |

## 4. Mitigation
- If secret mismatch: redeploy pods with consistent secret; keep old secret temporarily if dual-rotate supported.
- Enable stricter rate limiting / temporarily block abusive IP ranges (WAF / firewall) if attack.
- Revert RBAC policy to last good commit if recent change correlates.
- Fix time skew (restart time sync service, correct host clock) if detected.
- Communicate partial outage to stakeholders if auth unusable >10m.

## 5. Verification
Success:
1. Failure rate returns to baseline (<5/min sustained).
2. 401 / 403 ratios align with historical patterns.
3. No residual elevated latency on auth endpoints.

## 6. Post-Incident
- Add synthetic token validation canary.
- Automate secret rotation smoke test harness.
- Enhance alert to distinguish 401 vs 403 thresholds.

## 7. Prevention
- Implement dual-secret acceptance window for seamless rotation.
- Pre-merge RBAC policy linter / test harness.
- Add anomaly model for geo/IP distribution.

## 8. Escalation
If unresolved after 20m or indicates coordinated attack: escalate to Security & Platform leads; involve incident response process.
