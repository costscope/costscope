---
title: CI Object Storage Smoke Tests
description: Optional S3/GCS smoke tests configuration and GitHub Actions examples.
---
## Purpose
Early detection of breaking changes in container image (binary startup, health, metrics) and optional object storage integration (S3 / GCS) before full test matrix runs.

## Minimal Smoke Script
```bash
set -euo pipefail
docker run -d --rm --name costscope-smoke -p 18080:8080 costscope:PR
sleep 3
curl -fsS localhost:18080/health/ready >/dev/null
curl -fsS localhost:18080/metrics | grep -q costscope_http_requests_total
echo "Smoke OK"
```

## Optional S3 Credentials
Set ephemeral IAM user credentials (read-only bucket listing) via GitHub Actions secrets:
| Secret | Purpose |
|--------|---------|
| `AWS_ACCESS_KEY_ID` | Access key |
| `AWS_SECRET_ACCESS_KEY` | Secret key |
| `AWS_REGION` | Region (e.g. us-east-1) |

Within smoke job you can test provider init:
```bash
costscope providers info --format json | jq '.providers[] | select(.name=="aws")'
```

## GitHub Actions Example
```yaml
name: smoke
on: [pull_request]
jobs:
	container-smoke:
		runs-on: ubuntu-latest
		steps:
			- uses: actions/checkout@v4
			- name: Build image
				run: docker build -t costscope:PR .
			- name: Run smoke
				run: |
					set -euo pipefail
					docker run -d --rm --name costscope-smoke -p 18080:8080 costscope:PR
					for i in {1..15}; do curl -fsS localhost:18080/health/ready && break || sleep 1; done
					curl -fsS localhost:18080/metrics | grep costscope_http_request_duration_seconds
			- name: AWS provider check (optional)
				if: env.AWS_ACCESS_KEY_ID != ''
				env:
					AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
					AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
					AWS_REGION: ${{ secrets.AWS_REGION }}
				run: |
					docker exec costscope-smoke costscope providers list --format json | jq '.providers | length'
```

## Metrics Assertions
Prefer presence checks for stable metric names; avoid brittle exact counters early in lifecycle. Example:
```bash
curl -s localhost:18080/metrics | grep -E '^costscope_(http|mapper|rbac)'
```

## Failure Diagnostics
On failure capture logs:
```bash
docker logs costscope-smoke | tail -n 200
```
If container fails to start, build again with `--progress=plain` to surface build-step issues.

## Extensions
- Add TLS variant (self-signed cert) to detect scheme regression
- Include multi-arch (amd64 + arm64) build check using `buildx`
- Publish lightweight SBOM diff (optional)

Keep updated when health endpoints or base image strategy changes.

### Ensuring a clean working tree (recommended)

The release `smoke.sh` script requires a clean git working tree to avoid accidental inclusion of generated or uncommitted files in release artifacts. Prefer these simple local workflows to prepare a clean tree before running release tasks or smoke tests:

- Quick check:

```bash
git status --porcelain
```

- Stash local changes (temporary, easy to restore later):

```bash
git stash push -m "smoke-test-temp"
# run smoke/build
git stash pop || true
```

- Commit changes you intend to keep:

```bash
git add <files>
git commit -m "chore: save WIP changes before smoke"
# run smoke/build
```

- Use a temporary branch to test without touching main branches:

```bash
git checkout -b tmp/smoke-debug
git reset --hard HEAD
# run smoke/build
git checkout -
git branch -D tmp/smoke-debug
```

Do NOT attempt to bypass the clean-repo check in CI; CI enforces this to ensure reproducible, auditable releases.
