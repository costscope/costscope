package api

import (
	"github.com/costscope/costscope/internal/core/logging"
)

// testLogger returns a lightweight logger for use in unit tests.
func testLogger() *logging.Logger {
	return logging.NewLogger(logging.LevelInfo)
}
