# Examples – how to run

Example programs are excluded from normal builds via Go build tags. To run them locally, include the example tag (and duckdb for DuckDB-backed samples).

## Quick run

- Run a demo directly:
  - `go run -tags "example duckdb" ./examples/simple-duckdb`
  - `go run -tags "example duckdb" ./examples/query-builder`
  - `go run -tags "example duckdb" ./examples/csv-to-focus`

- Build a demo binary:
  - `go build -tags "example duckdb" -o ./bin/example-csv2focus ./examples/csv-to-focus`

## Notes
- Most examples depend on DuckDB; CGO must be enabled (default on macOS/Linux). If CGO is disabled, re-run with `CGO_ENABLED=1`.
- Examples are intentionally excluded from tests and standard builds; passing the tags opt-in keeps CI clean.
- Some non-DuckDB demos may live under `demo/` and use only the `example` tag; invoke with `-tags "example"` accordingly.

For production features and CLI, use the main `costscope` binary (see project README).
