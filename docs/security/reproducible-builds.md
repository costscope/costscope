---
title: Reproducible Builds
description: Deterministic build process, version metadata embedding, verification workflow and troubleshooting.
---

# Reproducible Builds & Version Metadata

This document describes how CostScope embeds and verifies build metadata, and how to produce reproducible (deterministic) release artifacts.

## Version Metadata

Every binary embeds four immutable fields (set at build time via `-ldflags`):

| Field | Description |
|-------|-------------|
| version | Semantic version (e.g. `1.0.0`) |
| commit | Git commit SHA used for the build |
| build_date | UTC timestamp (RFC3339) or epoch-derived (when reproducible) |
| go_version | Go toolchain used |

Show (JSON by default):

```bash
./costscope version
# {"version":"1.0.0","commit":"abc123def456","build_date":"2025-08-26T12:34:56Z","go_version":"go1.22.5"}
```

Human readable:

```bash
./costscope version --human
# CostScope 1.0.0
# commit: abc123def456
# build_date: 2025-08-26T12:34:56Z
# go_version: go1.22.5
```

## Deterministic / Reproducible Release Builds

Reproducible builds use a fixed source date to normalize timestamps so two builds of the same tree yield identical SHA256 checksums.

### One‑Shot Reproducible Build

```bash
SOURCE_DATE_EPOCH=1716249600 make build-release
sha256sum costscope
```

Run again with the same `SOURCE_DATE_EPOCH` and you should obtain the identical hash (byte-for-byte).

### Automated Repro Check Script

`scripts/repro-check.sh <epoch>` performs two sequential builds (with a 1s sleep between) using a fixed SOURCE_DATE_EPOCH and compares their SHA256 digests:

```bash
./scripts/repro-check.sh 1716249600
# [repro] SUCCESS: sha256 <hash>
```

Exit codes:
- 0: Reproducible (hashes match)
- 1: Drift detected (hashes differ)
- >1: Script or build error

### Requirements & Tips

1. Clean Git tree ( empty) – uncommitted changes break provenance.

```bash
git status
```
2. No embedded build timestamps outside ldflags – avoid calling `time.Now()` during variable initialization for metadata.
3. Avoid non‑deterministic compression settings (e.g. timestamps in compressed artifacts). The release target configures stable flags.
4. Go module download cache should not inject variability; module versions are pinned in `go.sum`.

### Integrating Into Release Workflow

During the release process (see `docs/release/checklist.md`):
1. Run the reproducibility script (or CI job) with the chosen epoch (e.g. tag time truncated to midnight UTC).
2. Publish the resulting SHA256 alongside binary artifacts in the GitHub Release.
3. (Optional) Provide the epoch used so external verifiers can reproduce.

### Troubleshooting Non-Determinism

| Symptom | Possible Cause | Mitigation |
|---------|----------------|-----------|
| Hashes differ but code unchanged | Embedded timestamp outside ldflags | Remove runtime timestamp emission from static vars |
| Different hash per machine | Non‑pinned transitive module | Ensure `go.sum(see example below)go mod verify` |
| Repro script exits 1 sporadically | File permission drift or leftover artifacts | (see example below) then retry |


```bash
committed; run
```

```bash
git clean -fdx
```
### Verification by Consumers

Consumers can verify authenticity & reproducibility by:
```bash
curl -L -o costscope <download-url>
echo "<expected_sha256>  costscope" | sha256sum -c -
```

If you publish the `SOURCE_DATE_EPOCH` they may rebuild and compare.

---
## See also
- `../release/checklist.md`
- `../release/promotion.md`
- `supply-chain.md`
- `../architecture/overview.md`

For API contract stability see `../api/contract-guard.mdsupply-chain.md`.

```bash
(if present) and supply chain guard steps in
```
