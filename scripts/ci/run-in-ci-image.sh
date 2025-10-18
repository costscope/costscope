#!/usr/bin/env bash
# Run a command inside the CI tools image with repo workspace mounted
# Usage:
#   scripts/ci/run-in-ci-image.sh [--with-docker-sock] [--workdir <dir>] -- <command>
# Examples:
#   scripts/ci/run-in-ci-image.sh -- make sbom
#   scripts/ci/run-in-ci-image.sh -- bash ./scripts/security/run-gosec.sh
#   scripts/ci/run-in-ci-image.sh --with-docker-sock -- bash ./scripts/security/run-trivy-image.sh

set -euo pipefail

# shellcheck source=lib/common.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

mount_docker_sock="false"
workdir="/work"

args=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --with-docker-sock)
      mount_docker_sock="true"; shift ;;
    --workdir)
      workdir="$2"; shift 2 ;;
    --)
      shift
      break ;;
    *)
      break ;;
  esac
done

if [[ $# -eq 0 ]]; then
  ci::die "No command provided. Usage: $0 [--with-docker-sock] [--workdir <dir>] -- <command>"
fi

image="${CI_TOOLS_IMAGE:-}"
if [[ -z "$image" ]]; then
  ci::die "CI_TOOLS_IMAGE env var is required"
fi

# Default host path for workspace is current repository root
host_root="$(ci::repo_root "${GITHUB_WORKSPACE:-}")"
if [[ ! -d "$host_root" ]]; then
  host_root="$PWD"
fi

docker_args=(
  run --rm
  -v "${host_root}:${workdir}"
  -w "${workdir}"
)

if [[ "$mount_docker_sock" == "true" ]]; then
  if [[ -S /var/run/docker.sock ]]; then
    docker_args+=( -v /var/run/docker.sock:/var/run/docker.sock )
  else
    ci::warn "--with-docker-sock set but /var/run/docker.sock not found on host"
  fi
fi

ci::log "Running in CI image: ${image} :: $*"
# Forward the command string as-is to bash -lc; use printf %q for logging if needed
exec docker "${docker_args[@]}" "$image" bash -lc "$*"
