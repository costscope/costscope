#!/usr/bin/env bash
set -euo pipefail

# Container smoke test script for CostScope
# Requirements:
#  - env IMAGE_STD (standard image, e.g. ghcr.io/org/costscope:sha)
#  - env IMAGE_DISTROLESS (distroless/secure image variant) (optional but recommended)
# Behaviour:
#  * For each provided image: run detached, wait for /health/live (HTTP) up to 40s
#  * Verify /metrics exposes at least one costscope_ metric
#  * Capture container logs into smoke-logs/<name>.log
#  * Optionally exercise TLS mode: generate self‑signed cert, start with --tls-enabled and probe HTTPS endpoints (skipped for distroless if openssl absent)
#  * Always cleans up containers
#  * Exits non‑zero on any failure

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${ROOT_DIR}/smoke-logs"
mkdir -p "${LOG_DIR}" || true

HTTP_TIMEOUT=40   # seconds total for health retry
SLEEP_INTERVAL=2  # seconds between retries

have_cmd() { command -v "$1" >/dev/null 2>&1; }

fail() { echo "[FAIL] $*" >&2; exit 1; }

log() { echo "[INFO] $*"; }

probe_http() {
  local url=$1; shift
  curl -sk --fail --max-time 5 "$url" "$@" >/dev/null
}

wait_health() {
  local base=$1; local proto=$2; local elapsed=0
  while (( elapsed < HTTP_TIMEOUT )); do
    if probe_http "${proto}://${base}/health/live"; then
      log "Health live OK (${proto}://${base}) after ${elapsed}s"; return 0; fi
    sleep "$SLEEP_INTERVAL"; elapsed=$(( elapsed + SLEEP_INTERVAL ))
  done
  return 1
}

check_metrics() {
  local base=$1; local proto=$2
  local body
  body=$(curl -sk --fail --max-time 5 "${proto}://${base}/metrics" || true)
  echo "$body" | grep -E '^costscope_' >/dev/null || {
    echo "Metrics body (truncated):" >&2
    echo "${body}" | head -n 40 >&2
    return 1
  }
}

run_container() {
  local image=$1; local name=$2; shift 2
  log "Starting container $name from $image ($*)"
  docker run -d --rm -p 8080:8080 -p 8443:8443 --name "$name" "$image" "$@" >/dev/null
}

collect_logs() {
  local name=$1
  docker logs "$name" >& "${LOG_DIR}/${name}.log" || true
}

smoke_http() {
  local image=$1; local variant=$2
  local name="smoke-${variant}-http"
  run_container "$image" "$name" api enterprise --host 0.0.0.0 --port 8080
  trap 'docker rm -f "$name" >/dev/null 2>&1 || true' RETURN
  if ! wait_health "127.0.0.1:8080" http; then collect_logs "$name"; fail "Health check failed (HTTP) for $image"; fi
  if ! check_metrics "127.0.0.1:8080" http; then collect_logs "$name"; fail "Metrics check failed (HTTP) for $image"; fi
  collect_logs "$name"
  docker rm -f "$name" >/dev/null 2>&1 || true
  trap - RETURN
  log "HTTP smoke passed for $image"
}

smoke_tls() {
  local image=$1; local variant=$2
  if ! have_cmd openssl; then
    log "openssl not present; skipping TLS smoke for $image"; return 0; fi
  local tmpd; tmpd=$(mktemp -d)
  openssl req -x509 -nodes -newkey rsa:2048 -subj "/CN=localhost" -days 1 -keyout "$tmpd/key.pem" -out "$tmpd/cert.pem" >/dev/null 2>&1
  local name="smoke-${variant}-tls"
  log "Starting TLS container $name from $image"
  docker run -d --rm -p 8443:8443 --name "$name" -v "$tmpd:/certs:ro" \
    "$image" api enterprise --host 0.0.0.0 --port 8443 --tls-enabled --tls-cert /certs/cert.pem --tls-key /certs/key.pem >/dev/null
  trap 'docker rm -f "$name" >/dev/null 2>&1 || true; rm -rf "$tmpd"' RETURN
  if ! wait_health "127.0.0.1:8443" https; then collect_logs "$name"; fail "Health check failed (TLS) for $image"; fi
  if ! check_metrics "127.0.0.1:8443" https; then collect_logs "$name"; fail "Metrics check failed (TLS) for $image"; fi
  collect_logs "$name"
  docker rm -f "$name" >/dev/null 2>&1 || true
  rm -rf "$tmpd" || true
  trap - RETURN
  log "TLS smoke passed for $image"
}

main() {
  local failures=0
  if [ -z "${IMAGE_STD:-}" ] && [ -z "${IMAGE_DISTROLESS:-}" ]; then
    fail "IMAGE_STD or IMAGE_DISTROLESS must be set"
  fi
  if [ -n "${IMAGE_STD:-}" ]; then
    smoke_http "$IMAGE_STD" std || failures=$((failures+1))
    smoke_tls "$IMAGE_STD" std || failures=$((failures+1))
  fi
  if [ -n "${IMAGE_DISTROLESS:-}" ]; then
    smoke_http "$IMAGE_DISTROLESS" distroless || failures=$((failures+1))
    smoke_tls "$IMAGE_DISTROLESS" distroless || failures=$((failures+1))
  fi
  if (( failures > 0 )); then
    fail "Smoke tests failed ($failures failures)"
  fi
  log "All container smoke tests passed"
}

main "$@"
