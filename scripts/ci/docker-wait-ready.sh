#!/usr/bin/env bash
set -euo pipefail

# docker-wait-ready: Wait for an HTTP endpoint to become ready.
# Usage: docker-wait-ready.sh <url> [retries] [sleep_seconds]
# Example: docker-wait-ready.sh http://127.0.0.1:8080/health/ready 20 3

URL=${1:-}
RETRIES=${2:-20}
SLEEP=${3:-3}

if [[ -z "$URL" ]]; then
  echo "Usage: $0 <url> [retries] [sleep_seconds]" >&2
  exit 2
fi

for i in $(seq 1 "$RETRIES"); do
  if curl -sS "$URL" >/dev/null 2>&1; then
    echo "service ready at $URL"
    exit 0
  fi
  echo "waiting for service ($i/$RETRIES)..." >&2
  sleep "$SLEEP"
done

echo "service not ready after $RETRIES attempts: $URL" >&2
exit 1
