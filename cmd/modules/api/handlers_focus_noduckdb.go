//go:build never

package api

import (
	"fmt"
	"local/costscope/internal/core/focus/quality"
)

// computeAPIInvariantsFromFile is retained only for historical reference and excluded from builds.
func computeAPIInvariantsFromFile(path string) (quality.InvariantMetrics, error) {
	return quality.InvariantMetrics{}, fmt.Errorf("invariants API removed")
}
