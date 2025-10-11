---
title: Documentation & Code Style Guide
description: Conventions for markdown docs, headings, front matter, code blocks, and commit messages.
---

# Style Guide

## Markdown Front Matter
Required for new top-level architecture, security, release, runbook, and dev docs:
```
---
title: <Short Title>
description: <Concise one-line purpose.>
last_reviewed: YYYY-MM-DD
---
```
Omit `last_reviewed` only for auto-generated docs (the generation pipeline should stamp it).

## Headings
Use `##########`).

```bash
once per file (title). Start major sections with
```

```bash
. Avoid skipping levels (no
```

```bash
without a preceding
```

## Tone & Voice
Concise, task-oriented, present tense. Avoid marketing adjectives. Use tables for matrices/thresholds.

## Code Blocks
Always specify language (`bashgoyamljsonsql`) to improve syntax highlighting and downstream tooling.

```bash
,
```

## Tables
Left-align all columns unless numeric-only, then right-align (optional). Keep headers short (`MetricPerformance Metric Name`).

```bash
, not
```

## Linking
Prefer relative links (`../architecture/overview.mddocs-link-check`).

```bash
) over absolute GitHub URLs for portability. Do not link to deleted or renamed file names. Run link check before merge (TODO: add CI target
```

## File Naming
Kebab-case for new docs: `focus-conversion.mdperformance-benchmarks.md`. Use existing directory taxonomy (architecture/, dev/, security/, release/, runbooks/, support/, providers/, tests/).

## Front Matter Keys
- `title` (required)
- `description` (required)
- `last_reviewed` (optional, ISO date)
- Additional keys allowed but must not conflict with build tooling.

## Diagrams & Assets
Place images under `docs/assets/.png.svg![Pipeline](../assets/focus_pipeline.png)`.

```bash
with descriptive kebab names. Use
```

```bash
for raster,
```

```bash
for diagrams. Reference relatively:
```

## Performance Data
When updating `tests/perf/baseline_bench_results.jsondev/performance-benchmarks.md` if thresholds shift.

```bash
, also update the table in
```

## Commit Messages
Conventional subset: `feat:fix:docs:refactor:perf:test:chore:docs: update conversion guide ratios`.

```bash
. Scope optional. Imperative mood:
```

## Removed / Outdated Content
Remove placeholder stubs entirely; do not leave "Moved from root" comments. Use a PR note if a widely referenced doc is relocated. Do not reference old file names; update links to point to the current canonical path.

## Review Checklist (Docs PR)
- Front matter present & current
- No references to old/deleted file names
- Links resolve locally
- Tables render (no broken pipes)
- Code blocks have language

## Automation (Planned)
-  — validate internal links

```bash
make docs-link-check
```
-  — check mandatory front matter keys

```bash
make docs-style-guard
```

## Contributing to Docs

Workflow:
1. Create a topic branch `docs/<short-topic>`.
2. Add or update markdown under `docs/` following front matter & naming rules.
3. Run locally:  and .

```bash
make check-docs
```
4. Include generated diagram artifacts if diagrams changed and document the generation command in `docs/DIAGRAMS.md`.

ADRs:
- Add new ADRs under `docs/adr/` with a numeric prefix and context. Keep ADRs short — problem, decision, consequences. Link ADRs from the relevant docs (architecture, release notes) when applicable.


Keep this guide lean. Expand only when friction recurs.
