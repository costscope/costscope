---
title: Integration Helpers
description: Shared helper functions and patterns for integration command modules.
---
## Purpose
Centralize helper utilities for integration action specs (validation, flag coercion, output formatting) to reduce duplication across generated command groups.

## Key Helpers (Examples)
| Helper | Responsibility |
|--------|----------------|
| Param normalization | Enforce lower-case categories, trim spaces |
| Dotted role allowance | Permit future `role:admin` style tokens (validated downstream) |
| Spec integrity checks | Detect duplicate use strings / missing fields |

## Testing Pattern
Table-driven tests assert accepted vs rejected parameter shapes; integrity tests ensure uniqueness & schema adherence.

## Extension Guidelines
1. Avoid side effects (pure functions) for easier testing
2. Keep logging out of helpers (caller decides)
3. Re-export only stable helper names (avoid leaking experimental API)

## Backlog
- Add JSON schema for spec validation
- Provide dry-run mode to emit planned command tree without generation

Keep updated as integration surface expands.
