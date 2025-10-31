#!/usr/bin/env bash
# Ensure essential tooling (node, npm, make, git) exists when running under act-like conditions
# Safe to run multiple times; installs only if missing. Designed for container jobs.
set -euo pipefail

# Detect act context via missing ACTIONS_RUNTIME_TOKEN or explicit flags
IS_ACT_SHIM="false"
if [[ -z "${ACTIONS_RUNTIME_TOKEN:-}" || "${GITHUB_ACTOR:-}" == "nektos/act" || "${IS_ACT:-false}" == "true" ]]; then
  IS_ACT_SHIM="true"
fi

if [[ "$IS_ACT_SHIM" != "true" ]]; then
  echo "[ensure-act-tooling] Not running under act; nothing to do"
  exit 0
fi

need_update="false"
install_pkgs=()

have() { command -v "$1" >/dev/null 2>&1; }

if have node; then
  echo "[ensure-act-tooling] node present: $(node --version)"
else
  install_pkgs+=(ca-certificates curl gnupg)
  echo "[ensure-act-tooling] node missing; will install"
fi

if have npm; then
  echo "[ensure-act-tooling] npm present: $(npm --version)"
else
  # npm comes with nodejs package in Debian/Ubuntu
  :
fi

if have make; then
  echo "[ensure-act-tooling] make present: $(make --version | head -n1)"
else
  install_pkgs+=(make)
  echo "[ensure-act-tooling] make missing; will install"
fi

if have git; then
  echo "[ensure-act-tooling] git present: $(git --version)"
else
  install_pkgs+=(git)
  echo "[ensure-act-tooling] git missing; will install"
fi

if (( ${#install_pkgs[@]} > 0 )) || ! have node; then
  export DEBIAN_FRONTEND=noninteractive
  echo "[ensure-act-tooling] Updating apt indexes"
  apt-get update
  # Install Node.js 20 from NodeSource if node is missing
  if ! have node; then
    echo "[ensure-act-tooling] Installing Node.js LTS (20.x)"
    if curl -fsSL https://deb.nodesource.com/setup_20.x -o /tmp/nodesetup.sh; then
      bash /tmp/nodesetup.sh
      apt-get install -y --no-install-recommends nodejs "${install_pkgs[@]}"
    else
      # Fallback to distro nodejs/npm
      apt-get install -y --no-install-recommends nodejs npm "${install_pkgs[@]}"
    fi
  else
    # Only other packages needed
    if (( ${#install_pkgs[@]} > 0 )); then
      apt-get install -y --no-install-recommends "${install_pkgs[@]}"
    fi
  fi
fi

# Final versions
node -v || true
npm -v || true
make --version | head -n1 || true
git --version || true

echo "[ensure-act-tooling] Completed"
