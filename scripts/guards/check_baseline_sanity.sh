#!/usr/bin/env bash
set -euo pipefail
# check_baseline_sanity.sh - Quick heuristic sanity check for invariants baseline size.
# Warns (exit 0) if baseline appears implausibly small vs expected synthetic dataset.
# Fails (exit 1) only if file missing or unreadable.

BASELINE=${INVARIANTS_BASELINE:-tests/fixtures/quality/baseline_invariants.json}
MIN_ROWS=${MIN_BASELINE_ROWS:-1000} # threshold to warn

if [[ ! -f "$BASELINE" ]]; then
  echo "[baseline-sanity] ERROR: baseline file missing: $BASELINE" >&2
  exit 1
fi

rows=$(jq -r '.row_count // empty' "$BASELINE" 2>/dev/null || true)
if [[ -z "$rows" ]]; then
  echo "[baseline-sanity] WARNING: cannot read row_count in $BASELINE" >&2
  exit 0
fi
if (( rows < MIN_ROWS )); then
  echo "[baseline-sanity] WARNING: baseline row_count=$rows < threshold=$MIN_ROWS (likely outdated tiny baseline)" >&2
  echo "[baseline-sanity] Suggest running: make invariants-update-baseline" >&2
fi

exit 0
