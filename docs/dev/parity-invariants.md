# Parity & Invariants Guards

This document describes how CostScope validates numerical consistency across the legacy (fast) and unified mapper conversion paths and guards aggregate dataset invariants over time.

## Overview

Pipeline (synthetic dataset ~20k rows by default):

1. Build required binaries (slim for parity, debug + optimized DuckDB for invariants).
2. Generate two parquet outputs:
   - `focus_fast.parquet` (fast path)
   - `focus_unified.parquet` (unified mapper)
3. Run `parity-check` tool (lite hash + aggregate comparisons) → produces `parity.json`.
4. Regenerate current invariants from latest fast parquet → `invariants_current.json`.
5. Diff against baseline invariants (`tests/fixtures/quality/baseline_invariants.json`) → `invariants.json`.

## Make Targets

| Target                       | Purpose                                                                                  |
| ---------------------------- | ---------------------------------------------------------------------------------------- |
| `prepare-parity-binaries`    | Build slim, optimized DuckDB, and debug DuckDB binaries (single pass).                   |
| `parity-json`                | Produce parity artifacts (`focus_fast.parquet`, `focus_unified.parquet`, `parity.json`). |
| `parity-smoke`               | Quick parity check on small dataset (fast pre-commit).                                   |
| `invariants-guard`           | Regenerate invariants and diff vs baseline.                                              |
| `data-parity-guard`          | Composite: `parity-json` + `invariants-guard`.                                           |
| `data-parity-smoke`          | Composite smoke: runs `parity-smoke`.                                                    |
| `invariants-update-baseline` | Recompute baseline invariants JSON from current fast path output.                        |

## Exit Codes

| Code | Source                      | Meaning                                                   |
| ---- | --------------------------- | --------------------------------------------------------- |
| 0    | all                         | Success / no drift                                        |
| 1    | scripts / make              | Generic failure (setup, missing file, unexpected)         |
| 2    | parity / invariants scripts | Parity mismatch OR convert/regenerate failure (non-drift) |
| 3    | invariants guard            | Invariants drift detected (semantic difference)           |

## Files

| File                                              | Description                                                           |
| ------------------------------------------------- | --------------------------------------------------------------------- |
| `parity.json`                                     | Summary of fast vs unified aggregates + lite hashes.                  |
| `invariants_current.json`                         | Re-generated invariants from latest fast parquet.                     |
| `invariants.json`                                 | Diff report vs baseline (only saved if drift? always saved on pass?). |
| `tests/fixtures/quality/baseline_invariants.json` | Authoritative baseline invariants (large synthetic dataset).          |

## Baseline Management

Regenerate baseline (after intentional data/model change):

```bash
make prepare-parity-binaries
make invariants-update-baseline
```

Commit the updated `baseline_invariants.json` with a concise explanation (e.g. mapper logic change, dataset scale change).

Heuristic sanity check (run automatically in workflow):

- `scripts/guards/check_baseline_sanity.sh` warns if `row_count` below `MIN_BASELINE_ROWS` (default 1000) while synthetic dataset is expected.

## Adding New Invariants

1. Extend invariants generation logic in the CLI (DuckDB path) to compute the new metric.
2. Update diff logic to compare and add violation specification.
3. Regenerate baseline (`make invariants-update-baseline`).
4. Document the invariant and rationale here.

## Performance Notes

- Slim binary is CGO-disabled for rapid parity conversion.
- DuckDB builds (optimized + debug) are reused for invariants; duplicates avoided via `prepare-parity-binaries`.
- Rotate size for parquet conversion is configurable via `PARQUET_ROTATE_SIZE` (default `10000000000`). Increase to reduce splits on large datasets; decrease to lower memory burst.
- Smoke guard (`parity-smoke` / `data-parity-smoke`) uses a tiny dataset (50 rows) and is suitable for pre-commit hooks.

## Pre-commit Integration (optional)

To catch regressions before pushing, you can enable a local pre-commit hook that runs the fast smoke guard:

1. Ensure you have pre-commit installed locally.
   - macOS: `brew install pre-commit`
2. In the repository root, a `.pre-commit-config.yaml` provides a local hook:
   - Hook: "Data parity smoke guard"
   - Command: `make data-parity-smoke`
3. Install the hooks:
   - `pre-commit install`

Notes:

- The hook does not pass filenames and runs the smoke dataset path only (quick check).
- If you need to tune performance, you may set `PARQUET_ROTATE_SIZE` in your environment.

## Troubleshooting

| Symptom                                                    | Likely Cause                         | Action                                                         |
| ---------------------------------------------------------- | ------------------------------------ | -------------------------------------------------------------- |
| Exit code 2, "Parity mismatch"                             | Unified vs fast aggregate divergence | Inspect `parity.json`, review recent conversion logic changes. |
| Exit code 3, invariants drift with massive row_count delta | Outdated baseline                    | Regenerate baseline; confirm dataset selection.                |
| Missing binaries error                                     | Skipped `prepare-parity-binaries`    | Run it or ensure workflow includes it.                         |
| Baseline sanity warning                                    | Tiny baseline or missing row_count   | Regenerate baseline.                                           |

## Future Enhancements

- Dual-baseline strategy (small smoke + large full) to reduce average guard runtime.
- Ratio-based tolerance gating for specific invariants.
- Structured machine-parsable drift summary artifact.

---

Last updated: 2025-10-14
