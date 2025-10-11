package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string, ext string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sample"+ext)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

func TestOptionsValidate_NegativeValues(t *testing.T) {
	opts := Options{Provider: "", File: ""}
	if err := opts.Validate(); err == nil {
		t.Fatalf("expected provider+file error")
	}
	opts = Options{Provider: "aws", File: "file", Tolerance: -1}
	if err := opts.Validate(); err == nil {
		t.Fatalf("expected tolerance error")
	}
	opts = Options{Provider: "aws", File: "file", ErrorThreshold: -2}
	if err := opts.Validate(); err == nil {
		t.Fatalf("expected threshold error")
	}
}

func TestProcess_ParseErrorsAndFormatDetection(t *testing.T) {
	// missing provider+file triggers parse stage error
	sum, err := Process(Options{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sum.ExitCode != ExitParseError {
		t.Fatalf("expected parse exit, got %d", sum.ExitCode)
	}
	// create csv file and run minimal parse
	csv := writeTempFile(t, "a,b,c\n1,2,3\n4,5,6\n", ".csv")
	sum2, err := Process(Options{Provider: "aws", File: csv, Limit: 2})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if sum2.Format != "csv" {
		t.Fatalf("expected csv format detection, got %s", sum2.Format)
	}
	if pr := sum2.Stages[StageParse]; pr.ProcessedRows == 0 {
		t.Fatalf("expected >0 processed rows")
	}
}

func TestProcess_StopAfterMapSkipsLater(t *testing.T) {
	csv := writeTempFile(t, "a\n1\n2\n3\n", ".csv")
	sum, _ := Process(Options{Provider: "aws", File: csv, StopAfter: string(StageMap)})
	if sum.Stages[StageMap].Status == StatusSkipped {
		t.Fatalf("map should run")
	}
	if sum.Stages[StageValidate].Status != StatusSkipped {
		t.Fatalf("validate should skip when stopAfter=map")
	}
}
