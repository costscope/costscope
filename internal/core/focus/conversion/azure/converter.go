package azure

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	pathLegacy  = "legacy"
	pathUnified = "unified"
)

// AzureConverter implements Azure Cost Management export to FOCUS conversion.
// This provider-scoped implementation mirrors the legacy root converter behavior
// while avoiding import cycles with a clear provider package boundary.
type AzureConverter struct {
	logger           *logging.Logger
	mappingRules     *types.MappingRules
	streamingEnabled bool
}

const (
	extCSV  = ".csv"
	extJSON = ".json"
)

// NewAzureConverter creates a new Azure converter (provider-scoped).
func NewAzureConverter() *AzureConverter {
	return &AzureConverter{
		logger:           logging.NewLogger(logging.LevelInfo),
		mappingRules:     getAzureMappingRules(),
		streamingEnabled: true,
	}
}

// Convert performs Azure export to FOCUS conversion (delegates to ConvertStream).
func (az *AzureConverter) Convert(ctx context.Context, config *types.ConversionConfig) (*types.ConversionResult, error) {
	if config.Streaming {
		return az.ConvertStream(ctx, config, nil)
	}
	return az.ConvertStream(ctx, config, nil)
}

// ConvertCostExport converts Azure Cost Management exports (alias to Convert).
func (az *AzureConverter) ConvertCostExport(ctx context.Context, config *types.ConversionConfig) (*types.ConversionResult, error) {
	return az.Convert(ctx, config)
}

// ConvertStream performs streaming conversion from CSV/JSON inputs.
func (az *AzureConverter) ConvertStream(ctx context.Context, config *types.ConversionConfig, progressCallback types.ProgressCallback) (*types.ConversionResult, error) { // reduced complexity via helper extraction
	startTime := time.Now()
	modeLabel := pathLegacy
	// Optional streaming invariants aggregator (mirrors AWS pattern)
	var invAgg *quality.InvariantsAggregator
	if config.InvariantsEnabled {
		invAgg = quality.NewInvariantsAggregator()
	}

	if config.UseUnifiedMapper {
		modeLabel = pathUnified
	}

	progress := &types.ConversionProgress{
		ConversionId: config.ConversionId,
		Status:       string(types.StatusRunning),
		StartTime:    startTime,
	}

	result := &types.ConversionResult{
		Success:          false,
		ConversionId:     config.ConversionId,
		StartTime:        startTime,
		InputFile:        config.InputPath,
		InputFormat:      "AZURE_COST_DETAILS",
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

	// Initialize writer with parquet options
	wctx := writers.WithParquetOptions(ctx, &config.Parquet)
	dw, outFmt, err := writers.NewWriter(wctx, config.OutputPath, config.OutputFormat, az.GetSchema())
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

	var handlerErr error
	recordCount, processedRecords, errorRecords, handlerErr = az.dispatchStream(ctx, dispatchParams{
		ctx:              ctx,
		config:           config,
		reader:           reader,
		ext:              ext,
		dw:               dw,
		progress:         progress,
		startTime:        startTime,
		progressCallback: progressCallback,
		modeLabel:        modeLabel,
		invAgg:           invAgg,
	})
	if handlerErr != nil {
		return nil, handlerErr
	}

	if err := dw.Flush(ctx); err != nil {
		return nil, fmt.Errorf("flush failed: %w", err)
	}

	az.finalizeResultAndTelemetry(result, startTime, recordCount, processedRecords, errorRecords, invAgg, config, modeLabel)
	return result, nil
}

// dispatchStream encapsulates the branching logic for unified vs legacy and CSV vs JSON.
type dispatchParams struct {
	ctx              context.Context
	config           *types.ConversionConfig
	reader           io.ReadCloser
	ext              string
	dw               types.DataWriter
	progress         *types.ConversionProgress
	startTime        time.Time
	progressCallback types.ProgressCallback
	modeLabel        string
	invAgg           *quality.InvariantsAggregator
}

func (az *AzureConverter) dispatchStream(ctx context.Context, p dispatchParams) (int64, int64, int64, error) { //nolint:cyclop
	var recordCount, processedRecords, errorRecords int64
	if p.config.UseUnifiedMapper {
		return az.handleUnified(ctx, p, &recordCount, &processedRecords, &errorRecords)
	}
	return az.handleLegacy(ctx, p, &recordCount, &processedRecords, &errorRecords)
}

func (az *AzureConverter) handleUnified(ctx context.Context, p dispatchParams, rc, pc, ec *int64) (int64, int64, int64, error) { //nolint:cyclop
	adapter := mapping.NewAdapter()
	fmCfg := adapter.FromRules(az.mappingRules)
	fm, merr := mapping.NewFieldMapper(fmCfg)
	if merr != nil {
		return 0, 0, 0, fmt.Errorf("failed to init unified field mapper: %w", merr)
	}
	switch p.ext {
	case extCSV:
		baseMapper := az.unifiedCSVMapper(p.config, fm)
		wrapped := az.wrapCSVMapperWithInvariants(baseMapper, p.invAgg, p.modeLabel)
		rct, pct, errc, perr := ProcessCSV(ctx, p.reader, p.config, p.dw, p.progress, p.startTime, p.progressCallback, p.modeLabel, wrapped)
		if perr != nil {
			return 0, 0, 0, perr
		}
		*rc, *pc, *ec = rct, pct, errc
	case extJSON:
		js, jerr := NewJSONStreamFromReader(p.reader)
		if jerr != nil {
			return 0, 0, 0, jerr
		}
		wrappedMap := az.wrapJSONMapperWithInvariants(func(o map[string]any) types.FocusRecord { return az.mapObjectUnified(fm)(o) }, p.invAgg, p.modeLabel)
		rct, pct, errc, perr := ProcessJSON(ctx, js, p.config, p.dw, p.modeLabel, wrappedMap, buildRecMap, func(rec map[string]string, fr *types.FocusRecord, cfg *types.ConversionConfig) {
			az.postProcessUnified(rec, fr, cfg)
		}, forceJSONChargeDescription, func(e error) { az.logger.Error(fmt.Sprintf("JSON decode error: %v", e)) })
		if perr != nil {
			return 0, 0, 0, perr
		}
		*rc, *pc, *ec = rct, pct, errc
	default:
		return 0, 0, 0, fmt.Errorf("unsupported Azure input format: %s (expected .csv or .json)", p.ext)
	}
	return *rc, *pc, *ec, nil
}

func (az *AzureConverter) handleLegacy(ctx context.Context, p dispatchParams, rc, pc, ec *int64) (int64, int64, int64, error) { //nolint:cyclop
	switch p.ext {
	case extCSV:
		baseMapper := az.legacyCSVMapper()
		wrapped := az.wrapCSVMapperWithInvariants(baseMapper, p.invAgg, p.modeLabel)
		rct, pct, errc, perr := ProcessCSV(ctx, p.reader, p.config, p.dw, p.progress, p.startTime, p.progressCallback, p.modeLabel, wrapped)
		if perr != nil {
			return 0, 0, 0, perr
		}
		*rc, *pc, *ec = rct, pct, errc
	case extJSON:
		js, jerr := NewJSONStreamFromReader(p.reader)
		if jerr != nil {
			return 0, 0, 0, jerr
		}
		wrappedMap := az.wrapJSONMapperWithInvariants(func(o map[string]any) types.FocusRecord { return az.mapObjectLegacy(o) }, p.invAgg, p.modeLabel)
		rct, pct, errc, perr := ProcessJSON(ctx, js, p.config, p.dw, p.modeLabel, wrappedMap, buildRecMap, az.postProcessLegacy, forceJSONChargeDescription, func(e error) { az.logger.Error(fmt.Sprintf("JSON decode error: %v", e)) })
		if perr != nil {
			return 0, 0, 0, perr
		}
		*rc, *pc, *ec = rct, pct, errc
	default:
		return 0, 0, 0, fmt.Errorf("unsupported Azure input format: %s (expected .csv or .json)", p.ext)
	}
	return *rc, *pc, *ec, nil
}

// wrap helpers
func (az *AzureConverter) wrapCSVMapperWithInvariants(base func([]string, [][]string) ([]types.FocusRecord, int), invAgg *quality.InvariantsAggregator, mode string) func([]string, [][]string) ([]types.FocusRecord, int) {
	if invAgg == nil {
		return base
	}
	return func(h []string, rows [][]string) ([]types.FocusRecord, int) {
		recs, errs := base(h, rows)
		for i := range recs {
			invAgg.Add(recs[i])
			telemetry.InvariantsRows.WithLabelValues("azure", mode).Add(1)
		}
		return recs, errs
	}
}
func (az *AzureConverter) wrapJSONMapperWithInvariants(base func(map[string]any) types.FocusRecord, invAgg *quality.InvariantsAggregator, mode string) func(map[string]any) types.FocusRecord {
	if invAgg == nil {
		return base
	}
	return func(obj map[string]any) types.FocusRecord {
		fr := base(obj)
		invAgg.Add(fr)
		telemetry.InvariantsRows.WithLabelValues("azure", mode).Add(1)
		return fr
	}
}

// finalizeResultAndTelemetry consolidates tail work after processing to keep main method small.
func (az *AzureConverter) finalizeResultAndTelemetry(result *types.ConversionResult, start time.Time, recordCount, processedRecords, errorRecords int64, invAgg *quality.InvariantsAggregator, config *types.ConversionConfig, modeLabel string) {
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
	if invAgg != nil { // finalize invariants
		invStart := time.Now()
		inv := invAgg.Produce()
		telemetry.InvariantsComputeDuration.WithLabelValues("azure", modeLabel).Observe(time.Since(invStart).Seconds())
		baselineLabel := "no"
		if config.InvariantsBaseline != "" {
			baselineLabel = "yes"
		}
		telemetry.InvariantsFeatureRuns.WithLabelValues("azure", modeLabel, baselineLabel).Inc()
		result.Invariants = inv
	}
	telemetry.ConverterDuration.WithLabelValues("azure", modeLabel).Observe(result.Duration.Seconds())
	if processedRecords > 0 {
		telemetry.ConverterRecords.WithLabelValues("azure", modeLabel, "ok").Add(float64(processedRecords))
	}
	if errorRecords > 0 {
		telemetry.ConverterRecords.WithLabelValues("azure", modeLabel, "error").Add(float64(errorRecords))
	}
	telemetry.UnifiedMapperDuration.WithLabelValues("azure", modeLabel).Observe(result.Duration.Seconds())
	if processedRecords > 0 {
		telemetry.UnifiedMapperRows.WithLabelValues("azure", modeLabel).Add(float64(processedRecords))
	}
	if errorRecords > 0 {
		telemetry.UnifiedMapperErrors.WithLabelValues("azure", modeLabel).Add(float64(errorRecords))
	}
	az.logger.Info(fmt.Sprintf("Azure conversion completed: %d processed, %d errors", processedRecords, errorRecords))
}

// ProcessChunk is not used in the Azure provider streaming pipeline; implemented to satisfy StreamingConverter.
// Universal converter only calls ConvertStream for streaming execution.
func (az *AzureConverter) ProcessChunk(ctx context.Context, chunk []byte, chunkNumber int) ([]types.FocusRecord, error) { //nolint:revive,unused
	return nil, fmt.Errorf("azure: ProcessChunk not supported; use ConvertStream")
}

// ValidateInput validates input file existence and minimal schema
func (az *AzureConverter) ValidateInput(_ context.Context, config *types.ConversionConfig) error { //nolint:cyclop
	if _, err := os.Stat(config.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", config.InputPath)
	}
	ext := strings.ToLower(filepath.Ext(config.InputPath))
	if ext == ".gz" {
		base := strings.TrimSuffix(config.InputPath, ".gz")
		ext = strings.ToLower(filepath.Ext(base))
	}
	switch ext {
	case ".csv":
		// Lightweight header validation
		f, err := os.Open(config.InputPath) // #nosec G304 - validated by ValidateInput
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		r := csv.NewReader(bufio.NewReader(f))
		headers, err := r.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV headers: %w", err)
		}
		req := map[string]bool{"SubscriptionId": false, "Quantity": false, "UnitOfMeasure": false}
		for _, h := range headers {
			if _, ok := req[strings.TrimSpace(h)]; ok {
				req[strings.TrimSpace(h)] = true
			}
		}
		for col, present := range req {
			if !present {
				az.logger.Warn(fmt.Sprintf("CSV missing typical column: %s", col))
			}
		}
		return nil
	case ".json":
		return nil
	case ".parquet":
		return fmt.Errorf("parquet input not yet supported in this build")
	default:
		return fmt.Errorf("unsupported input format: %s", ext)
	}
}

// EstimateConversion provides rough estimates based on file size
func (az *AzureConverter) EstimateConversion(_ context.Context, config *types.ConversionConfig) (*types.ConversionEstimate, error) {
	return c.EstimateFromFile(config.InputPath, config)
}

// GetSupportedFormats returns supported formats for Azure
func (az *AzureConverter) GetSupportedFormats() *types.SupportedFormats {
	return c.CommonSupportedFormats
}

// GetSchema returns FOCUS schema
func (az *AzureConverter) GetSchema() *types.FocusSchema { return types.GetFocusV12Schema() }

// ProcessBillingAccount processes Azure billing account metadata (noop placeholder)
func (az *AzureConverter) ProcessBillingAccount(_ context.Context, _ string) error { return nil }

// =============================== Local helpers (no import cycles) ==================

// unifiedCSVMapper returns a mapper func compatible with ProcessCSV that maps CSV rows using
// the unified FieldMapper and applies Azure-specific enrichment and normalization parity.
func (az *AzureConverter) unifiedCSVMapper(config *types.ConversionConfig, fm *mapping.FieldMapper) func([]string, [][]string) ([]types.FocusRecord, int) {
	return func(headers []string, chunk [][]string) ([]types.FocusRecord, int) { //nolint:cyclop
		// Build index map once
		idx := make(map[int]string, len(headers))
		for i, h := range headers {
			idx[i] = h
		}
		out := make([]types.FocusRecord, 0, len(chunk))
		errs := 0
		for _, rec := range chunk {
			// Construct string map for both mapping and downstream enrichment
			recMap := make(map[string]string, len(rec))
			for i, v := range rec {
				recMap[idx[i]] = v
			}
			// Map via unified mapper (expects interface{} values)
			obj := make(map[string]interface{}, len(recMap))
			for k, v := range recMap {
				obj[k] = v
			}
			fr, merr := fm.MapToFOCUS(obj)
			if merr != nil {
				errs++
				continue
			}
			// Post-processing parity (mirror root unified path)
			if v := strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(recMap, "BillingAccountId"), FirstNonEmptyField(recMap, "EnrollmentNumber"), FirstNonEmptyField(recMap, "BillingProfileId"))); v != "" {
				fr.BillingAccountId = v
			}
			fr.BillingCurrency = strings.ToUpper(FirstNonEmptyValue(FirstNonEmptyField(recMap, "BillingCurrency"), FirstNonEmptyField(recMap, "Currency")))
			if reg := NormalizeRegion(FirstNonEmptyValue(FirstNonEmptyField(recMap, "ResourceLocation"), FirstNonEmptyField(recMap, "Location"))); strings.TrimSpace(reg) != "" {
				r := reg
				fr.Region = &r
			}
			if tagsStr := strings.TrimSpace(FirstNonEmptyField(recMap, "Tags")); tagsStr != "" {
				fr.Tags = parseTags(tagsStr)
			}
			ApplyBenefitsMap(recMap, fr)
			cost := DeriveEffectiveCost(recMap)
			if fr.EffectiveCost == 0 && cost != 0 {
				fr.EffectiveCost = cost
			}
			if bc := GetBilledCostPtr(recMap); bc != nil {
				fr.BilledCost = bc
			}
			fr.ChargeCategory = DeriveChargeCategory(recMap, cost)
			// Shared discount normalization right after base category
			AzureEnsureDiscount(FirstNonEmptyValue(FirstNonEmptyField(recMap, "ChargeType")), FirstNonEmptyValue(FirstNonEmptyField(recMap, "BillingType")), fr)
			pc, cc := DerivePricing(recMap)
			if strings.TrimSpace(fr.PricingCategory) == "" {
				fr.PricingCategory = pc
			}
			if strings.TrimSpace(fr.ChargeClass) == "" {
				fr.ChargeClass = cc
			}
			fr.SourceProvider = providerAzure
			fr.SourceFileName = filepath.Base(config.InputPath)
			// Provider-agnostic classification adjustments (shared)
			c.ApplyUnifiedClassification("azure", recMap, fr)
			// Enrichment parity with unified path
			applyGenericUnifiedEnrichmentAzure(fr)
			EnrichUnified(recMap, fr, true)
			// Normalization parity (unified path)
			applyUnifiedNormalizationAzure(fr)
			out = append(out, *fr)
		}
		return out, errs
	}
}

// legacyCSVMapper returns a mapper using the provider-scoped FullRowMapper with injected deps.
func (az *AzureConverter) legacyCSVMapper() func([]string, [][]string) ([]types.FocusRecord, int) {
	return func(headers []string, records [][]string) ([]types.FocusRecord, int) {
		idx := NewHeaderIndex(headers)
		mapper := NewFullRowMapperWithDeps(
			idx,
			func(chargeType, billingType string, effectiveCost float64, billedCost *float64, candidateValues []string, provider string) string {
				return classifyChargeCategoryAzure(chargeType, billingType, effectiveCost, billedCost, candidateValues, provider)
			},
			func(h *HeaderIndex, r []string, fr *types.FocusRecord) { ApplyBenefitsRow(h, r, fr) },
			func(s string) types.Tags { return parseTags(s) },
			func(chargeType, billingType string, fr *types.FocusRecord) bool {
				return AzureEnsureDiscount(chargeType, billingType, fr)
			},
			time.Now,
		)
		out := make([]types.FocusRecord, 0, len(records))
		errs := 0
		for _, r := range records {
			if len(r) == 0 {
				errs++
				continue
			}
			fr, err := mapper.Map(r)
			if err != nil {
				errs++
				continue
			}
			out = append(out, fr)
		}
		return out, errs
	}
}

// mapObjectUnified builds a FocusRecord from a JSON object via FieldMapper and applies enrichment.
func (az *AzureConverter) mapObjectUnified(fm *mapping.FieldMapper) func(map[string]any) types.FocusRecord {
	return func(obj map[string]any) types.FocusRecord {
		fr, _ := fm.MapToFOCUS(obj)
		rec := buildRecMap(obj)
		// Post processing mirrors CSV unified
		if v := strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountId"), FirstNonEmptyField(rec, "EnrollmentNumber"), FirstNonEmptyField(rec, "BillingProfileId"))); v != "" {
			fr.BillingAccountId = v
		}
		fr.BillingCurrency = strings.ToUpper(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingCurrency"), FirstNonEmptyField(rec, "Currency")))
		if reg := NormalizeRegion(FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceLocation"), FirstNonEmptyField(rec, "Location"))); strings.TrimSpace(reg) != "" {
			r := reg
			fr.Region = &r
		}
		if tagsStr := strings.TrimSpace(FirstNonEmptyField(rec, "Tags")); tagsStr != "" {
			fr.Tags = parseTags(tagsStr)
		}
		ApplyBenefitsMap(rec, fr)
		cost := DeriveEffectiveCost(rec)
		if fr.EffectiveCost == 0 && cost != 0 {
			fr.EffectiveCost = cost
		}
		if bc := GetBilledCostPtr(rec); bc != nil {
			fr.BilledCost = bc
		}
		// Derive base charge category from explicit fields (unconditional for parity with CSV unified path)
		fr.ChargeCategory = DeriveChargeCategory(rec, fr.EffectiveCost)
		pc, cc := DerivePricing(rec)
		if strings.TrimSpace(fr.PricingCategory) == "" {
			fr.PricingCategory = pc
		}
		if strings.TrimSpace(fr.ChargeClass) == "" {
			fr.ChargeClass = cc
		}
		fr.SourceProvider = providerAzure
		// Azure discount normalization (substring match) to maintain parity
		AzureEnsureDiscount(
			FirstNonEmptyValue(FirstNonEmptyField(rec, "ChargeType")),
			FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingType")),
			fr,
		)
		// Classification adjustments shared
		c.ApplyUnifiedClassification("azure", rec, fr)
		applyGenericUnifiedEnrichmentAzure(fr)
		EnrichUnified(rec, fr, false)
		applyUnifiedNormalizationAzure(fr)
		return *fr
	}
}

// mapObjectLegacy maps a JSON object to FocusRecord (legacy fast path parity).
func (az *AzureConverter) mapObjectLegacy(obj map[string]any) types.FocusRecord {
	rec := buildRecMap(obj)
	quantity := ParseFloat(FirstNonEmptyValue(FirstNonEmptyField(rec, "Quantity")))
	start, end := DeriveDates(rec)
	pc, cc := DerivePricing(rec)
	cost := DeriveEffectiveCost(rec)
	category := DeriveChargeCategory(rec, cost)
	billedPtr := GetBilledCostPtr(rec)

	fr := types.FocusRecord{
		BillingAccountId:   FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountId"), FirstNonEmptyField(rec, "EnrollmentNumber"), FirstNonEmptyField(rec, "BillingProfileId")),
		BillingAccountName: FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountName"), FirstNonEmptyField(rec, "BillingProfileName")),
		BillingCurrency:    strings.ToUpper(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingCurrency"), FirstNonEmptyField(rec, "Currency"))),
		BillingPeriodStart: start,
		BillingPeriodEnd:   end,
		ChargeCategory:     category,
		ChargeClass:        cc,
		ChargeDescription:  FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterName"), FirstNonEmptyField(rec, "Product"), FirstNonEmptyField(rec, "ServiceName")),
		ChargeFrequency:    "Daily",
		ChargePeriodStart:  start,
		ChargePeriodEnd:    end,
		ChargeSubcategory:  FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterSubCategory"), FirstNonEmptyField(rec, "ServiceName"), FirstNonEmptyField(rec, "Product"), FirstNonEmptyField(rec, "ServiceInfo2"), FirstNonEmptyField(rec, "Operation")),
		EffectiveCost:      cost,
		InvoiceIssuerName:  types.ProviderNames.Azure,
		ListCost:           ParseFloat(FirstNonEmptyValue(FirstNonEmptyField(rec, "RetailPrice"))) * quantity,
		ListUnitPrice:      ParseFloat(FirstNonEmptyValue(FirstNonEmptyField(rec, "RetailPrice"), FirstNonEmptyField(rec, "UnitPrice"))),
		PricingCategory:    pc,
		PricingQuantity:    quantity,
		PricingUnit:        FirstNonEmptyValue(FirstNonEmptyField(rec, "UnitOfMeasure"), FirstNonEmptyField(rec, "MeterUnit")),
		ProviderName:       types.ProviderNames.Azure,
		PublisherName:      "Microsoft",
		ResourceId:         FirstNonEmptyField(rec, "ResourceId"),
		ResourceName:       FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceName"), FirstNonEmptyField(rec, "InstanceName")),
		ResourceType:       FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceType"), FirstNonEmptyField(rec, "ServiceTier")),
		ServiceCategory:    FirstNonEmptyValue(FirstNonEmptyField(rec, "ServiceFamily"), FirstNonEmptyField(rec, "MeterCategory")),
		ServiceName:        FirstNonEmptyValue(FirstNonEmptyField(rec, "ServiceName"), FirstNonEmptyField(rec, "Product")),
		SkuId:              FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterId"), FirstNonEmptyField(rec, "SkuId")),
		SkuPriceId:         FirstNonEmptyValue(FirstNonEmptyField(rec, "PartNumber"), FirstNonEmptyField(rec, "ProductOrderNumber")),
		SubAccountId:       FirstNonEmptyField(rec, "SubscriptionId"),
		SubAccountName:     FirstNonEmptyField(rec, "SubscriptionName"),
		UsageQuantity:      quantity,
		UsageUnit:          FirstNonEmptyValue(FirstNonEmptyField(rec, "UnitOfMeasure"), FirstNonEmptyField(rec, "MeterUnit")),
		AvailabilityZone:   nil,
		BilledCost:         billedPtr,
		Region: func() *string {
			reg := NormalizeRegion(FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceLocation"), FirstNonEmptyField(rec, "Location")))
			if strings.TrimSpace(reg) == "" {
				return nil
			}
			r := reg
			return &r
		}(),
		SourceProvider:      providerAzure,
		ConversionTimestamp: time.Now(),
	}
	AzureEnsureDiscount(FirstNonEmptyValue(FirstNonEmptyField(rec, "ChargeType")), FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingType")), &fr)
	if tagsStr := strings.TrimSpace(FirstNonEmptyField(rec, "Tags")); tagsStr != "" {
		fr.Tags = parseTags(tagsStr)
	}
	ApplyBenefitsMap(rec, &fr)
	if fr.ChargeCategory == types.ChargeCategories.Usage {
		if fr.EffectiveCost < 0 || (fr.BilledCost != nil && *fr.BilledCost < 0) {
			fr.ChargeCategory = types.ChargeCategories.Credit
		}
	}
	return fr
}

// postProcessLegacy mirrors root azurePostProcessLegacy without duplicate metrics.
func (az *AzureConverter) postProcessLegacy(rec map[string]string, fr *types.FocusRecord, config *types.ConversionConfig) {
	if v := strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountId"), FirstNonEmptyField(rec, "EnrollmentNumber"), FirstNonEmptyField(rec, "BillingProfileId"))); v != "" {
		fr.BillingAccountId = v
	}
	fr.BillingCurrency = strings.ToUpper(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingCurrency"), FirstNonEmptyField(rec, "Currency")))
	if reg := NormalizeRegion(FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceLocation"), FirstNonEmptyField(rec, "Location"))); strings.TrimSpace(reg) != "" {
		r := reg
		fr.Region = &r
	}
	if tagsStr := strings.TrimSpace(FirstNonEmptyField(rec, "Tags")); tagsStr != "" {
		fr.Tags = parseTags(tagsStr)
	}
	ApplyBenefitsMap(rec, fr)
	cost := DeriveEffectiveCost(rec)
	if fr.EffectiveCost == 0 && cost != 0 {
		fr.EffectiveCost = cost
	}
	if bc := GetBilledCostPtr(rec); bc != nil {
		fr.BilledCost = bc
	}
	if strings.TrimSpace(fr.ChargeCategory) == "" {
		fr.ChargeCategory = DeriveChargeCategory(rec, cost)
	}
	pc, cc := DerivePricing(rec)
	if strings.TrimSpace(fr.PricingCategory) == "" {
		fr.PricingCategory = pc
	}
	if strings.TrimSpace(fr.ChargeClass) == "" {
		fr.ChargeClass = cc
	}
	fr.SourceProvider = providerAzure
	fr.SourceFileName = filepath.Base(config.InputPath)
	c.ApplyUnifiedClassification("azure", rec, fr)
	EnrichUnified(rec, fr, false)
	// Legacy path normalization parity
	c.NormalizeFocusRecord(fr)
}

// postProcessUnified performs enrichment and normalization for unified JSON path and emits metrics in caller.
func (az *AzureConverter) postProcessUnified(rec map[string]string, fr *types.FocusRecord, config *types.ConversionConfig) {
	if v := strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountId"), FirstNonEmptyField(rec, "EnrollmentNumber"), FirstNonEmptyField(rec, "BillingProfileId"))); v != "" {
		fr.BillingAccountId = v
	}
	fr.BillingCurrency = strings.ToUpper(FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingCurrency"), FirstNonEmptyField(rec, "Currency")))
	if reg := NormalizeRegion(FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceLocation"), FirstNonEmptyField(rec, "Location"))); strings.TrimSpace(reg) != "" {
		r := reg
		fr.Region = &r
	}
	if tagsStr := strings.TrimSpace(FirstNonEmptyField(rec, "Tags")); tagsStr != "" {
		fr.Tags = parseTags(tagsStr)
	}
	ApplyBenefitsMap(rec, fr)
	cost := DeriveEffectiveCost(rec)
	if fr.EffectiveCost == 0 && cost != 0 {
		fr.EffectiveCost = cost
	}
	if bc := GetBilledCostPtr(rec); bc != nil {
		fr.BilledCost = bc
	}
	if strings.TrimSpace(fr.ChargeCategory) == "" {
		fr.ChargeCategory = DeriveChargeCategory(rec, cost)
	}
	pc, cc := DerivePricing(rec)
	if strings.TrimSpace(fr.PricingCategory) == "" {
		fr.PricingCategory = pc
	}
	if strings.TrimSpace(fr.ChargeClass) == "" {
		fr.ChargeClass = cc
	}
	fr.SourceProvider = providerAzure
	fr.SourceFileName = filepath.Base(config.InputPath)
	c.ApplyUnifiedClassification("azure", rec, fr)
	applyGenericUnifiedEnrichmentAzure(fr)
	EnrichUnified(rec, fr, false)
	applyUnifiedNormalizationAzure(fr)
}

// buildRecMap converts a JSON object into a string map used by helpers.
func buildRecMap(obj map[string]any) map[string]string {
	recMap := make(map[string]string, len(obj))
	for k, v := range obj {
		recMap[k] = fmt.Sprint(v)
	}
	return recMap
}

// forceJSONChargeDescription normalizes ChargeDescription for legacy JSON parity with unified.
func forceJSONChargeDescription(fr *types.FocusRecord, obj map[string]any) {
	if fr == nil {
		return
	}
	tmp := buildRecMap(obj)
	fr.ChargeDescription = FirstNonEmptyValue(FirstNonEmptyField(tmp, "MeterName"), FirstNonEmptyField(tmp, "Product"), FirstNonEmptyField(tmp, "ServiceName"))
}

// parseTags attempts to parse Azure Tags column as JSON into map.
func parseTags(s string) types.Tags {
	var m map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &m); err == nil {
		return m
	}
	return nil
}

// classifyChargeCategoryAzure mirrors the root classification logic used for initial category.
func classifyChargeCategoryAzure(chargeType, billingType string, effectiveCost float64, billedCost *float64, candidateValues []string, provider string) string { //nolint:cyclop
	ct := strings.ToLower(strings.TrimSpace(chargeType))
	bt := strings.ToLower(strings.TrimSpace(billingType))
	first := ct
	if first == "" {
		first = bt
	}
	hasDiscount := first == tokenDiscount || first == tokenUsageDisc || strings.Contains(first, tokenDiscount)
	switch first {
	case "usage":
		// continue
	case "purchase", "buy", kwReservation, kwSavingsPlan1, kwSavingsPlan2:
		return types.ChargeCategories.Purchase
	case "credit", "refund":
		return types.ChargeCategories.Credit
	case "tax":
		return types.ChargeCategories.Tax
	case tokenDiscount, tokenUsageDisc:
		if provider != providerAzure { // azure promotion handled later
			return liTypeDiscount
		}
	}
	if (effectiveCost < 0 || (billedCost != nil && *billedCost < 0)) && (provider != providerAzure || !hasDiscount) {
		return types.ChargeCategories.Credit
	}
	for _, v := range candidateValues {
		s := strings.TrimSpace(v)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "-") || (strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")")) {
			if provider != providerAzure || !hasDiscount {
				return types.ChargeCategories.Credit
			}
		}
	}
	return types.ChargeCategories.Usage
}

// applyGenericUnifiedEnrichmentAzure duplicates the tiny provider-agnostic enrichment used in unified path
// to avoid import cycles with the root package.
func applyGenericUnifiedEnrichmentAzure(fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.ChargeFrequency) == "" {
		fr.ChargeFrequency = "Daily"
	}
	if fr.PricingQuantity == 0 && fr.UsageQuantity != 0 {
		fr.PricingQuantity = fr.UsageQuantity
	}
	if fr.PricingUnit == "" {
		fr.PricingUnit = fr.UsageUnit
	}
	if fr.EffectiveCost < 0 || (fr.BilledCost != nil && *fr.BilledCost < 0) {
		if fr.ChargeCategory == "" || fr.ChargeCategory == types.ChargeCategories.Usage {
			fr.ChargeCategory = types.ChargeCategories.Credit
		}
	}
	if fr.ConversionTimestamp.IsZero() {
		fr.ConversionTimestamp = time.Now()
	}
}

// applyUnifiedNormalizationAzure mirrors the root helper using common normalization functions.
func applyUnifiedNormalizationAzure(fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if fr.Region != nil {
		fr.Region = c.NormalizeRegion("azure", fr.Region)
	}
	fr.BillingCurrency = c.NormalizeCurrency(fr.BillingCurrency)
	if fr.UsageUnit != "" {
		fr.UsageUnit = c.NormalizeUnit(fr.UsageUnit)
	}
	if fr.PricingUnit != "" {
		fr.PricingUnit = c.NormalizeUnit(fr.PricingUnit)
	}
	fr.ChargeCategory = strings.TrimSpace(fr.ChargeCategory)
	fr.ChargeClass = strings.TrimSpace(fr.ChargeClass)
	fr.PricingCategory = strings.TrimSpace(fr.PricingCategory)
}
