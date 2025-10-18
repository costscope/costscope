#!/usr/bin/env bash
set -euo pipefail

# retry.sh <retries> <sleep_seconds> <command...>
# Example: retry.sh 2 3 make build-production

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <retries> <sleep_seconds> <command...>" >&2
  exit 2
fi

retries=$1; shift
sleep_s=$1; shift

n=0
until "$@"; do
  exit_code=$?
  n=$((n+1))
  if [[ $n -gt $retries ]]; then
    echo "retry: giving up after $n attempts (last exit=$exit_code)" >&2
    exit "$exit_code"
  fi
  echo "retry: attempt $n failed (exit=$exit_code), sleeping ${sleep_s}s and retrying..." >&2
  sleep "$sleep_s"
  # optional small backoff increment
  sleep_s=$((sleep_s+1))
 done
