---
title: Verification (Post-Install)
description: Quick verification steps to run after installing a CostScope binary or deploying a release.
---

## Verification (Post-Install)

Use these quick checks after installing the binary to verify a working installation and basic runtime sanity.

### 1) Binary & version
```bash
./costscope --version
# expect: CostScope vX.Y.Z (commit: <sha>, built with go1.24.x)
```

### 2) Health endpoint (if API started)
```bash
export COSTSCOPE_JWT_SECRET="$(openssl rand -base64 48)"
./costscope api enterprise --port 8080 &

curl -sS http://localhost:8080/health/ready
curl -sS http://localhost:8080/metrics | grep -E '^costscope_' | head -n 20
```

### 3) Quick convert + validate round-trip
```bash
# convert (streaming, invariants)
./costscope focus convert --provider aws --input ./aws-cur.csv.gz \
  --output ./out.focus.parquet --streaming --invariants --invariants-report ./inv.json --quiet

# validate produced parquet and invariants
./costscope validate ./out.focus.parquet --all --output validation.json
jq . ./inv.json
```

### 4) Invariants smoke-check (baseline diff)
```bash
./costscope invariants regenerate ./out.focus.parquet --output invariants_current.json --tolerance 0.01
./costscope invariants diff invariants_current.json --baseline tests/fixtures/quality/baseline_synth_invariants.json --tolerance 0.01 --report invariants_diff.json

# Check for violations
jq '.violations | length' invariants_diff.json
```

Notes:
- The invariants diff should produce zero violations for a matching baseline; non-zero violations indicate drift and require investigation.
- For production: always produce SBOM and checksums locally and attach them to release artifacts.

### SBOM & checksums (short)
```bash
# Generate SBOM and checksums locally before publishing
make sbom       # produces sbom-syft.json / sbom.json
make checksums  # populates checksums.txt
```

Verify produced artifacts (basic):
```bash
[ -f sbom-syft.json ] && jq '.components | length' sbom-syft.json || echo "sbom missing or empty"
[ -s checksums.txt ] && echo "checksums present" || echo "checksums missing or empty"
```

### Placeholder assets / images
If README or docs show placeholder images, add PNG/SVG files to `docs/assets/docs/assets/placeholder.txt` guidance for filenames and sizes. After adding images, the docs link-checker in CI will validate local references.

```bash
and follow
```

---

For deeper release verification (reproducible builds, cosign verification, supply-chain gates), see `docs/security/supply-chain.mddocs/security/reproducible-builds.md`.

```bash
and
```
