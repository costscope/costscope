package main

import (
	"github.com/costscope/costscope/scripts/tools/internal/gcshared"
)

func main() {
	// Name distinguishes this entrypoint; defaults are fine for the smoke test.
	gcshared.Run("gc-test", nil)
}
