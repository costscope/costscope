package exporters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"
)

type dummyReport struct {
	Name string `json:"name"`
}

func TestFileURL_JSONExporter(t *testing.T) {
	// use a temp dir
	dir := t.TempDir()
	// construct file:// destination via resolver default naming when explicit base is dir
	// Here we pass an explicit file:// path to ensure file scheme works
	out := "file://" + filepath.ToSlash(filepath.Join(dir, "custom.json"))
	// Write via JSON exporter
	exp := NewJSONExporter()
	report := &dummyReport{Name: "x"}
	ctx := context.Background()
	if _, _, err := exp.Export(ctx, report, types.ExportFormatJSON, out); err != nil {
		t.Fatalf("export json: %v", err)
	}
	// verify file exists without scheme
	p := strings.TrimPrefix(out, "file://")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("stat: %v", err)
	}
}

func TestResolveOutputPath_FileSchemeBase(t *testing.T) {
	dir := t.TempDir()
	base := "file://" + filepath.ToSlash(dir)
	p, err := outputpath.ResolveOutputPath(base, "", types.ExportFormatCSV)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !strings.HasPrefix(p, "file://") {
		t.Fatalf("expected file scheme, got %s", p)
	}
}
