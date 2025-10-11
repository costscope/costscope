# Invariants baselines (examples)

This directory contains small example invariants baseline JSON files intended for smoke testing and documentation.

Files:

- `baseline_aws_smoke.json` — small baseline matching `tests/fixtures/aws/cur_smoke.csv` example.
- `baseline_azure_smoke.json` — small baseline matching `tests/fixtures/azure/usage_smoke.csv` example.
- `baseline_gcp_smoke.json` — small baseline matching `tests/fixtures/gcp/usage_smoke.csv` example.

How to use:

1. Run convert with invariants and write report:

```bash
./costscope convert --provider aws --input tests/fixtures/aws/cur_smoke.csv --output ./tmp/focus_aws_smoke.parquet --streaming --invariants --invariants-report ./tmp/inv_aws.json --quiet
```

2. Compare run report to baseline using `invariants diff` (or use `jq` to inspect fields):

```bash
# Using the CLI diff (requires ./costscope built with duckdb tag for regenerate)
./costscope invariants diff ./tmp/inv_aws.json --baseline examples/invariants/baseline_aws_smoke.json --tolerance 0.01 --report ./tmp/inv_diff.json

# Quick jq check
jq '.row_count, .sum_effective_cost, .negative_usage_violation_count' ./tmp/inv_aws.json
```

3. Regenerate baseline from a produced FOCUS file:

```bash
# Makefile helper (invokes ./costscope invariants regenerate)
make invariants-baseline file=./tmp/focus_aws_smoke.parquet out=./tmp/baseline_aws_smoke.json tol=0.01
```

Notes:
- These are small example baselines for documentation and smoke tests only. Do not treat them as canonical production baselines.
- If you want to update baselines, regenerate from representative production FOCUS outputs and commit a well-reviewed replacement.
