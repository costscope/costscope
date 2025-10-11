#!/usr/bin/env bash
# Build a distroless variant by extracting a binary from an existing image
# Usage: build-distroless-from-image.sh <source_image> <target_image>
set -euo pipefail

SRC_IMAGE="${1:-}"
DST_IMAGE="${2:-}"

if [[ -z "${SRC_IMAGE}" || -z "${DST_IMAGE}" ]]; then
  echo "usage: $0 <source_image> <target_image>" >&2
  exit 2
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

# Extract binary from source image
id=$(docker create "$SRC_IMAGE")
mkdir -p "$workdir/extracted"
docker cp "$id:/app/costscope" "$workdir/extracted/costscope"
docker rm -f "$id" >/dev/null

# Build distroless Dockerfile
cat > "$workdir/Dockerfile.distroless" <<'EOF'
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY costscope .
EXPOSE 8080 8443
USER nonroot
CMD ["./costscope", "api", "enterprise", "--host", "0.0.0.0", "--port", "8080"]
EOF

# Build and optionally push
docker build -t "$DST_IMAGE" -f "$workdir/Dockerfile.distroless" "$workdir/extracted"
if [[ "${GITHUB_ACTOR:-}" != "nektos/act" ]]; then
  docker push "$DST_IMAGE"
else
  echo "[act] Skipping push for $DST_IMAGE"
fi
