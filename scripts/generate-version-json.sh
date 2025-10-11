#!/usr/bin/env bash
set -euo pipefail

# Generate a version.json file at the repository root.
# Intended for release pipelines or local reproducible builds.

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_FILE="$ROOT_DIR/version.json"

: "${SOURCE_DATE_EPOCH:=}"

git_describe() {
  git describe --tags --always --dirty 2>/dev/null || echo "dev"
}

git_commit() {
  git rev-parse --short=12 HEAD 2>/dev/null || echo "none"
}

build_date() {
  if [ -n "$SOURCE_DATE_EPOCH" ]; then
    # prefer GNU date but fall back to python for portability
    date -u -d "@$SOURCE_DATE_EPOCH" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || \
      python -c "from datetime import datetime,os; print(datetime.utcfromtimestamp(int(os.environ.get('SOURCE_DATE_EPOCH'))).strftime('%Y-%m-%dT%H:%M:%SZ'))"
  else
    date -u +%Y-%m-%dT%H:%M:%SZ
  fi
}

go_version() {
  command -v go >/dev/null 2>&1 && go version | awk '{print $3}' || echo "unknown"
}

VERSION=$(git_describe)
COMMIT=$(git_commit)
BUILD_DATE=$(build_date)
GOVERSION=$(go_version)

cat > "$OUT_FILE" <<EOF
{"version":"${VERSION}","commit":"${COMMIT}","build_date":"${BUILD_DATE}","go_version":"${GOVERSION}"}
EOF

echo "Wrote $OUT_FILE"
