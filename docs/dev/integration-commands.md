---
title: Integration Commands (Generated)
description: Placeholder for auto-generated integration command documentation.
---
## Overview
Integration command trees are generated from action specification files (YAML/JSON) describing category, use string, flags, and examples. This avoids manual Cobra boilerplate and ensures consistent UX across heterogeneous external system operations (alerts, connections, sync jobs).

## Generation
```bash
make gen-commands          # regenerates all spec-driven builders (analytics/multicloud/etc.)
make gen-commands-drift    # regenerate + fail if diff (CI)
```
Integration specs live under `cmd/modules/integration/`:
| Path | Purpose |
|------|---------|
| `integration_action_specs.go` | Loads spec set & constructs command tree |
| `integration_action_specs_test.go` | Integrity tests (presence / duplicates) |
| `alerts/` | Advanced alert rule & channel commands |
| `connections/` | Connection manager (list/create/update) |

## Action Spec Schema (Simplified)
| Field | Meaning |
|-------|---------|
| `Use(see example below)Use` string |
| `Short` | Short description |
| `Long` | Long help (optional) |
| `Example` | Example usage (optional) |
| `Category` | High-level grouping (alerts, connection, sync) |
| `Params` | Flag definitions (name, type, default, required) |


```bash
| Cobra
```
## Adding a New Action
1. Extend spec list in `integration_action_specs.go` (or load from external file in future).
2. Add validation / test case in `integration_action_integrity_test.go`.
3. Regenerate the command builders and verify there are no diffs.

```bash
make gen-commands-drift
```
4. Update docs here if new category introduced.

## Testing
Integrity tests assert:
- No duplicate `Use` strings
- Category path uniqueness
- Required fields present

End-to-end: run a dry invocation with `--help` to verify flag wiring.

## Examples
```bash
costscope integration alerts create --name high-cost --threshold 5000 --window 1h
costscope integration connections list
```

## Drift Prevention
CI workflow calls the generator and enforces regenerated builders are committed.

```bash
make gen-commands-drift
```

## Backlog / Future
- Externalize action specs to `configs/integration/` for hot-reload
- Add machine-readable export  for `integration` root command

```bash
--format json
```
- Auto-generate markdown command reference page (pipe to docs)

Keep updated as new integration categories ship.
