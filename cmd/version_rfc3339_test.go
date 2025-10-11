package cmd

import (
	"testing"
	"time"
)

// Ensures that when BuildDate is set (not the placeholder) it conforms to RFC3339.
// This guards against accidental format regressions in reproducible builds.
func TestBuildDateRFC3339Conditional(t *testing.T) {
	vi := buildVersionInfo()
	if vi.BuildDate == "" {
		t.Fatalf("BuildDate empty")
	}
	if vi.BuildDate == "unknown" { // dev placeholder accepted
		return
	}
	if _, err := time.Parse(time.RFC3339, vi.BuildDate); err != nil {
		t.Fatalf("BuildDate not RFC3339: %s", vi.BuildDate)
	}
}
