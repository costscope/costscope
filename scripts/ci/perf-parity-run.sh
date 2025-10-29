#!/usr/bin/env bash
set -euo pipefail

# Wrapper to run perf + parity checks consistently from CI and locally.
# Keeps workflow YAML clean and avoids duplicated inline run blocks.

echo "[perf-parity] starting"
make perf-parity
echo "[perf-parity] completed"
