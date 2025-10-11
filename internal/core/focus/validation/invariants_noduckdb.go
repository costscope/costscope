//go:build !duckdb

package validation

import (
	"fmt"
	"path/filepath"
	"strings"

	"local/costscope/internal/core/focus/quality"
)

// ComputeInvariantsFromFile stub when DuckDB is not linked. This preserves functionality
// of core validation while making invariants optional in no-duckdb builds.
func ComputeInvariantsFromFile(path string) (quality.InvariantMetrics, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".parquet", ".csv", ".json":
		return quality.InvariantMetrics{}, fmt.Errorf("invariants require duckdb build tag in this binary")
	default:
		return quality.InvariantMetrics{}, fmt.Errorf("unsupported invariants input format: %s", ext)
	}
}
