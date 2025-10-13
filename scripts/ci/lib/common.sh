#!/usr/bin/env bash
# Common CI helpers for CostScope
# shellcheck shell=bash
set -euo pipefail

# Logging helpers
ci::ts() { date +"%Y-%m-%dT%H:%M:%S%z"; }
ci::log() { echo "[$(ci::ts)] [ci] $*"; }
ci::warn() { echo "[$(ci::ts)] [ci][warn] $*" >&2; }
ci::die() { echo "[$(ci::ts)] [ci][error] $*" >&2; exit 1; }
ci::debug() {
  if [[ "${CI_DEBUG:-false}" == "true" ]]; then
    echo "[$(ci::ts)] [ci][debug] $*" >&2
  fi
}

# Env detection
ci::is_act() {
  if [[ "${IS_ACT:-}" == "true" || "${ACT:-}" == "true" || "${GITHUB_ACTOR:-}" == "nektos/act" ]]; then
    return 0
  fi
  return 1
}

ci::is_github_actions() {
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then return 0; fi
  return 1
}

# Require a command to exist
ci::require_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || ci::die "required command not found: $cmd"
}

# Resolve repository root directory
# 1) Use git rev-parse --show-toplevel if available
# 2) Fallback to the provided hint (arg1) if it contains go.mod
# 3) Else fallback to current working directory
ci::repo_root() {
  local hint="${1:-}"
  local root=""
  if command -v git >/dev/null 2>&1; then
    if root="$(git rev-parse --show-toplevel 2>/dev/null)"; then
      if [[ -n "$root" && -f "$root/go.mod" ]]; then
        echo "$root"
        return 0
      fi
    fi
  fi

  if [[ -n "$hint" && -d "$hint" && -f "$hint/go.mod" ]]; then
    echo "$hint"
    return 0
  fi

  echo "$PWD"
}
