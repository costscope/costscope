package gcp

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"strings"
	"time"

	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	"github.com/costscope/costscope/internal/core/focus/quality"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/focus/writers"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/framework/mapping"
)

const (
	gcpPathLegacy  = "legacy"
	gcpPathUnified = "unified"
)

// GCPConverter implements GCP Billing export to FOCUS conversion.
// Mirrors the legacy behavior while living in the provider package to preserve public APIs.
type GCPConverter struct {
	logger           *logging.Logger
	mappingRules     *types.MappingRules
	streamingEnabled bool
}

// NewGCPConverter creates a new provider-scoped converter.
func NewGCPConverter() *GCPConverter {
	return &GCPConverter{
		logger:           logging.NewLogger(logging.LevelInfo),
		mappingRules:     GetMappingRules(),
		streamingEnabled: true,
	}
}

// Convert delegates to ConvertStream (streaming default kept for parity).
func (g *GCPConverter) Convert(ctx context.Context, config *types.ConversionConfig) (*types.ConversionResult, error) {
	return g.ConvertStream(ctx, config, nil)
}

// ConvertStream performs streaming conversion from CSV/JSON inputs.
func (g *GCPConverter) ConvertStream(ctx context.Context, config *types.ConversionConfig, progressCallback types.ProgressCallback) (*types.ConversionResult, error) {
	start := time.Now()
	pathLabel := gcpPathLegacy
	if config.UseUnifiedMapper {
		pathLabel = gcpPathUnified
	}

	progress := &types.ConversionProgress{
		ConversionId: config.ConversionId,
		Status:       string(types.StatusRunning),
		StartTime:    start,
	}

	result := &types.ConversionResult{
		Success:          false,
		ConversionId:     config.ConversionId,
		StartTime:        start,
		InputFile:        config.InputPath,
		InputFormat:      "GCP_BILLING_EXPORT",
		OutputFile:       config.OutputPath,
		OutputFormat:     "FOCUS_NDJSON",
		FocusVersion:     "1.2",
		SchemaVersion:    "FOCUS_1.2",
		ConverterVersion: "CostScope_1.0",
	}

	// Open input (supports .gz) and detect inner extension
	reader, ext, err := OpenInput(config.InputPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	// Initialize writer
	wctx := writers.WithParquetOptions(ctx, &config.Parquet)
	dw, outFmt, err := writers.NewWriter(wctx, config.OutputPath, config.OutputFormat, g.GetSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to open writer: %w", err)
	}
	result.OutputFormat = outFmt
	defer func() { _ = dw.Close() }()

	var (
		recordCount      int64
		processedRecords int64
		errorRecords     int64
	)

	// Optional streaming invariants aggregator (mirrors AWS/Azure pattern)
	var invAgg *quality.InvariantsAggregator
	if config.InvariantsEnabled {
		invAgg = quality.NewInvariantsAggregator()
	}

	if config.UseUnifiedMapper {
		rc, pc, ec, err := g.convertUnifiedByExt(ctx, ext, reader, config, dw, progress, start, progressCallback, pathLabel, invAgg)
		if err != nil {
			return nil, err
		}
		recordCount, processedRecords, errorRecords = rc, pc, ec
	} else {
		rc, pc, ec, err := g.convertLegacyByExt(ctx, ext, reader, config, dw, progress, start, progressCallback, pathLabel, invAgg)
		if err != nil {
			return nil, err
		}
		recordCount, processedRecords, errorRecords = rc, pc, ec
	}

	if err := dw.Flush(ctx); err != nil {
		return nil, fmt.Errorf("flush failed: %w", err)
	}

	end := time.Now()
	result.EndTime = end
	result.Duration = end.Sub(start)
	result.InputRecords = recordCount
	result.OutputRecords = processedRecords
	result.SuccessRecords = processedRecords
	result.ErrorRecords = errorRecords
	if result.Duration > 0 {
		result.RecordsPerSecond = float64(processedRecords) / result.Duration.Seconds()
	}
	result.Success = true

	// Finalize invariants
	if invAgg != nil {
		invStart := time.Now()
		inv := invAgg.Produce()
		telemetry.InvariantsComputeDuration.WithLabelValues("gcp", pathLabel).Observe(time.Since(invStart).Seconds())
		baselineLabel := "no"
		if config.InvariantsBaseline != "" {
			baselineLabel = "yes"
		}
		telemetry.InvariantsFeatureRuns.WithLabelValues("gcp", pathLabel, baselineLabel).Inc()
		result.Invariants = inv
	}

	// Telemetry summary (mirror root behavior)
	telemetry.ConverterDuration.WithLabelValues("gcp", pathLabel).Observe(result.Duration.Seconds())
	if processedRecords > 0 {
		telemetry.ConverterRecords.WithLabelValues("gcp", pathLabel, "ok").Add(float64(processedRecords))
	}
	if errorRecords > 0 {
		telemetry.ConverterRecords.WithLabelValues("gcp", pathLabel, "error").Add(float64(errorRecords))
	}
	telemetry.UnifiedMapperDuration.WithLabelValues("gcp", pathLabel).Observe(result.Duration.Seconds())
	if processedRecords > 0 {
		telemetry.UnifiedMapperRows.WithLabelValues("gcp", pathLabel).Add(float64(processedRecords))
	}
	if errorRecords > 0 {
		telemetry.UnifiedMapperErrors.WithLabelValues("gcp", pathLabel).Add(float64(errorRecords))
	}

	g.logger.Info(fmt.Sprintf("GCP conversion completed: %d processed, %d errors", processedRecords, errorRecords))
	return result, nil
}

// ProcessChunk is not used; required by interface.
func (g *GCPConverter) ProcessChunk(_ context.Context, _ []byte, _ int) ([]types.FocusRecord, error) { //nolint:revive,unused
	return nil, fmt.Errorf("gcp: ProcessChunk not supported; use ConvertStream")
}

// ValidateInput validates input existence and basic format.
func (g *GCPConverter) ValidateInput(_ context.Context, config *types.ConversionConfig) error { //nolint:cyclop
	if _, err := os.Stat(config.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", config.InputPath)
	}
	// Use provider-aware opener to transparently handle gzip and detect inner type
	rc, innerExt, err := OpenInput(config.InputPath)
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	switch strings.ToLower(innerExt) {
	case ".csv":
		// Lightweight CSV header validation on decompressed stream
		r := csv.NewReader(bufio.NewReader(rc))
		headers, err := r.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV headers: %w", err)
		}
		req := map[string]bool{"usage_start_time": false, "cost": false}
		for _, h := range headers {
			if _, ok := req[strings.TrimSpace(h)]; ok {
				req[strings.TrimSpace(h)] = true
			}
		}
		for col, present := range req {
			if !present {
				g.logger.Warn(fmt.Sprintf("CSV missing typical column: %s", col))
			}
		}
		return nil
	case ".json":
		// Basic presence validated by Stat above
		return nil
	default:
		return fmt.Errorf("unsupported input format: %s (expected .csv or .json, optionally .gz)", innerExt)
	}
}

// EstimateConversion provides simple estimates by file size.
func (g *GCPConverter) EstimateConversion(_ context.Context, config *types.ConversionConfig) (*types.ConversionEstimate, error) {
	return c.EstimateFromFile(config.InputPath, config)
}

// GetSupportedFormats returns supported formats for GCP
func (g *GCPConverter) GetSupportedFormats() *types.SupportedFormats { return c.CommonSupportedFormats }

// GetSchema returns FOCUS schema
func (g *GCPConverter) GetSchema() *types.FocusSchema { return types.GetFocusV12Schema() }

// ---------------------- small path delegates to reduce complexity ----------------------

func (g *GCPConverter) convertUnifiedByExt(
	ctx context.Context,
	ext string,
	reader io.Reader,
	config *types.ConversionConfig,
	dw types.DataWriter,
	progress *types.ConversionProgress,
	start time.Time,
	progressCallback types.ProgressCallback,
	pathLabel string,
	invAgg *quality.InvariantsAggregator,
) (int64, int64, int64, error) {
	adapter := mapping.NewAdapter()
	fmCfg := adapter.FromRules(g.mappingRules)
	fm, merr := mapping.NewFieldMapper(fmCfg)
	if merr != nil {
		return 0, 0, 0, fmt.Errorf("failed to init unified field mapper: %w", merr)
	}
	switch ext {
	case extCSV:
		base := g.unifiedCSVMapper(config, fm, pathLabel)
		wrapped := func(h []string, rows [][]string) ([]types.FocusRecord, int) {
			recs, errs := base(h, rows)
			if invAgg != nil {
				for i := range recs {
					invAgg.Add(recs[i])
					telemetry.InvariantsRows.WithLabelValues("gcp", pathLabel).Add(1)
				}
			}
			return recs, errs
		}
		return ProcessCSV(
			ctx,
			reader,
			config,
			dw,
			progress,
			start,
			progressCallback,
			pathLabel,
			wrapped,
		)
	case extJSON:
		mapper := g.mapObjectUnified(fm, config)
		wrapped := g.wrapJSONMapper(pathLabel, func(obj map[string]any) *types.FocusRecord {
			fr := mapper(obj)
			if invAgg != nil && fr != nil {
				invAgg.Add(*fr)
				telemetry.InvariantsRows.WithLabelValues("gcp", pathLabel).Add(1)
			}
			return fr
		})
		return ProcessJSON(
			ctx,
			reader,
			config,
			dw,
			wrapped,
		)
	default:
		return 0, 0, 0, fmt.Errorf("unsupported GCP input format: %s (expected .csv or .json)", ext)
	}
}

func (g *GCPConverter) convertLegacyByExt(
	ctx context.Context,
	ext string,
	reader io.Reader,
	config *types.ConversionConfig,
	dw types.DataWriter,
	progress *types.ConversionProgress,
	start time.Time,
	progressCallback types.ProgressCallback,
	pathLabel string,
	invAgg *quality.InvariantsAggregator,
) (int64, int64, int64, error) {
	switch ext {
	case extCSV:
		base := g.legacyCSVMapper(pathLabel)
		wrapped := func(h []string, rows [][]string) ([]types.FocusRecord, int) {
			recs, errs := base(h, rows)
			if invAgg != nil {
				for i := range recs {
					invAgg.Add(recs[i])
					telemetry.InvariantsRows.WithLabelValues("gcp", pathLabel).Add(1)
				}
			}
			return recs, errs
		}
		return ProcessCSV(
			ctx,
			reader,
			config,
			dw,
			progress,
			start,
			progressCallback,
			pathLabel,
			wrapped,
		)
	case extJSON:
		mapper := g.mapObjectLegacy(config)
		wrapped := g.wrapJSONMapper(pathLabel, func(obj map[string]any) *types.FocusRecord {
			fr := mapper(obj)
			if invAgg != nil && fr != nil {
				invAgg.Add(*fr)
				telemetry.InvariantsRows.WithLabelValues("gcp", pathLabel).Add(1)
			}
			return fr
		})
		return ProcessJSON(
			ctx,
			reader,
			config,
			dw,
			wrapped,
		)
	default:
		return 0, 0, 0, fmt.Errorf("unsupported GCP input format: %s (expected .csv or .json)", ext)
	}
}
