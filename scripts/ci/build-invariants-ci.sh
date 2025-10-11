#!/usr/bin/env bash
set -euo pipefail

go build -tags duckdb -o bin/invariants-ci ./scripts/tools/invariants-ci
