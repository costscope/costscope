---
title: ADR 0001 - Unified Mapper Adoption
date: 2025-09-08
status: proposed
---

# ADR 0001: Unified Mapper Adoption

Decision: Consolidate the previous `fastunified` mapper code path by default, guarded by perf & parity checks in CI.

```bash
mapper and experimental paths into a single
```

Context: Two separate mapping implementations caused maintenance overhead and divergent behaviors. Parity tests and perf benchmarks will prevent regressions during migration.

Consequences:
- CI must run parity & perf checks on changes touching conversion code.
- Baseline artifacts must be versioned and occasionally updated with explicit review.
- Rollback strategy must remain available (revert to fast mapper binary) for production incidents.
