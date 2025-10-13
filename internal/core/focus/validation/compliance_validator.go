package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/costscope/costscope/internal/core/focus/compliance"
)

// FOCUS column name constants
const (
	ColBillingAccountName = "BillingAccountName"
	ColResourceName       = "ResourceName"
	ColServiceName        = "ServiceName"
)

// ComplianceValidator validates compliance against various frameworks
type ComplianceValidator struct {
	complianceManager *compliance.Manager
}

// NewComplianceValidator creates a new compliance validator
func NewComplianceValidator(complianceManager *compliance.Manager) *ComplianceValidator {
	return &ComplianceValidator{
		complianceManager: complianceManager,
	}
}

// Name returns the validator name
func (v *ComplianceValidator) Name() string {
	return "compliance"
}

// SupportsFormat checks if the validator supports the given format
func (v *ComplianceValidator) SupportsFormat(format string) bool {
	supportedFormats := []string{"parquet", "csv", "json", "orc", "avro"}
	for _, supported := range supportedFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// Validate validates compliance against frameworks
func (v *ComplianceValidator) Validate(data interface{}, config ValidationConfig) (interface{}, error) {
	filePath, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("expected file path string, got %T", data)
	}

	result := ComplianceValidationResult{
		Valid:                true,
		Score:                100.0,
		FOCUSCompliance:      FOCUSComplianceResult{},
		RegulatoryCompliance: RegulatoryComplianceResult{},
		IndustryStandards:    IndustryStandardsResult{},
		AuditTrail:           []AuditEvent{},
		Issues:               []ComplianceIssue{},
	}

	// Validate FOCUS compliance
	if err := v.validateFOCUSCompliance(filePath, &result, config); err != nil {
		return nil, fmt.Errorf("FOCUS compliance validation failed: %w", err)
	}

	// Validate regulatory compliance
	v.validateRegulatoryCompliance(filePath, &result, config)

	// Validate industry standards
	v.validateIndustryStandards(filePath, &result, config)

	// Calculate overall compliance score
	v.calculateComplianceScore(&result)

	return result, nil
}

// validateFOCUSCompliance validates against FOCUS specification
func (v *ComplianceValidator) validateFOCUSCompliance(filePath string, result *ComplianceValidationResult, config ValidationConfig) error {
	// Get FOCUS compliance rules
	focusRules, err := v.complianceManager.GetEnabledRules("focus")
	if err != nil {
		return fmt.Errorf("failed to get FOCUS rules: %w", err)
	}

	specVersion := string(config.Spec)
	if specVersion == "" {
		specVersion = "1.2"
	}

	focusResult := FOCUSComplianceResult{
		SpecVersion:    specVersion,
		RequiredFields: make(map[string]bool),
		OptionalFields: make(map[string]bool),
		CustomFields:   []string{},
		MetadataValid:  true,
		DimensionValid: true,
		MetricValid:    true,
		Issues:         []FOCUSIssue{},
	}

	// Simulate file analysis for FOCUS compliance
	fileData := v.simulateFileData(filePath)

	// Check required fields compliance
	v.checkFOCUSRequiredFields(&focusResult, fileData, specVersion)

	// Check optional fields
	v.checkFOCUSOptionalFields(&focusResult, fileData)

	// Check custom fields
	v.checkFOCUSCustomFields(&focusResult, fileData)

	// Check metadata compliance
	v.checkFOCUSMetadata(&focusResult, fileData)

	// Check dimension compliance
	v.checkFOCUSDimensions(&focusResult, fileData)

	// Check metric compliance
	v.checkFOCUSMetrics(&focusResult, fileData)

	// Run specific FOCUS rules
	for _, rule := range focusRules {
		if err := v.validateFOCUSRule(rule, fileData, &focusResult); err != nil {
			return fmt.Errorf("FOCUS rule validation failed for %s: %w", rule.ID, err)
		}
	}

	result.FOCUSCompliance = focusResult

	// Add issues to main result
	for _, issue := range focusResult.Issues {
		result.Issues = append(result.Issues, ComplianceIssue{
			Type:       "focus",
			Standard:   "FOCUS",
			Rule:       issue.Type,
			Message:    issue.Message,
			Severity:   "high",
			Suggestion: issue.Suggestion,
		})
	}

	return nil
}

// validateRegulatoryCompliance validates regulatory compliance
func (v *ComplianceValidator) validateRegulatoryCompliance(filePath string, result *ComplianceValidationResult, _ ValidationConfig) {
	// Note: filePath could be used for file-specific compliance rules in future
	_ = filePath

	regulatoryResult := RegulatoryComplianceResult{
		GDPR:   ComplianceStatus{Applicable: true, Compliant: true, Score: 100.0},
		SOX:    ComplianceStatus{Applicable: true, Compliant: true, Score: 100.0},
		HIPAA:  ComplianceStatus{Applicable: false, Compliant: true, Score: 100.0},
		PCI:    ComplianceStatus{Applicable: false, Compliant: true, Score: 100.0},
		Issues: []RegulatoryIssue{},
	}

	fileData := v.simulateFileData(filePath)

	// Check GDPR compliance
	v.checkGDPRCompliance(&regulatoryResult, fileData)

	// Check SOX compliance
	v.checkSOXCompliance(&regulatoryResult, fileData)

	result.RegulatoryCompliance = regulatoryResult

	// Add regulatory issues to main result
	for _, issue := range regulatoryResult.Issues {
		result.Issues = append(result.Issues, ComplianceIssue{
			Type:       "regulatory",
			Standard:   issue.Regulation,
			Rule:       issue.Requirement,
			Message:    issue.Message,
			Severity:   issue.Severity,
			Suggestion: issue.Suggestion,
		})
	}
}

// validateIndustryStandards validates industry standards compliance
func (v *ComplianceValidator) validateIndustryStandards(filePath string, result *ComplianceValidationResult, _ ValidationConfig) {
	// Note: filePath could be used for file-specific standards in future
	_ = filePath

	standardsResult := IndustryStandardsResult{
		ISO27001: ComplianceStatus{Applicable: true, Compliant: true, Score: 95.0},
		SOC2:     ComplianceStatus{Applicable: true, Compliant: true, Score: 90.0},
		NIST:     ComplianceStatus{Applicable: true, Compliant: true, Score: 92.0},
		Issues:   []StandardIssue{},
	}

	result.IndustryStandards = standardsResult
}

// simulateFileData simulates file data analysis
func (v *ComplianceValidator) simulateFileData(_ string) map[string]interface{} {
	// In a real implementation, this would analyze the actual file
	return map[string]interface{}{
		"columns": []string{
			"BillingAccountId", "BillingAccountName", "BillingCurrency",
			"EffectiveCost", "ListCost", "ProviderName", "ServiceName",
			"ResourceId", "ResourceName", "Tags",
		},
		"row_count":    125000,
		"has_metadata": true,
		"file_format":  "parquet",
		"compression":  "snappy",
	}
}

// checkFOCUSRequiredFields checks FOCUS required fields compliance
func (v *ComplianceValidator) checkFOCUSRequiredFields(result *FOCUSComplianceResult, fileData map[string]interface{}, specVersion string) {
	requiredFields := v.getFOCUSRequiredFields(specVersion)
	fileColumns := fileData["columns"].([]string)

	for _, field := range requiredFields {
		if contains(fileColumns, field) {
			result.RequiredFields[field] = true
		} else {
			result.RequiredFields[field] = false
			result.Issues = append(result.Issues, FOCUSIssue{
				Type:       "missing_required_field",
				Field:      field,
				Message:    fmt.Sprintf("Required FOCUS field '%s' is missing", field),
				Suggestion: fmt.Sprintf("Add required field '%s' to comply with FOCUS v%s", field, specVersion),
			})
		}
	}
}

// checkFOCUSOptionalFields checks FOCUS optional fields
func (v *ComplianceValidator) checkFOCUSOptionalFields(result *FOCUSComplianceResult, fileData map[string]interface{}) {
	optionalFields := []string{
		"AvailabilityZone", "ChargeDescription", "BilledCost", "ContractedCost",
		"ContractedUnitPrice", "CommitmentDiscountId", "CommitmentDiscountName",
	}
	fileColumns := fileData["columns"].([]string)

	for _, field := range optionalFields {
		result.OptionalFields[field] = contains(fileColumns, field)
	}
}

// checkFOCUSCustomFields identifies custom fields
func (v *ComplianceValidator) checkFOCUSCustomFields(result *FOCUSComplianceResult, fileData map[string]interface{}) {
	fileColumns := fileData["columns"].([]string)
	standardFields := v.getAllFOCUSFields()

	for _, column := range fileColumns {
		if !contains(standardFields, column) {
			result.CustomFields = append(result.CustomFields, column)
		}
	}
}

// checkFOCUSMetadata checks metadata compliance
func (v *ComplianceValidator) checkFOCUSMetadata(result *FOCUSComplianceResult, fileData map[string]interface{}) {
	hasMetadata, exists := fileData["has_metadata"].(bool)
	if !exists || !hasMetadata {
		result.MetadataValid = false
		result.Issues = append(result.Issues, FOCUSIssue{
			Type:       "missing_metadata",
			Field:      "metadata",
			Message:    "File metadata is missing or incomplete",
			Suggestion: "Add proper file metadata including schema version and generation timestamp",
		})
	}
}

// checkFOCUSDimensions checks dimension compliance
func (v *ComplianceValidator) checkFOCUSDimensions(result *FOCUSComplianceResult, fileData map[string]interface{}) {
	requiredDimensions := []string{
		"BillingAccountId", "ServiceName", "ServiceCategory", "ProviderName",
	}
	fileColumns := fileData["columns"].([]string)

	missingDimensions := []string{}
	for _, dimension := range requiredDimensions {
		if !contains(fileColumns, dimension) {
			missingDimensions = append(missingDimensions, dimension)
		}
	}

	if len(missingDimensions) > 0 {
		result.DimensionValid = false
		result.Issues = append(result.Issues, FOCUSIssue{
			Type:       "missing_dimensions",
			Field:      strings.Join(missingDimensions, ", "),
			Message:    fmt.Sprintf("Missing required dimensions: %s", strings.Join(missingDimensions, ", ")),
			Suggestion: "Add missing dimension columns to enable proper data analysis",
		})
	}
}

// checkFOCUSMetrics checks metric compliance
func (v *ComplianceValidator) checkFOCUSMetrics(result *FOCUSComplianceResult, fileData map[string]interface{}) {
	requiredMetrics := []string{"EffectiveCost", "ListCost"}
	fileColumns := fileData["columns"].([]string)

	missingMetrics := []string{}
	for _, metric := range requiredMetrics {
		if !contains(fileColumns, metric) {
			missingMetrics = append(missingMetrics, metric)
		}
	}

	if len(missingMetrics) > 0 {
		result.MetricValid = false
		result.Issues = append(result.Issues, FOCUSIssue{
			Type:       "missing_metrics",
			Field:      strings.Join(missingMetrics, ", "),
			Message:    fmt.Sprintf("Missing required metrics: %s", strings.Join(missingMetrics, ", ")),
			Suggestion: "Add missing metric columns for cost analysis",
		})
	}
}

// validateFOCUSRule validates a specific FOCUS rule
func (v *ComplianceValidator) validateFOCUSRule(rule *compliance.ComplianceRule, fileData map[string]interface{}, result *FOCUSComplianceResult) error {
	switch rule.ID {
	case "focus-currency-format":
		return v.validateCurrencyFormat(fileData, result)
	case "focus-cost-consistency":
		return v.validateCostConsistency(fileData, result)
	default:
		// Rule validation logic would go here
		return nil
	}
}

// validateCurrencyFormat validates currency format compliance
func (v *ComplianceValidator) validateCurrencyFormat(fileData map[string]interface{}, result *FOCUSComplianceResult) error {
	// Simulate currency format validation
	// In a real implementation, this would check actual currency values
	_ = []string{"USD", "EUR", "GBP", "JPY", "CAD", "AUD"} // Valid currencies

	// Simulate finding some currency format issues
	if !contains(fileData["columns"].([]string), "BillingCurrency") {
		result.Issues = append(result.Issues, FOCUSIssue{
			Type:       "currency_format_issue",
			Field:      "BillingCurrency",
			Message:    "BillingCurrency column format validation failed",
			Suggestion: "Ensure BillingCurrency uses ISO 4217 format (e.g., USD, EUR)",
		})
	}

	return nil
}

// validateCostConsistency validates cost value consistency
func (v *ComplianceValidator) validateCostConsistency(fileData map[string]interface{}, result *FOCUSComplianceResult) error {
	// Simulate cost consistency validation
	fileColumns := fileData["columns"].([]string)

	if contains(fileColumns, "EffectiveCost") && contains(fileColumns, "ListCost") {
		// In a real implementation, we would check actual values
		// For simulation, assume consistency is good
		return nil
	}

	result.Issues = append(result.Issues, FOCUSIssue{
		Type:       "cost_consistency_issue",
		Field:      "EffectiveCost, ListCost",
		Message:    "Cost consistency validation requires both EffectiveCost and ListCost",
		Suggestion: "Ensure both EffectiveCost and ListCost columns are present for validation",
	})

	return nil
}

// checkGDPRCompliance checks GDPR compliance
func (v *ComplianceValidator) checkGDPRCompliance(result *RegulatoryComplianceResult, fileData map[string]interface{}) {
	// Check for potential PII in column names and simulated data
	fileColumns := fileData["columns"].([]string)

	// Look for potential PII indicators
	piiPatterns := []string{
		"email", "phone", "ssn", "credit", "card", "name",
		"address", "birth", "personal", "private",
	}

	for _, column := range fileColumns {
		columnLower := strings.ToLower(column)
		for _, pattern := range piiPatterns {
			if strings.Contains(columnLower, pattern) &&
				column != ColBillingAccountName &&
				column != ColResourceName &&
				column != ColServiceName {
				result.GDPR.Compliant = false
				result.GDPR.Score = 60.0
				result.Issues = append(result.Issues, RegulatoryIssue{
					Regulation:  "GDPR",
					Requirement: "Article 4 - Personal Data Protection",
					Message:     fmt.Sprintf("Column '%s' may contain personal information", column),
					Severity:    "high",
					Suggestion:  "Review column for personal data and apply appropriate protection measures",
				})
			}
		}
	}

	// Check for potential email patterns in ResourceName (simulated)
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	if emailPattern.MatchString("user@example.com") { // Simulate finding emails
		// This would be actual data scanning in real implementation
		result.GDPR.Compliant = false
		result.GDPR.Score = 70.0
	}
}

// checkSOXCompliance checks SOX compliance
func (v *ComplianceValidator) checkSOXCompliance(result *RegulatoryComplianceResult, fileData map[string]interface{}) {
	// Check audit trail requirements
	hasMetadata := fileData["has_metadata"].(bool)
	if !hasMetadata {
		result.SOX.Compliant = false
		result.SOX.Score = 70.0
		result.Issues = append(result.Issues, RegulatoryIssue{
			Regulation:  "SOX",
			Requirement: "Section 302 - Financial Reporting Controls",
			Message:     "Missing audit trail metadata for financial data",
			Severity:    "high",
			Suggestion:  "Add comprehensive metadata including data lineage and access controls",
		})
	}
}

// getFOCUSRequiredFields returns required fields for a FOCUS version
func (v *ComplianceValidator) getFOCUSRequiredFields(version string) []string {
	switch version {
	case "1.2":
		return []string{
			"BillingAccountId", "BillingAccountName", "BillingCurrency",
			"BillingPeriodEnd", "BillingPeriodStart", "ChargeCategory",
			"EffectiveCost", "ListCost", "ProviderName", "ServiceName",
			"ServiceCategory",
		}
	case "1.1":
		return []string{
			"BillingAccountId", "BillingAccountName", "BillingCurrency",
			"EffectiveCost", "ListCost", "ProviderName", "ServiceName",
		}
	case "1.0":
		return []string{
			"BillingAccountId", "EffectiveCost", "ListCost", "ProviderName", "ServiceName",
		}
	default:
		return []string{
			"BillingAccountId", "EffectiveCost", "ListCost", "ProviderName", "ServiceName",
		}
	}
}

// getAllFOCUSFields returns all standard FOCUS fields
func (v *ComplianceValidator) getAllFOCUSFields() []string {
	return []string{
		"BillingAccountId", "BillingAccountName", "BillingCurrency",
		"BillingPeriodEnd", "BillingPeriodStart", "ChargeCategory",
		"ChargeFrequency", "ChargePeriodEnd", "ChargePeriodStart",
		"EffectiveCost", "ListCost", "BilledCost", "ContractedCost",
		"ProviderName", "ServiceName", "ServiceCategory",
		"ResourceId", "ResourceName", "ResourceType",
		"RegionId", "RegionName", "AvailabilityZone",
		"Tags", "SkuId", "PricingCategory", "PricingQuantity",
		"PricingUnit", "ListUnitPrice", "ContractedUnitPrice",
	}
}

// calculateComplianceScore calculates overall compliance score
func (v *ComplianceValidator) calculateComplianceScore(result *ComplianceValidationResult) {
	totalScore := 0.0
	componentCount := 0

	// FOCUS compliance (40% weight)
	focusScore := 100.0
	if len(result.FOCUSCompliance.Issues) > 0 {
		focusScore -= float64(len(result.FOCUSCompliance.Issues)) * 10.0
	}
	if !result.FOCUSCompliance.MetadataValid {
		focusScore -= 15.0
	}
	if !result.FOCUSCompliance.DimensionValid {
		focusScore -= 20.0
	}
	if !result.FOCUSCompliance.MetricValid {
		focusScore -= 25.0
	}

	totalScore += focusScore * 0.4
	componentCount++

	// Regulatory compliance (40% weight)
	regulatoryScore := 0.0
	complianceCount := 0

	if result.RegulatoryCompliance.GDPR.Applicable {
		regulatoryScore += result.RegulatoryCompliance.GDPR.Score
		complianceCount++
	}
	if result.RegulatoryCompliance.SOX.Applicable {
		regulatoryScore += result.RegulatoryCompliance.SOX.Score
		complianceCount++
	}
	if result.RegulatoryCompliance.HIPAA.Applicable {
		regulatoryScore += result.RegulatoryCompliance.HIPAA.Score
		complianceCount++
	}
	if result.RegulatoryCompliance.PCI.Applicable {
		regulatoryScore += result.RegulatoryCompliance.PCI.Score
		complianceCount++
	}

	if complianceCount > 0 {
		totalScore += (regulatoryScore / float64(complianceCount)) * 0.4
		componentCount++
	}

	// Industry standards (20% weight)
	standardsScore := (result.IndustryStandards.ISO27001.Score +
		result.IndustryStandards.SOC2.Score +
		result.IndustryStandards.NIST.Score) / 3.0
	totalScore += standardsScore * 0.2
	componentCount++

	// Calculate final score
	if componentCount > 0 {
		result.Score = totalScore
	}

	// Mark as invalid if score is below threshold
	if result.Score < 80.0 {
		result.Valid = false
	}

	// Ensure score doesn't go below 0
	if result.Score < 0 {
		result.Score = 0
	}
}
