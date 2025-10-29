#!/usr/bin/env bash
set -euo pipefail

# Quick module sanity check used in lightweight CI jobs.
go list ./... >/dev/null
echo "[module-sanity] go list OK"
