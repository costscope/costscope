package exporters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"

	"gopkg.in/yaml.v3"
)

// YAMLExporter implements exporting reports to YAML format
type YAMLExporter struct{}

// NewYAMLExporter creates a new YAML exporter
func NewYAMLExporter() *YAMLExporter {
	return &YAMLExporter{}
}

// Export exports a report to YAML format
func (e *YAMLExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatYAML {
		return 0, "", fmt.Errorf("unsupported format for YAML exporter: %s", format)
	}

	// Marshal report to YAML
	data, err := yaml.Marshal(report)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal report to YAML: %w", err)
	}

	// Resolve destination and write via ObjectStore (supports fs, file://, s3://, gs://)
	dest, err := outputpath.ResolveOutputPath("", output, types.ExportFormatYAML)
	if err != nil {
		return 0, "", err
	}
	store, _, err := NewObjectStore(ctx, dest)
	if err != nil {
		return 0, "", err
	}
	if err := store.Put(ctx, dest, bytes.NewReader(data)); err != nil {
		return 0, "", fmt.Errorf("failed to store YAML output: %w", err)
	}
	sum := sha256.Sum256(data)
	return int64(len(data)), hex.EncodeToString(sum[:]), nil
}

// GetSupportedFormats returns the formats supported by this exporter
func (e *YAMLExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatYAML}
}
