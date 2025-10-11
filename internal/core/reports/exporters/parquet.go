package exporters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"local/costscope/internal/core/reports/outputpath"

	focustypes "local/costscope/internal/core/focus/types"
	focuswriters "local/costscope/internal/core/focus/writers"
	"local/costscope/internal/core/reports/types"
)

// ParquetExporter writes generic report structs as Parquet by embedding the JSON payload into a string field of a minimal FocusRecord.
type ParquetExporter struct{}

func NewParquetExporter() *ParquetExporter { return &ParquetExporter{} }

// Export writes the report as Parquet using the established FOCUS Parquet writer for compatibility
func (e *ParquetExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatParquet {
		return 0, "", fmt.Errorf("unsupported format for Parquet exporter: %s", format)
	}

	payload, err := json.Marshal(report)
	if err != nil {
		return 0, "", fmt.Errorf("marshal report: %w", err)
	}

	now := time.Now().UTC()
	rec := focustypes.FocusRecord{
		BillingAccountId:    "reports",
		BillingAccountName:  "reports",
		BillingCurrency:     "N/A",
		BillingPeriodStart:  now,
		BillingPeriodEnd:    now,
		ChargeCategory:      "Report",
		ChargeClass:         "Metadata",
		ChargeDescription:   "Report Export",
		EffectiveCost:       0,
		ListCost:            0,
		UsageQuantity:       1,
		UsageUnit:           "item",
		ProviderName:        "costscope",
		ServiceName:         "reports",
		ResourceId:          "report",
		ResourceName:        string(payload), // store JSON payload
		ConversionTimestamp: now,
	}

	out, err := outputpath.ResolveOutputPath("", output, types.ExportFormatParquet)
	if err != nil {
		return 0, "", err
	}

	tmpPath := filepath.Join(filepath.Dir(out), ".report-export.tmp.parquet")
	opts := &focustypes.ParquetOptions{CompressionCodec: "snappy", RotateSizeBytes: -1}
	ctx = focuswriters.WithParquetOptions(ctx, opts)
	w, _, err := focuswriters.NewWriter(ctx, tmpPath, "parquet", focustypes.GetFocusV12Schema())
	if err != nil {
		return 0, "", err
	}
	if err := w.Write(ctx, []focustypes.FocusRecord{rec}); err != nil {
		_ = w.Close()
		return 0, "", err
	}
	if err := w.Close(); err != nil {
		return 0, "", err
	}

	data, err := os.ReadFile(tmpPath) // #nosec G304 - controlled temp path
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
	_ = os.Remove(tmpPath)
	sum := sha256.Sum256(data)
	return int64(len(data)), hex.EncodeToString(sum[:]), nil
}

func (e *ParquetExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatParquet}
}
