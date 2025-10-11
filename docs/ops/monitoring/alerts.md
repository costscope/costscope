# Alerting rules and rationale

This document lists alerting rules, thresholds, and escalation policies.

Example alert:

- id: `high_cpu_utilization`
  - expression:

```bash
avg by (instance) (rate(costscope_system_cpu_seconds_total[5m])) > 0.8
```
  - severity: `warning`
  - for: `5m`
  - description: High sustained CPU usage on an instance
  - runbook: `../runbooks/api-latency.md`

Guidance: keep alerts actionable, have a single responsible team, and avoid alert fatigue.
