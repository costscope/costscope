package security_test

import (
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/security"
)

func TestSecurityScannerBasic(t *testing.T) {
	logger := logging.NewLogger("info")
	s := security.NewScanner(logger)

	res, err := s.ScanVulnerabilities("./")
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if res == nil || res.GeneratedAt.IsZero() {
		t.Fatalf("invalid scan result")
	}
	if res.Summary == "" {
		t.Errorf("summary should not be empty")
	}
}
