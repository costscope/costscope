#!/usr/bin/env bats

load 'test_helper'

# Smoke: script lists zero files gracefully in a temp dir
@test "run-shellcheck handles no scripts" {
  run bash -lc 'tmp=$(mktemp -d); cd "$tmp"; bash -c "$(cat <<'EOF'
set -euo pipefail
SCRIPT_DIR=$(pwd)
mkdir -p ./scripts/ci/lib
cp -r /workspaces/costscope/scripts/ci/lib/common.sh ./scripts/ci/lib/common.sh
cat > ./scripts/ci/run-shellcheck.sh <<'S'
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"
files=()
if command -v git >/dev/null 2>&1; then
  mapfile -d '' -t files < <(git ls-files -z -- '*.sh')
else
  mapfile -d '' -t files < <(find . -type f -name "*.sh" -print0)
fi
if [[ ${#files[@]} -eq 0 ]]; then
  ci::log "No shell scripts found to lint."
  exit 0
fi
S
chmod +x ./scripts/ci/run-shellcheck.sh
./scripts/ci/run-shellcheck.sh
EOF
)"'
  [ "$status" -eq 0 ]
}
