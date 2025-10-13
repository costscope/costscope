package security

import (
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestScanner_ScanVulnerabilities(t *testing.T) {
	s := NewScanner(logging.NewLogger(logging.LevelWarn))
	res, err := s.ScanVulnerabilities("dummy-target")
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.Summary == "" {
		t.Fatalf("expected summary")
	}
	if len(res.Vulnerabilities) == 0 {
		t.Fatalf("expected at least one vulnerability")
	}
}
