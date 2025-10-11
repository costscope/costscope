package compliance

import (
	"fmt"
	"time"
)

// ComplianceFramework represents a compliance framework
type ComplianceFramework struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Rules       []ComplianceRule  `json:"rules"`
	Categories  []string          `json:"categories"`
	Metadata    FrameworkMetadata `json:"metadata"`
}

// ComplianceRule represents a compliance rule
type ComplianceRule struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Category     string         `json:"category"`
	Severity     string         `json:"severity"`
	Framework    string         `json:"framework"`
	Control      string         `json:"control"`
	Requirements []string       `json:"requirements"`
	Validation   RuleValidation `json:"validation"`
	Enabled      bool           `json:"enabled"`
	Metadata     RuleMetadata   `json:"metadata"`
}

// RuleValidation defines how to validate a rule
type RuleValidation struct {
	Type       string                 `json:"type"`
	Field      string                 `json:"field,omitempty"`
	Pattern    string                 `json:"pattern,omitempty"`
	Values     []interface{}          `json:"values,omitempty"`
	Range      *ValueRange            `json:"range,omitempty"`
	Script     string                 `json:"script,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ValueRange defines a numeric range
type ValueRange struct {
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Exclusive bool     `json:"exclusive,omitempty"`
}

// FrameworkMetadata contains framework metadata
type FrameworkMetadata struct {
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedBy  string    `json:"created_by"`
	Source     string    `json:"source"`
	Website    string    `json:"website"`
	Applicable []string  `json:"applicable"`
}

// RuleMetadata contains rule metadata
type RuleMetadata struct {
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedBy  string    `json:"created_by"`
	Tags       []string  `json:"tags"`
	References []string  `json:"references"`
}

// Manager manages compliance frameworks and rules
type Manager struct {
	frameworks map[string]*ComplianceFramework
	rules      map[string]*ComplianceRule
}

// NewManager creates a new compliance manager
func NewManager() *Manager {
	manager := &Manager{
		frameworks: make(map[string]*ComplianceFramework),
		rules:      make(map[string]*ComplianceRule),
	}

	// Load default compliance frameworks
	manager.loadDefaultFrameworks()

	return manager
}

// loadDefaultFrameworks loads built-in compliance frameworks
func (m *Manager) loadDefaultFrameworks() {
	now := time.Now()

	// FOCUS Framework
	focusFramework := &ComplianceFramework{
		ID:          "focus",
		Name:        "FinOps Open Cost and Usage Specification",
		Version:     "1.2",
		Description: "FOCUS compliance framework for cloud cost and usage data",
		Categories:  []string{"data_structure", "metadata", "pricing", "billing"},
		Metadata: FrameworkMetadata{
			CreatedAt:  now,
			UpdatedAt:  now,
			CreatedBy:  "system",
			Source:     "FOCUS Specification",
			Website:    "https://github.com/FinOps-Open-Cost-and-Usage-Spec/FOCUS_Spec",
			Applicable: []string{"cloud_billing", "cost_data", "usage_data"},
		},
		Rules: []ComplianceRule{
			{
				ID:          "focus-required-columns",
				Name:        "Required Columns Present",
				Description: "All FOCUS required columns must be present",
				Category:    "data_structure",
				Severity:    "critical",
				Framework:   "focus",
				Control:     "FOCUS-DS-001",
				Requirements: []string{
					"All required columns as per FOCUS specification must be present",
					"Column names must match exactly",
					"Column types must be compatible",
				},
				Validation: RuleValidation{
					Type: "schema_validation",
					Parameters: map[string]interface{}{
						"check_required_columns": true,
						"strict_naming":          true,
					},
				},
				Enabled: true,
				Metadata: RuleMetadata{
					CreatedAt:  now,
					UpdatedAt:  now,
					CreatedBy:  "system",
					Tags:       []string{"schema", "required", "structure"},
					References: []string{"FOCUS Spec v1.2 Section 3.1"},
				},
			},
			{
				ID:          "focus-data-types",
				Name:        "Data Type Compliance",
				Description: "Column data types must conform to FOCUS specification",
				Category:    "data_structure",
				Severity:    "high",
				Framework:   "focus",
				Control:     "FOCUS-DS-002",
				Requirements: []string{
					"Numeric columns must contain valid numbers",
					"Date columns must be in ISO 8601 format",
					"String columns must be properly encoded",
				},
				Validation: RuleValidation{
					Type: "data_type_validation",
					Parameters: map[string]interface{}{
						"strict_types":     true,
						"validate_formats": true,
					},
				},
				Enabled: true,
				Metadata: RuleMetadata{
					CreatedAt:  now,
					UpdatedAt:  now,
					CreatedBy:  "system",
					Tags:       []string{"types", "format", "validation"},
					References: []string{"FOCUS Spec v1.2 Section 3.2"},
				},
			},
			{
				ID:          "focus-currency-format",
				Name:        "Currency Format Compliance",
				Description: "BillingCurrency must be ISO 4217 compliant",
				Category:    "metadata",
				Severity:    "high",
				Framework:   "focus",
				Control:     "FOCUS-MT-001",
				Requirements: []string{
					"BillingCurrency must be 3-character ISO 4217 code",
					"Currency codes must be uppercase",
					"Must be valid currency codes",
				},
				Validation: RuleValidation{
					Type:    "field_validation",
					Field:   "BillingCurrency",
					Pattern: "^[A-Z]{3}$",
				},
				Enabled: true,
				Metadata: RuleMetadata{
					CreatedAt:  now,
					UpdatedAt:  now,
					CreatedBy:  "system",
					Tags:       []string{"currency", "format", "iso4217"},
					References: []string{"FOCUS Spec v1.2 Section 4.1"},
				},
			},
			{
				ID:          "focus-cost-consistency",
				Name:        "Cost Value Consistency",
				Description: "Cost values must be consistent and non-negative",
				Category:    "pricing",
				Severity:    "medium",
				Framework:   "focus",
				Control:     "FOCUS-PR-001",
				Requirements: []string{
					"EffectiveCost must be non-negative",
					"ListCost must be non-negative",
					"EffectiveCost should not exceed ListCost significantly",
				},
				Validation: RuleValidation{
					Type: "cost_validation",
					Parameters: map[string]interface{}{
						"check_non_negative": true,
						"check_consistency":  true,
						"tolerance_percent":  10.0,
					},
				},
				Enabled: true,
				Metadata: RuleMetadata{
					CreatedAt:  now,
					UpdatedAt:  now,
					CreatedBy:  "system",
					Tags:       []string{"cost", "consistency", "validation"},
					References: []string{"FOCUS Spec v1.2 Section 5.1"},
				},
			},
		},
	}

	m.frameworks["focus"] = focusFramework

	// GDPR Framework
	gdprFramework := &ComplianceFramework{
		ID:          "gdpr",
		Name:        "General Data Protection Regulation",
		Version:     "2018",
		Description: "GDPR compliance framework for data protection",
		Categories:  []string{"data_protection", "privacy", "consent", "rights"},
		Metadata: FrameworkMetadata{
			CreatedAt:  now,
			UpdatedAt:  now,
			CreatedBy:  "system",
			Source:     "EU GDPR",
			Website:    "https://gdpr-info.eu/",
			Applicable: []string{"personal_data", "eu_data", "privacy"},
		},
		Rules: []ComplianceRule{
			{
				ID:          "gdpr-pii-detection",
				Name:        "Personal Information Detection",
				Description: "Detect potential personal information in cost data",
				Category:    "data_protection",
				Severity:    "critical",
				Framework:   "gdpr",
				Control:     "GDPR-ART-4",
				Requirements: []string{
					"No direct personal identifiers in cost data",
					"Resource names should not contain personal information",
					"Tag values should not expose personal data",
				},
				Validation: RuleValidation{
					Type: "pii_detection",
					Parameters: map[string]interface{}{
						"check_fields": []string{"ResourceName", "Tags", "BillingAccountName"},
						"patterns": []string{
							"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
							"\\b\\d{3}-\\d{2}-\\d{4}\\b",
							"\\b\\d{4}[\\s-]?\\d{4}[\\s-]?\\d{4}[\\s-]?\\d{4}\\b",
						},
					},
				},
				Enabled: true,
				Metadata: RuleMetadata{
					CreatedAt:  now,
					UpdatedAt:  now,
					CreatedBy:  "system",
					Tags:       []string{"pii", "privacy", "detection"},
					References: []string{"GDPR Article 4"},
				},
			},
		},
	}

	m.frameworks["gdpr"] = gdprFramework

	// SOX Framework
	soxFramework := &ComplianceFramework{
		ID:          "sox",
		Name:        "Sarbanes-Oxley Act",
		Version:     "2002",
		Description: "SOX compliance framework for financial reporting",
		Categories:  []string{"financial_reporting", "audit_trail", "controls"},
		Metadata: FrameworkMetadata{
			CreatedAt:  now,
			UpdatedAt:  now,
			CreatedBy:  "system",
			Source:     "US SOX Act",
			Website:    "https://www.sec.gov/about/laws/soa2002.pdf",
			Applicable: []string{"financial_data", "audit", "public_companies"},
		},
		Rules: []ComplianceRule{
			{
				ID:          "sox-audit-trail",
				Name:        "Audit Trail Requirements",
				Description: "Financial data must have proper audit trail",
				Category:    "audit_trail",
				Severity:    "high",
				Framework:   "sox",
				Control:     "SOX-302",
				Requirements: []string{
					"Data lineage must be traceable",
					"Changes must be logged",
					"Access must be controlled and logged",
				},
				Validation: RuleValidation{
					Type: "audit_validation",
					Parameters: map[string]interface{}{
						"require_metadata": true,
						"check_lineage":    true,
					},
				},
				Enabled: true,
				Metadata: RuleMetadata{
					CreatedAt:  now,
					UpdatedAt:  now,
					CreatedBy:  "system",
					Tags:       []string{"audit", "trail", "controls"},
					References: []string{"SOX Section 302"},
				},
			},
		},
	}

	m.frameworks["sox"] = soxFramework

	// Register all rules
	for _, framework := range []*ComplianceFramework{focusFramework, gdprFramework, soxFramework} {
		for _, rule := range framework.Rules {
			m.rules[rule.ID] = &rule
		}
	}
}

// GetFramework returns a compliance framework by ID
func (m *Manager) GetFramework(id string) (*ComplianceFramework, error) {
	framework, exists := m.frameworks[id]
	if !exists {
		return nil, fmt.Errorf("compliance framework not found: %s", id)
	}
	return framework, nil
}

// GetRule returns a compliance rule by ID
func (m *Manager) GetRule(id string) (*ComplianceRule, error) {
	rule, exists := m.rules[id]
	if !exists {
		return nil, fmt.Errorf("compliance rule not found: %s", id)
	}
	return rule, nil
}

// GetAvailableFrameworks returns list of available framework IDs
func (m *Manager) GetAvailableFrameworks() []string {
	var ids []string
	for id := range m.frameworks {
		ids = append(ids, id)
	}
	return ids
}

// GetRulesForFramework returns all rules for a specific framework
func (m *Manager) GetRulesForFramework(frameworkID string) ([]*ComplianceRule, error) {
	var rules []*ComplianceRule

	for _, rule := range m.rules {
		if rule.Framework == frameworkID {
			rules = append(rules, rule)
		}
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("no rules found for framework: %s", frameworkID)
	}

	return rules, nil
}

// GetEnabledRules returns all enabled rules for a framework
func (m *Manager) GetEnabledRules(frameworkID string) ([]*ComplianceRule, error) {
	allRules, err := m.GetRulesForFramework(frameworkID)
	if err != nil {
		return nil, err
	}

	var enabledRules []*ComplianceRule
	for _, rule := range allRules {
		if rule.Enabled {
			enabledRules = append(enabledRules, rule)
		}
	}

	return enabledRules, nil
}

// ValidateRule validates data against a specific rule
func (m *Manager) ValidateRule(ruleID string, data interface{}) (bool, []string, error) {
	rule, err := m.GetRule(ruleID)
	if err != nil {
		return false, nil, err
	}

	if !rule.Enabled {
		return true, nil, nil
	}

	// This would contain the actual validation logic
	// For now, return placeholder results
	issues := []string{}
	valid := true

	switch rule.Validation.Type {
	case "schema_validation":
		// Placeholder for schema validation
	case "data_type_validation":
		// Placeholder for data type validation
	case "field_validation":
		// Placeholder for field validation
	case "cost_validation":
		// Placeholder for cost validation
	case "pii_detection":
		// Placeholder for PII detection
	case "audit_validation":
		// Placeholder for audit validation
	default:
		valid = false
		issues = append(issues, fmt.Sprintf("Unknown validation type: %s", rule.Validation.Type))
	}

	return valid, issues, nil
}

// RegisterFramework registers a custom compliance framework
func (m *Manager) RegisterFramework(framework *ComplianceFramework) error {
	if framework == nil {
		return fmt.Errorf("framework cannot be nil")
	}
	if framework.ID == "" {
		return fmt.Errorf("framework ID cannot be empty")
	}

	m.frameworks[framework.ID] = framework

	// Register all rules from the framework
	for _, rule := range framework.Rules {
		rule.Framework = framework.ID
		m.rules[rule.ID] = &rule
	}

	return nil
}

// EnableRule enables a compliance rule
func (m *Manager) EnableRule(ruleID string) error {
	rule, exists := m.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	rule.Enabled = true
	return nil
}

// DisableRule disables a compliance rule
func (m *Manager) DisableRule(ruleID string) error {
	rule, exists := m.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	rule.Enabled = false
	return nil
}
