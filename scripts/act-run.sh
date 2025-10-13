#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR"; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"

# Wrapper for invoking `act` that ensures containers receive a hosts mapping for 'mic'.
# Usage: scripts/act-run.sh [act-args...]

HOST_ARG="--add-host=mic:127.0.0.1"

args=()
host_present=0

for a in "$@"; do
  args+=("$a")
  if [[ "$a" == "$HOST_ARG" || "$a" == --add-host=*mic* ]]; then
    host_present=1
  fi
done

if [ "$host_present" -eq 0 ]; then
  # Insert HOST_ARG before the final `--` (act uses `--` to forward to container runtime),
  # or append it if not present.
  inserted=0
  out_args=()
  for a in "${args[@]}"; do
    if [ "$a" = "--" ] && [ "$inserted" -eq 0 ]; then
      out_args+=("$HOST_ARG")
      inserted=1
    fi
    out_args+=("$a")
  done
  if [ "$inserted" -eq 0 ]; then
    out_args+=("$HOST_ARG")
  fi
else
  out_args=("${args[@]}")
fi

ci::log "Executing act with args: ${out_args[*]}"
exec act "${out_args[@]}"
