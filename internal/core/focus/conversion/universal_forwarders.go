//go:build never

package conversion

import (
	"context"

	u "local/costscope/internal/core/focus/conversion/universal"
	"local/costscope/internal/core/focus/types"
)

// UniversalConverter is a type alias to preserve backward-compatible imports
type UniversalConverter = u.UniversalConverter

// NewUniversalConverter forwards to the new package constructor
func NewUniversalConverter() *UniversalConverter { return u.NewUniversalConverter() }

// Back-compat method shims for any interfaces that used the concrete type in this package
// Note: These wrappers aren’t strictly necessary due to the type alias above, but keep
// the method set visible to tools that rely on concrete symbols in this package.

func (uc *UniversalConverter) RegisterConverter(provider string, converter types.Converter) error {
	return (*u.UniversalConverter)(uc).RegisterConverter(provider, converter)
}

func (uc *UniversalConverter) Convert(ctx context.Context, config *types.ConversionConfig) (*types.ConversionResult, error) {
	return (*u.UniversalConverter)(uc).Convert(ctx, config)
}

func (uc *UniversalConverter) ConvertBatch(ctx context.Context, configs []*types.ConversionConfig) ([]*types.ConversionResult, error) {
	return (*u.UniversalConverter)(uc).ConvertBatch(ctx, configs)
}

func (uc *UniversalConverter) ValidateInput(ctx context.Context, config *types.ConversionConfig) error {
	return (*u.UniversalConverter)(uc).ValidateInput(ctx, config)
}

func (uc *UniversalConverter) EstimateConversion(ctx context.Context, config *types.ConversionConfig) (*types.ConversionEstimate, error) {
	return (*u.UniversalConverter)(uc).EstimateConversion(ctx, config)
}

func (uc *UniversalConverter) GetSupportedFormats() map[string]*types.SupportedFormats {
	return (*u.UniversalConverter)(uc).GetSupportedFormats()
}

func (uc *UniversalConverter) GetSchema() *types.FocusSchema {
	return (*u.UniversalConverter)(uc).GetSchema()
}
