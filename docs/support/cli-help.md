---
title: CLI Help & Examples
description: Frequently used CostScope CLI commands and experimental feature flags.
---
## Core Workflow

1. Convert raw billing export → FOCUS parquet
2. Validate (schema + invariants)
3. Analyze / Diff / Reports
4. (Optional) Serve API / Streaming

### Quick Start
```bash
costscope focus convert --provider aws --input cur.csv.gz --output focus.parquet
costscope focus validate focus.parquet --all --format table
costscope analytics analyze focus.parquet --output analysis.json
```

### Global Flags (excerpt)
| Flag | Description |
|------|-------------|
| `--config` | Path to YAML config (merged with ENV + flags) |
| `--log-level` | debug|info|warn|error |
| `--format` | output format (table,json,yaml) where supported |
| `--output` | Write result to file (format inferred by extension) |
| `--experimental` | Enable experimental analytics blocks |
| `--use-focus-engine` | Activate extended focus engine phases (anomalies/forecast/optimizations) |

### Major Command Groups
| Group | Purpose | Notes |
|-------|---------|-------|
| `focus` | Conversion, validation, jobs | Streaming + batch modes |
| `analytics` | Core analytics (diff, analyze, convert) | Generated + hand-written mix |
| `analytics-complex(see example below)--experimental` |
| `reports` | Report & metadata persistence | Output dir + metadata store |
| `providers` | Provider info & sync utilities | Lists versions/mappers |
| `streaming` | Manage streaming sessions | Enterprise / experimental gating |
| `production` | Readiness, diagnostics | Health, config, invariants guards |
| `integration` | External systems (alerts, connections) | Spec-driven builders |
| `invariants` | Invariants regenerate/diff | Parity + drift control |
| `config` | Config inspection & version | Precedence debug |
| `api` | Serve HTTP/WS API | Use with container or standalone |


```bash
| Advanced composite analytics | May require
```
### Focus Conversion
Single file:
```bash
costscope focus convert --provider aws --input cur.csv.gz --output focus.parquet --streaming
```
Batch (rotate large export into multiple parquet shards):
```bash
costscope focus convert --provider aws --input cur.csv.gz --output focus_fast.parquet --streaming --rotate-size 134217728
```
Unified mapper parity testing (env flag):
```bash
COSTSCOPE_USE_UNIFIED_MAPPER=1 costscope focus convert --provider aws --input cur.csv.gz --output focus_unified.parquet --streaming
```

### Validation
```bash
costscope focus validate focus.parquet --all --format table
costscope focus validate focus.parquet --all --output validation.json   # JSON (by extension)
costscope focus validate focus.parquet --all --output validation.html   # HTML
costscope focus validate focus.parquet --all --output validation.csv    # CSV
costscope focus validate batch ./focus-dir --pattern "*.parquet" --output-dir ./validation-reports
```
Note: older flags such as `--report-html--report-csv--json--output <file>--format`.

```bash
,
```

```bash
were replaced; prefer
```

```bash
or
```

### Analytics & Diff
```bash
costscope analytics analyze focus.parquet --output analyze.json
costscope analytics diff old.parquet new.parquet --insights --forecast-periods 30 --output diff.json
```

### Reports
```bash
costscope reports generate focus.parquet --output-dir reports/
```
Metadata store (SQLite) if `reports.metadata_store_path` set in config | env.

### Invariants & Parity
```bash
costscope invariants regenerate focus.parquet --output invariants_current.json --tolerance 0.01
costscope invariants diff invariants_current.json --baseline tests/fixtures/quality/baseline_synth_invariants.json --tolerance 0.01 --report invariants.json
```

### Performance / Bench Helpers
See `dev/performance-benchmarks.md` for thresholds. Common shortcuts:
```bash
make perf-bench            # regression guard (previous vs unified)
make perf-short            # quick 3-iteration bench
make parity-check          # aggregate parity (effective_cost, usage_quantity, records)
make perf-parity           # short bench + parity
```

### Experimental Focus Engine
Add extended JSON block when combined with  or file output:

```bash
--format json
```
```bash
costscope analytics analyze focus.parquet --use-focus-engine --output analysis.json
```
Extended block: (see example below).


```bash
extended.anomalies | forecasts | trends | optimizations | key_findings | recommendations | per_phase
```
### Configuration Inspection
```bash
costscope config version           # show embedded version/build info
costscope config dump --format yaml
```

### API / Streaming
```bash
costscope api serve --listen :8080 --metrics --enable-websocket
costscope streaming start focus.parquet --session my-run --follow
```

### Health & Diagnostics
```bash
curl -s localhost:8080/health/ready
curl -s localhost:8080/metrics | grep costscope_
```

### Make Target Discovery
```bash
make help       # human readable
make help-json  # machine readable JSON
```

### Exit Codes (selected)
| Context | Code | Meaning |
|---------|------|---------|
| Parity guard | 2 | Parity mismatch |
| Invariants diff | 3 | Invariants drift |
| Perf bench | 1 | Ratio threshold exceeded |

### Tips
- Prefer  over format flags where possible.

```bash
--output file.ext
```
- Use environment for experimental gating in CI to avoid accidental activation.
- Record baseline artifacts (bench_results.json, invariants.json) when diagnosing regressions.

See also: `dev/performance-benchmarks.mddev/performance-engine.mdintegration-commands.md`.

### Metrics & Tracing (excerpt)
- Prometheus: `costscope_http_requests_totalcostscope_http_request_duration_seconds`.
- Domain: converter / exporter / streaming metrics (see ops docs).
- RBAC: `costscope_rbac_checks_totalcostscope_rbac_audit_soft_denies_total`.

Health endpoints: `/health/health/live/health/ready`.

This page is synthesized from command builders & repo conventions; keep updated when adding new top-level command groups.

### CLI flags & examples (detailed)

Below are copy-paste friendly examples that show commonly used flags related to conversion, invariants and rotation. They are intentionally explicit so CI / automation can reuse them directly.

```bash
# Convert sample AWS CUR → FOCUS Parquet (streaming, invariants, rotation)
./costscope focus convert \
  --provider aws \
  --input ./aws-cur.csv.gz \
  --output ./focus.parquet \
  --streaming \
  --invariants \
  --invariants-report ./invariants_report.json \
  --rotate-size 134217728 \
  --quiet

# Notes:
# - --provider: aws|azure|gcp
# - --streaming: streaming conversion mode (recommended for large files)
# - --invariants: enable invariants computation during conversion
# - --invariants-report: path to write invariants JSON report
# - --invariants-baseline: (optional) baseline invariants JSON for comparisons
# - --fail-on-invariants: exit non-zero if invariants drift beyond tolerance
# - --rotate-size: rotation threshold in bytes (use -1 to disable rotation)
# - --quiet: reduce CLI output for automated runs
```

### Invariants drift — quick fail example

This minimal example demonstrates how to run a conversion that will fail when invariants drift beyond the configured tolerance. It is useful for CI gates that must block on data quality regressions.

```bash
# Run conversion and fail if invariants drift vs a strict baseline
./costscope focus convert \
  --provider aws --input ./aws-cur-modified.csv.gz --output ./out.parquet \
  --streaming --invariants --invariants-report ./inv_curr.json \
  --invariants-baseline tests/fixtures/quality/baseline_synth_invariants.json \
  --fail-on-invariants --quiet

# Expected behavior:
# - If computed invariants deviate beyond tolerance (default 1%), the CLI exits non-zero
#   and returns a short reason (example: "invariants_drift"). Inspect ./inv_curr.json for details.
```
