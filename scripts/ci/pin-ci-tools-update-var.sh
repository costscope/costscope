#!/usr/bin/env bash
set -euo pipefail

# pin-ci-tools-update-var.sh
# Creates or updates the repository Actions variable CI_TOOLS_IMAGE via GitHub API
# Inputs (env):
#   OWNER        - repository owner (required)
#   REPO         - repository name (required)
#   GH_TOKEN     - token to call GitHub API (required)
#   IMAGE        - image reference to pin (required)
#   GH_API       - API base (optional; default https://api.github.com)

owner="${OWNER:-}"
repo="${REPO:-}"
token="${GH_TOKEN:-}"
image="${IMAGE:-}"
api_base="${GH_API:-https://api.github.com}"

if [[ -z "$owner" || -z "$repo" || -z "$token" || -z "$image" ]]; then
  echo "OWNER, REPO, GH_TOKEN, and IMAGE are required" >&2
  exit 2
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
status=$(curl -sS -o /tmp/resp.json -w "%{http_code}" -X PATCH \
  "${api_base}/repos/${owner}/${repo}/actions/variables/${name}" \
  "${headers[@]}" \
  -d "$update_body") || true

if [[ "$status" == "404" ]]; then
  status=$(curl -sS -o /tmp/resp.json -w "%{http_code}" -X POST \
    "${api_base}/repos/${owner}/${repo}/actions/variables" \
    "${headers[@]}" \
    -d "$create_body") || true
fi

if [[ "$status" =~ ^2[0-9]{2}$ ]]; then
  echo "Pinned CI_TOOLS_IMAGE to ${image}"
else
  echo "API call failed with status ${status}:" >&2
  cat /tmp/resp.json >&2 || true
  exit 4
fi
