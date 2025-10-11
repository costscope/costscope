#!/usr/bin/env bash
set -euo pipefail

variant="${1:-}"
if [[ -z "$variant" ]]; then
  echo "usage: $0 <variant>" >&2
  exit 2
fi

# Guard: when running under nektos/act only run the slim variant locally (unless ACT_FULL=true)
if [[ "${IS_ACT:-false}" == "true" && "$variant" != "slim" && "${ACT_FULL:-false}" != "true" ]]; then
  echo "[act] Skipping variant '$variant' because IS_ACT=true"
  exit 0
fi
