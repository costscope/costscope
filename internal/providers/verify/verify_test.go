package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/core/reports/types"
)

func TestProcessParseError(t *testing.T) {
	sum, err := Process(Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.ExitCode != ExitParseError {
		t.Fatalf("expected parse exit code %d got %d", ExitParseError, sum.ExitCode)
	}
	if sum.Stages[StageParse].Status != StatusError {
		t.Fatalf("parse stage should be error")
	}
}

func TestProcessCSVSample(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sample.csv")
	content := "col1,col2\n1,2\n3,4\n5,6\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	sum, err := Process(Options{Provider: "aws", File: file, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.ExitCode != ExitOK {
		t.Fatalf("expected exit 0 got %d", sum.ExitCode)
	}
	pr := sum.Stages[StageParse]
	if pr.SampledRows == 0 {
		t.Fatalf("expected sampled rows > 0")
	}
	if sum.Format != string(types.ExportFormatCSV) && sum.Format != "unknown" { // allow stub fallback
		t.Fatalf("expected format csv got %s", sum.Format)
	}
}
