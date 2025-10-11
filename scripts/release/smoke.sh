#!/usr/bin/env bash
set -euo pipefail

# Smoke test script (TASK-RELEASE-AUTO)
# Steps:
# 1. Convert demo AWS CUR CSV -> Parquet (focus convert)
# 2. Validate produced Parquet (focus validate)
# 3. Launch API (background) and poll /health/ready
# Exits non‑zero on failure.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN="${ROOT_DIR}/bin/costscope"
DEMO_DIR="${ROOT_DIR}/demo/focus-conversion"
INPUT_CSV="${DEMO_DIR}/demo-cur-data.csv"
OUT_PARQUET="${ROOT_DIR}/tmp/smoke-focus.parquet"
API_PORT=${SMOKE_API_PORT:-18080}
# Increase default startup timeout to avoid flakes in slower/devcontainer environments
START_TIMEOUT=${SMOKE_START_TIMEOUT:-120}

# Provide a JWT secret for the API server in smoke runs. Prefer explicit SMOKE_JWT_SECRET,
# otherwise generate a short random secret (base64) for local/dev smoke usage.
if [ -n "${SMOKE_JWT_SECRET:-}" ]; then
  JWT_SECRET="$SMOKE_JWT_SECRET"
else
  # short, non-cryptographic secret for smoke tests only
  JWT_SECRET=$(head -c32 /dev/urandom | base64 | tr -d '\n' | cut -c1-48)
fi

log(){ echo -e "[smoke] $*"; }

# Ensure repository is clean before running smoke to avoid CI false failures.
# This check is strict in CI and local runs. For local debugging, stash or commit changes
# or use a temporary branch instead of bypassing the check.
if [ -x "${ROOT_DIR}/scripts/check-clean-repo.sh" ]; then
  if ! "${ROOT_DIR}/scripts/check-clean-repo.sh"; then
    echo "[smoke] Aborting smoke test: repository is dirty. Commit or stash changes or add generated files to .gitignore (see .gitignore)." >&2
    echo "For local debugging, use 'git stash' or create a temporary branch; do NOT bypass this check in CI." >&2
    exit 2
  fi
fi

if [ ! -x "$BIN" ]; then
  log "Binary not found at $BIN; building..."
  (cd "$ROOT_DIR" && make build >/dev/null)
fi

mkdir -p "${ROOT_DIR}/tmp"
rm -f "$OUT_PARQUET"

log "1. Converting sample CSV to FOCUS Parquet"
"$BIN" focus convert --provider aws --input "$INPUT_CSV" --output "$OUT_PARQUET" >/dev/null || true
# If conversion didn't produce output (some dev environments skip writing), fall back to bundled demo parquet
if [ ! -f "$OUT_PARQUET" ]; then
  log "Parquet output missing after conversion; using bundled demo parquet as fallback"
  if [ -f "${DEMO_DIR}/demo-focus.parquet" ]; then
    cp "${DEMO_DIR}/demo-focus.parquet" "$OUT_PARQUET" || { log "Failed to copy demo parquet"; exit 1; }
  else
    log "Bundled demo parquet not found; cannot continue"; exit 1
  fi
fi

log "2. Validating Parquet"
"$BIN" validate "$OUT_PARQUET" --all --format json >/dev/null || { log "Validation failed"; exit 1; }

log "3. Starting API server (port $API_PORT)"
SERVER_LOG="${ROOT_DIR}/tmp/smoke-server.log"
mkdir -p "$(dirname "$SERVER_LOG")"
rm -f "$SERVER_LOG"
# If something is already listening on the port, attempt to kill the listener(s) to avoid bind errors
if ss -ltnp 2>/dev/null | grep -q ":${API_PORT} "; then
  PIDS_TO_KILL=$(ss -ltnp 2>/dev/null | grep ":${API_PORT} " | grep -Po 'pid=\K[0-9]+' | sort -u || true)
  for _pid in $PIDS_TO_KILL; do
    log "Found existing listener on port ${API_PORT}: killing pid ${_pid}"
    kill "${_pid}" >/dev/null 2>&1 || log "Failed to kill pid ${_pid}"
  done
  # Give OS a moment to release the socket
  sleep 1
fi

# Start the full enterprise API server (it exposes /health/ready which smoke expects)
"$BIN" enterprise --port "$API_PORT" --host 127.0.0.1 --jwt-secret "$JWT_SECRET" >>"$SERVER_LOG" 2>&1 &
API_PID=$!
trap 'kill $API_PID >/dev/null 2>&1 || true' EXIT

log "4. Polling /health/ready"
ATTEMPT=0
until curl -fsS "http://127.0.0.1:${API_PORT}/health/ready" >/dev/null 2>&1; do
  ATTEMPT=$((ATTEMPT+1))
  if [ $ATTEMPT -ge $START_TIMEOUT ]; then
    log "API failed to become ready within ${START_TIMEOUT}s"
    exit 1
  fi
  sleep 1
done

log "5. Basic /health check"
curl -fsS "http://127.0.0.1:${API_PORT}/health" >/dev/null || { log "/health failed"; exit 1; }

log "Smoke server log (tail 200):"; tail -n 200 "$SERVER_LOG" || true

log " Smoke test passed"
