---
title: ADR 0002 - JobStore Backend Choice
date: 2025-09-08
status: proposed
---

# ADR 0002: JobStore Backend Choice

Decision: Prefer BoltDB (embedded, low ops) for initial durable JobStore; abstract via interface to allow future migration to SQLite or badger.

Context: Need durable, small-footprint persistence for async conversion job metadata without introducing external DB. Bolt offers simplicity and performance; SQLite is considered later for richer queries.

Consequences:
- Provide compaction & pruning runbook (`docs/runbooks/bolt-jobstore-maintenance.md`).
- Ensure backup & migration tooling exists.
