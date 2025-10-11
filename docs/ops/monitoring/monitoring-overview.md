# Monitoring (Operations)

This directory collects operational documentation for monitoring, metrics catalog, alerting, dashboards, and tracing.

Current files in this area include:

- `../monitoring.md` — general monitoring notes (previous location).
- `../monitoring_implementation_complete.md` — implementation details and examples.

Recommended roadmap

1. Consolidate content into this directory under clear filenames:
   - `implementation.md` — implementation details and examples (canonical).
   - `catalog.md` — metrics catalog and metric names.
   - `alerts.md` — alerting rules and rationale.
   - `dashboards.mdmonitoring/grafana`.

```bash
— dashboard JSON examples and links to
```
   - `tracing.md` — tracing spans and correlation guidance.
2. Update `DOCUMENTATION_INDEX.mdops/monitoring/`.

```bash
to point to the canonical files under
```
3. Keep this README as an index/alias for at least one release cycle while the migration completes.

Migration notes for contributors

- When moving a file, add a short alias file at the old path that points to the new location, and update `DOCUMENTATION_INDEX.md` in the same PR.
- Keep the documentation text in English.
- Use kebab-case for new file names (for example: `implementation.mdalerts.md`).

```bash
,
```

If you'd like, I can now: 1) create the canonical files and migrate content from `monitoring_implementation_complete.mdops/monitoring/implementation.mdDOCUMENTATION_INDEX.md` links. Which do you want next?

```bash
into
```

```bash
(safe migration with alias), or 2) add a small check script to validate
```
