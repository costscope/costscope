# CLI Command Drift Guard

The CI job "CLI Command Drift Guard" ensures that generated Cobra command builders are in sync with the checked-in code. If you modify command specs, flags, or related generation inputs, run the drift guard locally before pushing:

## How to run locally

1. Generate or validate builders:

   ```bash
   bash scripts/ci/cli-drift-check.sh
   ```

2. Verify the working tree is clean:

   ```bash
   git status --porcelain
   ```

- If nothing is listed, you’re good to push.
- If files like `zz_generated_*` changed, add them to your commit:

```bash
git add .
git commit -m "chore(cli): update generated command builders"
```

## Troubleshooting CI failures

- If the CI job still reports drift after you pushed, download the artifact diff from the job and compare it against your local generation.
- Re-run the job on GitHub using “Re-run failed jobs” if the prior run was for an older SHA (this can surface as a false positive).
- Align the spec or commit regenerated files to resolve.

## Notes

- Keep generator/tool versions stable to avoid non-deterministic output.
- Prefer small, isolated changes to command specs so drift is easy to review.
