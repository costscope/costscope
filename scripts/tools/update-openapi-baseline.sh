#!/usr/bin/env bash
set -euo pipefail
# update-openapi-baseline.sh
# Regenerates baseline OpenAPI JSON snapshots from current YAML specs.
# Prefers local yq; falls back to running inside the CI contract image via docker.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

GEN_PUBLIC="$ROOT_DIR/internal/api/docs/openapi.yaml"
GEN_ENT="$ROOT_DIR/internal/api/docs/enterprise-openapi.yaml"
OUT_PUBLIC="$ROOT_DIR/api/baseline/openapi.v1.json"
OUT_ENT="$ROOT_DIR/api/baseline/openapi.enterprise.v1.json"

convert() {
  local in="$1" out="$2"
  if [ ! -f "$in" ]; then
    echo "[baseline] skip: input not found $in" >&2
    return 0
  fi
  mkdir -p "$(dirname "$out")"
  if command -v yq >/dev/null 2>&1; then
    yq -o=json eval '.' "$in" > "$out"
  else
    # fallback via docker with CI contract image that contains yq
    local img="ghcr.io/costscope/ci-contract:dev-local"
    if ! docker image inspect "$img" >/dev/null 2>&1; then
      img="ghcr.io/costscope/ci-contract:latest"
    fi
    docker run --rm -v "$ROOT_DIR":"/work" "$img" \
      bash -lc "yq -o=json eval '.' \"${in//"$ROOT_DIR"/\/work}\" > \"${out//"$ROOT_DIR"/\/work}\""
  fi
  echo "[baseline] wrote $(realpath "$out" 2>/dev/null || echo "$out")"
}

convert "$GEN_PUBLIC" "$OUT_PUBLIC"
convert "$GEN_ENT" "$OUT_ENT"
echo "[baseline] done"
