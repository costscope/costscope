#!/usr/bin/env bash
# Build a distroless variant by extracting a binary from an existing image
# Usage: build-distroless-from-image.sh <source_image> <target_image>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

SRC_IMAGE="${1:-}"
DST_IMAGE="${2:-}"

if [[ -z "${SRC_IMAGE}" || -z "${DST_IMAGE}" ]]; then
  ci::die "usage: $0 <source_image> <target_image>"
fi

ci::require_cmd docker

workdir="$(mktemp -d)"
container_id=""
cleanup() {
  local ec=$?
  if [[ -n "$container_id" ]]; then
    docker rm -f "$container_id" >/dev/null 2>&1 || true
  fi
  rm -rf "$workdir" || true
  exit $ec
}
trap cleanup EXIT INT TERM

ci::log "Extracting binary from $SRC_IMAGE"
container_id=$(docker create "$SRC_IMAGE")
mkdir -p "$workdir/extracted"
docker cp "$container_id:/app/costscope" "$workdir/extracted/costscope"

ci::log "Building distroless image $DST_IMAGE"
cat > "$workdir/Dockerfile.distroless" <<'EOF'
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY costscope .
EXPOSE 8080 8443
USER nonroot
CMD ["./costscope", "api", "enterprise", "--host", "0.0.0.0", "--port", "8080"]
EOF

docker build -t "$DST_IMAGE" -f "$workdir/Dockerfile.distroless" "$workdir/extracted"
if ci::is_act; then
  ci::log "[act] Skipping push for $DST_IMAGE"
else
  ci::log "Pushing $DST_IMAGE"
  docker push "$DST_IMAGE"
fi
