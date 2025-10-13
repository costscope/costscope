package exporters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"
)

type pdfStubRenderer struct{}

func (r *pdfStubRenderer) Render(ctx context.Context, report interface{}) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	content := fmt.Sprintf("%%PDF-1.4\n%% CostScope PDF stub\nReportGenerated: %s\nReportType: %T\n%%EOF\n", now, report)
	return []byte(content), nil
}
func (r *pdfStubRenderer) ContentType() string { return "application/pdf" }

type PDFExporter struct{ Renderer types.RenderReport }

func NewPDFExporter() *PDFExporter { return &PDFExporter{Renderer: &pdfStubRenderer{}} }

func (e *PDFExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatPDF {
		return 0, "", fmt.Errorf("unsupported format for PDF exporter: %s", format)
	}
	if e.Renderer == nil {
		return 0, "", errors.New("pdf exporter renderer not configured")
	}
	data, err := e.Renderer.Render(ctx, report)
	if err != nil {
		return 0, "", err
	}
	out, err := outputpath.ResolveOutputPath("", output, types.ExportFormatPDF)
	if err != nil {
		return 0, "", err
	}
	store, _, err := NewObjectStore(ctx, out)
	if err != nil {
		return 0, "", err
	}
	if err := store.Put(ctx, out, bytes.NewReader(data)); err != nil {
		return 0, "", err
	}
	sum := sha256.Sum256(data)
	return int64(len(data)), hex.EncodeToString(sum[:]), nil
}
func (e *PDFExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatPDF}
}
