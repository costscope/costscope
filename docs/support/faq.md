---
title: Frequently Asked Questions
description: Quick answers about building, running, operating, and contributing.
---

# FAQ

## Getting Started
**Q:** How do I build the CLI from source?
**A:**

```bash
make build
# or
make
```

**Q:** Is Docker the recommended runtime?
**A:** For production use the published image (see `../ops/`), for local iteration native build is fine.

**Q:** Where is the high‑level architecture?
**A:** See `../architecture/overview.md`.

## Configuration
**Q:** Where do configuration files live?
**A:** Examples in `configs/`. Precedence: explicit flag > env var > config file default.

**Q:** How do I list all configurable options?
**A:** Run  or subcommand `--help`.

```bash
costscope --help
```

## Data & Storage
**Q:** What formats are supported for analysis input?
**A:** Parquet (primary), plus experimental raw formats via helper commands.

**Q:** Where are temporary or cache files written?
**A:** `tmp/--work-dir`).

```bash
by default (override via
```

## Performance
**Q:** How do I profile?
**A:** Use the perf bench targets or run with profiling flags (see `../ops/monitoring/`).

```bash
make perf-bench
```

**Q:** Quick wins if a run feels slow?
**A:** Prefer Parquet, reduce debug logging, adjust worker concurrency `COSTSCOPE_WORKERS`.

## Troubleshooting
**Q:** Something failed—where to start?
**A:** See `troubleshooting.mdvscode-troubleshooting.md`.

```bash
; for editor issues see
```

**Q:** Logs look truncated.
**A:** Redirect to file or ensure terminal doesn’t soft-wrap aggressively.

## Security & Supply Chain
**Q:** How do I verify the binary?
**A:** Use `checksums.txt.sig`) and SBOM at repo root.

```bash
(+
```

**Q:** Any policies around RBAC?
**A:** See `../security/` for model & policies.

## Contributing
**Q:** Where is the contribution guide?
**A:** `CONTRIBUTING.md` at repo root.

**Q:** Coding style / lint?
**A:** Go: vet & lint; docs: run the docs checks.

```bash
make check-docs
```

**Q:** How do I add a new doc page?
**A:** Place under correct domain folder (lowercase kebab-case); run .

```bash
make check-docs
```

## Release & Versioning
**Q:** Programmatic version?
**A:**  or inspect build metadata in `version.go`.

```bash
costscope version --json
```

**Q:** How are releases signed?
**A:** See release docs (after Step B) + checksum/signature artifacts.

---
### See also
- `support.md`
- `troubleshooting.md`
- `vscode-troubleshooting.md`
- `../architecture/overview.md`
