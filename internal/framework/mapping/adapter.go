package mapping

import (
	ftypes "local/costscope/internal/core/focus/types"
	"strings"
)

// Adapter provides a bridge from existing MappingRules to the new FieldMapper.
// This is a placeholder for incremental migration and is not wired yet.
type Adapter interface {
	// FromRules builds a FieldMappingConfig from MappingRules
	FromRules(rules *ftypes.MappingRules) *FieldMappingConfig
}

type defaultAdapter struct{}

// NewAdapter returns a default adapter implementation.
func NewAdapter() Adapter { return &defaultAdapter{} }

func (a *defaultAdapter) FromRules(rules *ftypes.MappingRules) *FieldMappingConfig {
	if rules == nil {
		return &FieldMappingConfig{}
	}
	cfg := &FieldMappingConfig{
		ProviderName:    rules.Provider,
		FieldMappings:   make(map[string]FieldMapping),
		EnumMappings:    make(map[string]map[string]string),
		DefaultValues:   make(map[string]interface{}),
		ValidationRules: make(map[string]ValidationRule),
		TimeFormats:     make(map[string]string),
		CustomMappings:  make(map[string]CustomMappingFunction),
	}
	// Translate field maps; infer FocusRecord fields and basic types
	for tgt, fm := range rules.FieldMaps {
		focusField := snakeToCamelExported(tgt)
		cfg.FieldMappings[focusField] = FieldMapping{
			SourceField: fm.SourceField,
			TargetField: focusField,
			FieldType:   inferFieldType(focusField),
			IsRequired:  fm.Required,
			Transform:   firstTransform(fm.Transformations),
		}
	}
	// NOTE: enum mappings and validations can be filled by provider-specific code later.
	return cfg
}

func firstTransform(ts []string) string {
	if len(ts) == 0 {
		return ""
	}
	return ts[0]
}

// snakeToCamelExported converts snake_case to ExportedCamelCase
func snakeToCamelExported(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		p := parts[i]
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

// inferFieldType provides a minimal heuristic for known FocusRecord field types
func inferFieldType(field string) FieldType {
	switch field {
	// required numeric
	case "EffectiveCost", "ListCost", "ListUnitPrice", "PricingQuantity", "UsageQuantity":
		return FieldTypeFloat64
	// optional numeric
	case "BilledCost", "ConsumedQuantity", "ContractedCost", "ContractedUnitPrice":
		return FieldTypeFloat64
	// timestamps
	case "BillingPeriodStart", "BillingPeriodEnd", "ChargePeriodStart", "ChargePeriodEnd":
		return FieldTypeTime
	default:
		return FieldTypeString
	}
}
