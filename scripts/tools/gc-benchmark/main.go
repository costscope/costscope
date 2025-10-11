package main

import (
	"local/costscope/scripts/tools/internal/gcshared"
)

func main() {
	// Explicit name helps identify the tool in output and prevents file-level duplication.
	gcshared.Run("gc-benchmark", []int{50, 100, 200, 300})
}
