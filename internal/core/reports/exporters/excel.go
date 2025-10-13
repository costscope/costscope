package exporters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"
)

type excelStubRenderer struct{}

func (r *excelStubRenderer) Render(ctx context.Context, report interface{}) ([]byte, error) {
	b, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}
	return []byte("XLSX-STUB\n" + string(b)), nil
}
func (r *excelStubRenderer) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

type ExcelExporter struct{ Renderer types.RenderReport }

func NewExcelExporter() *ExcelExporter { return &ExcelExporter{Renderer: &excelStubRenderer{}} }

func (e *ExcelExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatExcel {
		return 0, "", fmt.Errorf("unsupported format for Excel exporter: %s", format)
	}
	if e.Renderer == nil {
		return 0, "", errors.New("excel exporter renderer not configured")
	}
	data, err := e.Renderer.Render(ctx, report)
	if err != nil {
		return 0, "", err
	}
	out, err := outputpath.ResolveOutputPath("", output, types.ExportFormatExcel)
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
func (e *ExcelExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatExcel}
}
