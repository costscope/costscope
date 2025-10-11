//go:build experimental && !duckdb

package focus

// Stubbed helpers when DuckDB is not included.
// Return nil slices so callers can fallback to baseline-derived averages.
func loadValuesOptional(path string) ([]float64, []float64) { return nil, nil }
