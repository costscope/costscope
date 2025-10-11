---
title: Bolt JobStore Maintenance
description: Operational guidance for pruning and compacting the Bolt-backed conversion JobStore.
last_reviewed: 2025-09-08
---
## Purpose
Periodic maintenance for the BoltDB-backed JobStore used by asynchronous conversion jobs (compaction + prune historical entries) to maintain bounded disk usage and read latency.

## Symptoms Requiring Maintenance
| Symptom | Indicator |
|---------|----------|
| Slow job list queries | p95 > 300ms for job enumeration |
| Large DB file | Size grows while job retention window small |
| Fragmentation | High free pages reported by stats tool |

## Retention Policy
Environment / config value controls max completed job age (e.g., 168h). Jobs older than window pruned during maintenance run.

## Maintenance Procedure
1. Put system in maintenance mode (optional) to reduce write churn
2. Run pruning:

```bash
costscope focus jobs prune --older-than 168h
```
3. Compact:
	- Stop writer processes if possible
	- Copy live DB to temp (bolt recommendation)
	- Open copy and sequentially rewrite buckets (future automated tool)
4. Swap compacted file atomically
5. Restart services & verify `/metrics` (job store gauge sizes)

## Metrics
| Metric | Description |
|--------|-------------|
| `costscope_jobs_total` | Total jobs (active + completed) |
| `costscope_jobs_pruned_total` | Cumulative pruned jobs |
| `costscope_jobstore_compact_duration_seconds` | Duration of compaction (planned) |

## Backup
Bolt file can be snapshotted via filesystem copy (no special tooling). Ensure copy while writers quiescent to reduce risk of partially flushed pages.

## Failure Recovery
If compaction corrupts file:
1. Restore last backup copy
2. Re-run pruning only
3. Open issue with logs & bolt stats

## Future Enhancements
- Online compaction without full copy
- Partitioned job buckets by month

Review quarterly or after Bolt library upgrades.
