````markdown
## Batch Conversion (CSV → FOCUS / DuckDB)

Use this command to convert a directory of CSV files into CostScope FOCUS files (or into a consolidated DuckDB file). It supports parallel workers, dry-run and idempotent output paths.

Basic usage:
```bash
# Convert all CSVs in ./input to ./out, 4 workers (default)
./bin/costscope convert batch --input-dir ./input --output-dir ./out --workers 4
```

Options:

- `--input-dir`    Path to directory containing input CSV files
- `--output-dir`   Path where converted files will be written
- `--format`       Output format: `focus` | `duckdb` (default: `focus`)
- `--workers`      Parallel conversion workers (default: 2)
- `--dry-run`      Validate inputs and show actions without writing outputs
- `--yes`          Non-interactive: overwrite outputs without prompting

Examples:

```bash
# Dry-run to preview actions
./bin/costscope convert batch --input-dir ./csv --output-dir ./out --dry-run

# Convert and write a consolidated DuckDB file
./bin/costscope convert batch --input-dir ./csv --format duckdb --output-dir ./out --workers 8
```

Notes:

- For large datasets prefer increasing `--workers` and ensure sufficient memory.
- Use `--dry-run` in CI to validate inputs before running real conversion.

````
