package azure_test

import (
	"bufio"
	"compress/gzip"
	"io"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func writeTempGzipFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path) // #nosec G304 - test temp path under control
	if err != nil {
		t.Fatalf("create gzip file: %v", err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatalf("write gzip content: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}
	return path
}

func slurpAll(t *testing.T, r io.Reader) string {
	t.Helper()
	br := bufio.NewReader(r)
	data, err := io.ReadAll(br)
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	return string(data)
}

func TestOpenAzureInput_NonGzipCSV(t *testing.T) {
	dir := t.TempDir()
	csvContent := "col1,col2\n1,2\n"
	path := writeTempFile(t, dir, "sample.csv", csvContent)

	rc, ext, err := azure.OpenInput(path)
	if err != nil {
		t.Fatalf("openAzureInput error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if ext != ".csv" {
		t.Fatalf("expected .csv ext, got %q", ext)
	}
	if got := slurpAll(t, rc); got != csvContent {
		t.Fatalf("content mismatch: got %q want %q", got, csvContent)
	}
}

func TestOpenAzureInput_GzipCSV(t *testing.T) {
	dir := t.TempDir()
	csvContent := "a,b\n3,4\n"
	path := writeTempGzipFile(t, dir, "gzsample.csv.gz", csvContent)

	rc, ext, err := azure.OpenInput(path)
	if err != nil {
		t.Fatalf("openAzureInput error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if ext != ".csv" {
		t.Fatalf("expected inner .csv ext, got %q", ext)
	}
	if got := slurpAll(t, rc); got != csvContent {
		t.Fatalf("content mismatch: got %q want %q", got, csvContent)
	}
}
