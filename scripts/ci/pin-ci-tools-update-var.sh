#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# pin-ci-tools-update-var.sh
# Creates or updates the repository Actions variable CI_TOOLS_IMAGE via GitHub API
# Inputs (env):
#   OWNER        - repository owner (required)
#   REPO         - repository name (required)
#   GH_TOKEN     - token to call GitHub API (required)
#   IMAGE        - image reference to pin (required)
#   GH_API       - API base (optional; default https://api.github.com)

ci::require_cmd curl

owner="${OWNER:-}"
repo="${REPO:-}"
token="${GH_TOKEN:-}"
image="${IMAGE:-}"
api_base="${GH_API:-https://api.github.com}"

# If running under act, skip making API calls to GitHub but succeed for local validation
if ci::is_act; then
  ci::log "Act detected; skipping GitHub API call. Would pin CI_TOOLS_IMAGE to ${image}"
  exit 0
fi

if [[ -z "$owner" || -z "$repo" || -z "$token" || -z "$image" ]]; then
  ci::die "OWNER, REPO, GH_TOKEN, and IMAGE are required"
fi

name="CI_TOOLS_IMAGE"

headers=(
  -H "Accept: application/vnd.github+json"
  -H "Authorization: Bearer ${token}"
  -H "X-GitHub-Api-Version: 2022-11-28"
)

update_body=$(printf '{"value":"%s"}' "${image}")
create_body=$(printf '{"name":"%s","value":"%s"}' "${name}" "${image}")

# Try update (PATCH). If 404, try create (POST).
resp_file=$(mktemp)
trap 'rm -f "$resp_file" 2>/dev/null || true' EXIT INT TERM

status=$(curl -sS -o "$resp_file" -w "%{http_code}" -X PATCH \
  "${api_base}/repos/${owner}/${repo}/actions/variables/${name}" \
  "${headers[@]}" \
  -d "$update_body") || true

if [[ "$status" == "404" ]]; then
  status=$(curl -sS -o "$resp_file" -w "%{http_code}" -X POST \
    "${api_base}/repos/${owner}/${repo}/actions/variables" \
    "${headers[@]}" \
    -d "$create_body") || true
fi

if [[ "$status" =~ ^2[0-9]{2}$ ]]; then
  ci::log "Pinned CI_TOOLS_IMAGE to ${image}"
else
  ci::die "API call failed with status ${status}: $(cat "$resp_file" 2>/dev/null || echo '<no body>')"
fi
