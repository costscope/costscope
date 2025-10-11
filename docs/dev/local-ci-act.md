---
title: Local CI with act
description: Running GitHub Actions locally using act for faster feedback.
---
## Goal
Replicate GitHub Actions workflow locally for faster iteration on CI issues (matrix failures, missing targets, permission mistakes).

## Installation
```bash
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | bash
```

## Usage
List jobs:
```bash
act -l
```
Run default (pull_request) with medium image:
```bash
act pull_request -P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-latest
```
Run specific job:
```bash
act -j quality
```

### Fix noisy sudo hostname warning

If you see repeated warnings like "sudo: unable to resolve host mic: Name or service not known" from containers started by `act`, start `act` with a Docker run arg to add the hosts entry, e.g.:

```bash
act -j quality -- --add-host=mic:127.0.0.1
```

Or set an environment variable used by your shell wrapper:

```bash
export ACT_RUN_ARGS="--add-host=mic:127.0.0.1"
act -j quality $ACT_RUN_ARGS
```

This ensures the container hostname `mic` resolves and silences the sudo warning in logs.

Recommended wrapper
-------------------

We've added a small wrapper script `scripts/act-run.sh` which automatically injects the required `--add-host` argument if it's not present. Prefer running `act` through the wrapper when working in this repo:

```bash
./scripts/act-run.sh -j quality
```

The wrapper is idempotent: if you already pass `--add-host=mic:127.0.0.1` it won't add a duplicate entry.


## Environment & Secrets
Provide required secrets via `.secrets-s KEY=VALUE`. Mask production credentials; use scoped test tokens.

```bash
file or inline
```

## Performance Tips
- Cache Go modules: mount `~/go/pkg/mod` into container
- Use smaller base images for non-docker build jobs

## Limitations
- Service containers networking differences (ports) vs GitHub-hosted
- Some Actions using service principals / OIDC may need stubs

## Debugging
Add `--verbosedocker psdocker logs`.

```bash
or inspect job container logs using
```

```bash
+
```

Update when workflow structure changes.
