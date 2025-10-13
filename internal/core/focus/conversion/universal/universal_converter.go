package universal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
	val "github.com/costscope/costscope/internal/core/focus/validation"
	"github.com/costscope/costscope/internal/core/logging"
)

// UniversalConverter implements comprehensive multi-cloud FOCUS conversion
// Moved to its own package to keep provider-specific concerns decoupled.
type UniversalConverter struct {
	logger           *logging.Logger
	converters       map[string]types.Converter
	streamingSupport bool
	mutex            sync.RWMutex
}

// NewUniversalConverter creates a new universal FOCUS converter
func NewUniversalConverter() *UniversalConverter {
	return &UniversalConverter{
		logger:           logging.NewLogger(logging.LevelInfo),
		converters:       make(map[string]types.Converter),
		streamingSupport: true,
	}
}

// RegisterConverter registers a provider-specific converter
func (uc *UniversalConverter) RegisterConverter(provider string, converter types.Converter) error {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	if _, exists := uc.converters[provider]; exists {
		return fmt.Errorf("converter for provider %s already registered", provider)
	}

	uc.converters[provider] = converter
	uc.logger.Info(fmt.Sprintf("Registered converter for provider: %s", provider))
	return nil
}

// Convert performs universal conversion based on provider
func (uc *UniversalConverter) Convert(ctx context.Context, config *types.ConversionConfig) (*types.ConversionResult, error) {
	uc.logger.Info(fmt.Sprintf("Starting FOCUS conversion for provider: %s", config.Provider))

	// Validate configuration
	if err := uc.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get provider-specific converter
	converter, err := uc.getConverter(config.Provider)
	if err != nil {
		return nil, err
	}

	// Prepare conversion context
	conversionCtx := uc.prepareConversionContext(ctx, config)

	// Execute conversion based on streaming preference
	if config.Streaming {
		return uc.convertWithStreaming(conversionCtx, converter, config)
	}

	return uc.convertStandard(conversionCtx, converter, config)
}

// Removed obsolete methods: ConvertBatch, ValidateInput, EstimateConversion, GetSupportedFormats, GetSchema.

// -------------------------------------------------------------------------------------
// Private Methods
// -------------------------------------------------------------------------------------

func (uc *UniversalConverter) validateConfig(config *types.ConversionConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	if config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if config.InputPath == "" {
		return fmt.Errorf("input path is required")
	}
	if config.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if config.Workers <= 0 {
		config.Workers = 4
	}
	if config.ChunkSize <= 0 {
		config.ChunkSize = 10000
	}
	if config.MaxMemoryMB <= 0 {
		config.MaxMemoryMB = 1024
	}
	return nil
}

func (uc *UniversalConverter) getConverter(provider string) (types.Converter, error) {
	uc.mutex.RLock()
	defer uc.mutex.RUnlock()

	converter, exists := uc.converters[provider]
	if !exists {
		return nil, fmt.Errorf("no converter registered for provider: %s", provider)
	}
	return converter, nil
}

func (uc *UniversalConverter) prepareConversionContext(ctx context.Context, config *types.ConversionConfig) context.Context {
	if config.ConversionId == "" {
		config.ConversionId = fmt.Sprintf("conv_%d", time.Now().Unix())
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	if config.CreatedBy == "" {
		config.CreatedBy = "CostScope Universal Converter"
	}
	return ctx
}

func (uc *UniversalConverter) convertWithStreaming(ctx context.Context, converter types.Converter, config *types.ConversionConfig) (*types.ConversionResult, error) {
	uc.logger.Info("Using streaming conversion for large dataset")

	streamingConverter, ok := converter.(types.StreamingConverter)
	if !ok {
		uc.logger.Warn("Converter does not support streaming, falling back to standard conversion")
		return uc.convertStandard(ctx, converter, config)
	}

	progressCallback := func(progress *types.ConversionProgress) {
		uc.logger.Debug(fmt.Sprintf("Conversion progress: %d/%d records (%.1f%%)",
			progress.ProcessedRecords, progress.TotalRecords,
			float64(progress.ProcessedRecords)/float64(progress.TotalRecords)*100))
	}

	res, err := streamingConverter.ConvertStream(ctx, config, progressCallback)
	if err != nil {
		return nil, err
	}
	uc.maybeValidateOutput(res, config)
	return res, nil
}

func (uc *UniversalConverter) convertStandard(ctx context.Context, converter types.Converter, config *types.ConversionConfig) (*types.ConversionResult, error) {
	uc.logger.Info("Using standard conversion")
	res, err := converter.Convert(ctx, config)
	if err != nil {
		return nil, err
	}
	uc.maybeValidateOutput(res, config)
	return res, nil
}

func (uc *UniversalConverter) maybeValidateOutput(res *types.ConversionResult, config *types.ConversionConfig) {
	if res == nil || !config.ValidateOutput || res.OutputFile == "" {
		return
	}
	eng := val.NewEngine()
	cfg := val.ValidationConfig{
		Level:             val.ValidationLevelStandard,
		Spec:              val.SpecFOCUS12,
		Format:            res.OutputFormat,
		EnableCompliance:  true,
		EnableQuality:     true,
		EnablePerformance: false,
		Quiet:             true,
	}
	out, err := eng.Validate(res.OutputFile, cfg)
	if err != nil || out == nil {
		uc.logger.Warn(fmt.Sprintf("Post-conversion validation failed: %v", err))
		return
	}
	res.DataQualityScore = out.QualityAssessment.Score
	res.ComplianceScore = out.ComplianceValidation.Score
}
