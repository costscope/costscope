package exporters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"
)

// JSONExporter implements exporting reports to JSON format
type JSONExporter struct{}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{}
}

// Export exports a report to JSON format
func (e *JSONExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatJSON {
		return 0, "", fmt.Errorf("unsupported format for JSON exporter: %s", format)
	}

	// Marshal report to JSON with pretty printing
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	// Resolve path (uses env/config + auto-naming rules)
	outputPath, err := outputpath.ResolveOutputPath("", output, types.ExportFormatJSON)
	if err != nil {
		return 0, "", err
	}
	store, _, err := NewObjectStore(ctx, outputPath)
	if err != nil {
		return 0, "", err
	}
	if err := store.Put(ctx, outputPath, bytes.NewReader(data)); err != nil {
		return 0, "", fmt.Errorf("failed to store JSON output: %w", err)
	}
	sum := sha256.Sum256(data)
	return int64(len(data)), hex.EncodeToString(sum[:]), nil
}

// GetSupportedFormats returns the formats supported by this exporter
func (e *JSONExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatJSON}
}
