This file documents the flag snapshot used by `flag_snapshot_test.go`.

How to regenerate:

- Run the tests with the environment variable to update the on-disk snapshot:

```bash
UPDATE_SNAPSHOT=1 go test ./cmd/modules/multicloud/commands -run TestMulticloudCommands_FlagSnapshot -v
```

- After updating the file on disk, refresh any embedded copies (if using //go:embed) by updating the embedded data and committing both files.

The committed `testdata/command_flags_snapshot.json` serves as the accepted baseline for CI and local runs.
