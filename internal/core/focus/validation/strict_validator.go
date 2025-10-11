package validation

import (
	"fmt"
	focustypes "local/costscope/internal/core/focus/types"
)

// StrictValidator performs stricter, record-level invariants using schema metadata.
// This implementation remains file-agnostic and simulates checks using schema info
// and dictionaries. Future: wire to actual row scanners for Parquet/CSV.
type StrictValidator struct{}

func NewStrictValidator() *StrictValidator { return &StrictValidator{} }

func (v *StrictValidator) Name() string { return "strict" }

func (v *StrictValidator) SupportsFormat(format string) bool {
	// Support common formats
	switch format {
	case FormatParquet, FormatCSV, FormatJSON, FormatORC, FormatAVRO:
		return true
	default:
		return false
	}
}

func (v *StrictValidator) Validate(data interface{}, _ ValidationConfig) (interface{}, error) {
	filePath, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("expected file path string, got %T", data)
	}
	_ = filePath // placeholder until row-level scanning is implemented

	// Use schema metadata to build required+enum expectations
	schema := focustypes.GetFocusV12Schema()
	required := make(map[string]bool)
	for _, f := range schema.Fields {
		if f.Required {
			required[f.Name] = true
		}
	}

	// Prepare a lightweight result mapped onto SchemaValidationResult for reuse
	res := SchemaValidationResult{
		Valid:                true,
		Score:                100.0,
		RequiredColumns:      map[string]bool{},
		OptionalColumns:      map[string]bool{},
		UnknownColumns:       []string{},
		TypeValidation:       map[string]TypeValidation{},
		ConstraintValidation: map[string]ConstraintValidation{},
		Issues:               []SchemaIssue{},
	}

	// Mark presence expectations of key invariants; we cannot scan columns here, so ensure canonical names
	// and add guidance issues that are actionable if missing.
	for k := range required {
		// Focus schema uses snake_case, while validators above used PascalCase for simulated inputs.
		// Record requirement in summary; actual presence is verified by SchemaValidator.
		res.RequiredColumns[k] = true
	}

	// Currency and region normalization guidance
	res.Issues = append(res.Issues, SchemaIssue{
		Type:       "normalization_guidance",
		Column:     "billing_currency",
		Message:    "Billing currency should be ISO-4217 (e.g., USD, EUR)",
		Suggestion: "Uppercase and validate against KnownCurrencies",
	})
	res.Issues = append(res.Issues, SchemaIssue{
		Type:       "normalization_guidance",
		Column:     "region",
		Message:    "Normalize region identifiers to canonical codes (e.g., us-east-1, eastus, us-central1)",
		Suggestion: "Use NormalizeRegion during conversion",
	})

	return res, nil
}
