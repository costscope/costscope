package schemas

import (
	"fmt"
)

// FOCUSSchema represents a FOCUS specification schema
type FOCUSSchema struct {
	Version         string                `json:"version"`
	RequiredColumns map[string]ColumnSpec `json:"required_columns"`
	OptionalColumns map[string]ColumnSpec `json:"optional_columns"`
	Dimensions      []string              `json:"dimensions"`
	Metrics         []string              `json:"metrics"`
	Metadata        SchemaMetadata        `json:"metadata"`
}

// ColumnSpec defines the specification for a column
type ColumnSpec struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Nullable    bool          `json:"nullable"`
	Format      string        `json:"format,omitempty"`
	Constraints []Constraint  `json:"constraints,omitempty"`
	Examples    []interface{} `json:"examples,omitempty"`
}

// Constraint defines a validation constraint
type Constraint struct {
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
}

// SchemaMetadata contains schema metadata
type SchemaMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	LastUpdated string `json:"last_updated"`
	Source      string `json:"source"`
}

// Manager manages FOCUS schemas
type Manager struct {
	schemas map[string]*FOCUSSchema
}

// NewManager creates a new schema manager
func NewManager() *Manager {
	manager := &Manager{
		schemas: make(map[string]*FOCUSSchema),
	}

	// Load default FOCUS schemas
	manager.loadDefaultSchemas()

	return manager
}

// loadDefaultSchemas loads built-in FOCUS schemas
func (m *Manager) loadDefaultSchemas() {
	// FOCUS v1.2 schema
	focus12 := &FOCUSSchema{
		Version: "1.2",
		RequiredColumns: map[string]ColumnSpec{
			"BillingAccountId": {
				Name:        "BillingAccountId",
				Type:        "string",
				Description: "Unique identifier for the billing account",
				Required:    true,
				Nullable:    false,
			},
			"BillingAccountName": {
				Name:        "BillingAccountName",
				Type:        "string",
				Description: "Display name for the billing account",
				Required:    true,
				Nullable:    false,
			},
			"BillingCurrency": {
				Name:        "BillingCurrency",
				Type:        "string",
				Description: "Currency used for billing",
				Required:    true,
				Nullable:    false,
				Format:      "ISO 4217",
				Constraints: []Constraint{
					{Type: "length", Value: 3, Description: "Must be 3-character ISO code"},
				},
			},
			"BillingPeriodEnd": {
				Name:        "BillingPeriodEnd",
				Type:        "timestamp",
				Description: "End date of the billing period",
				Required:    true,
				Nullable:    false,
				Format:      "ISO 8601",
			},
			"BillingPeriodStart": {
				Name:        "BillingPeriodStart",
				Type:        "timestamp",
				Description: "Start date of the billing period",
				Required:    true,
				Nullable:    false,
				Format:      "ISO 8601",
			},
			"ChargeCategory": {
				Name:        "ChargeCategory",
				Type:        "string",
				Description: "High-level categorization of the charge",
				Required:    true,
				Nullable:    false,
				Constraints: []Constraint{
					{Type: "enum", Value: []string{"Usage", "Tax", "Adjustment", "Credit"},
						Description: "Must be one of the predefined categories"},
				},
			},
			"ChargeFrequency": {
				Name:        "ChargeFrequency",
				Type:        "string",
				Description: "Frequency of the charge",
				Required:    true,
				Nullable:    false,
			},
			"ChargePeriodEnd": {
				Name:        "ChargePeriodEnd",
				Type:        "timestamp",
				Description: "End date of the charge period",
				Required:    true,
				Nullable:    false,
				Format:      "ISO 8601",
			},
			"ChargePeriodStart": {
				Name:        "ChargePeriodStart",
				Type:        "timestamp",
				Description: "Start date of the charge period",
				Required:    true,
				Nullable:    false,
				Format:      "ISO 8601",
			},
			"CommitmentDiscountId": {
				Name:        "CommitmentDiscountId",
				Type:        "string",
				Description: "Unique identifier for commitment discount",
				Required:    true,
				Nullable:    true,
			},
			"CommitmentDiscountName": {
				Name:        "CommitmentDiscountName",
				Type:        "string",
				Description: "Display name for commitment discount",
				Required:    true,
				Nullable:    true,
			},
			"CommitmentDiscountType": {
				Name:        "CommitmentDiscountType",
				Type:        "string",
				Description: "Type of commitment discount",
				Required:    true,
				Nullable:    true,
			},
			"EffectiveCost": {
				Name:        "EffectiveCost",
				Type:        "decimal",
				Description: "Cost after applying all discounts and adjustments",
				Required:    true,
				Nullable:    false,
				Constraints: []Constraint{
					{Type: "min", Value: 0, Description: "Must be non-negative"},
				},
			},
			"InvoiceIssuerName": {
				Name:        "InvoiceIssuerName",
				Type:        "string",
				Description: "Name of the entity issuing the invoice",
				Required:    true,
				Nullable:    false,
			},
			"ListCost": {
				Name:        "ListCost",
				Type:        "decimal",
				Description: "Cost at list/published rates",
				Required:    true,
				Nullable:    false,
				Constraints: []Constraint{
					{Type: "min", Value: 0, Description: "Must be non-negative"},
				},
			},
			"ListUnitPrice": {
				Name:        "ListUnitPrice",
				Type:        "decimal",
				Description: "Unit price at list rates",
				Required:    true,
				Nullable:    true,
			},
			"PricingCategory": {
				Name:        "PricingCategory",
				Type:        "string",
				Description: "Category of pricing model",
				Required:    true,
				Nullable:    false,
			},
			"PricingQuantity": {
				Name:        "PricingQuantity",
				Type:        "decimal",
				Description: "Quantity of units for pricing",
				Required:    true,
				Nullable:    true,
				Constraints: []Constraint{
					{Type: "min", Value: 0, Description: "Must be non-negative"},
				},
			},
			"PricingUnit": {
				Name:        "PricingUnit",
				Type:        "string",
				Description: "Unit of measurement for pricing",
				Required:    true,
				Nullable:    true,
			},
			"ProviderName": {
				Name:        "ProviderName",
				Type:        "string",
				Description: "Name of the cloud provider",
				Required:    true,
				Nullable:    false,
			},
			"PublisherName": {
				Name:        "PublisherName",
				Type:        "string",
				Description: "Name of the service publisher",
				Required:    true,
				Nullable:    false,
			},
			"RegionId": {
				Name:        "RegionId",
				Type:        "string",
				Description: "Unique identifier for the region",
				Required:    true,
				Nullable:    true,
			},
			"RegionName": {
				Name:        "RegionName",
				Type:        "string",
				Description: "Display name for the region",
				Required:    true,
				Nullable:    true,
			},
			"ResourceId": {
				Name:        "ResourceId",
				Type:        "string",
				Description: "Unique identifier for the resource",
				Required:    true,
				Nullable:    true,
			},
			"ResourceName": {
				Name:        "ResourceName",
				Type:        "string",
				Description: "Display name for the resource",
				Required:    true,
				Nullable:    true,
			},
			"ResourceType": {
				Name:        "ResourceType",
				Type:        "string",
				Description: "Type/category of the resource",
				Required:    true,
				Nullable:    true,
			},
			"ServiceCategory": {
				Name:        "ServiceCategory",
				Type:        "string",
				Description: "High-level category of the service",
				Required:    true,
				Nullable:    false,
			},
			"ServiceName": {
				Name:        "ServiceName",
				Type:        "string",
				Description: "Name of the service",
				Required:    true,
				Nullable:    false,
			},
			"SkuId": {
				Name:        "SkuId",
				Type:        "string",
				Description: "Unique identifier for the SKU",
				Required:    true,
				Nullable:    true,
			},
			"SkuPriceId": {
				Name:        "SkuPriceId",
				Type:        "string",
				Description: "Unique identifier for the SKU price",
				Required:    true,
				Nullable:    true,
			},
			"SubAccountId": {
				Name:        "SubAccountId",
				Type:        "string",
				Description: "Unique identifier for the sub-account",
				Required:    true,
				Nullable:    true,
			},
			"SubAccountName": {
				Name:        "SubAccountName",
				Type:        "string",
				Description: "Display name for the sub-account",
				Required:    true,
				Nullable:    true,
			},
			"Tags": {
				Name:        "Tags",
				Type:        "map",
				Description: "Key-value pairs for resource tags",
				Required:    true,
				Nullable:    true,
			},
		},
		OptionalColumns: map[string]ColumnSpec{
			"AvailabilityZone": {
				Name:        "AvailabilityZone",
				Type:        "string",
				Description: "Availability zone identifier",
				Required:    false,
				Nullable:    true,
			},
			"BilledCost": {
				Name:        "BilledCost",
				Type:        "decimal",
				Description: "Amount actually billed",
				Required:    false,
				Nullable:    true,
			},
			"ChargeDescription": {
				Name:        "ChargeDescription",
				Type:        "string",
				Description: "Detailed description of the charge",
				Required:    false,
				Nullable:    true,
			},
			"ContractedCost": {
				Name:        "ContractedCost",
				Type:        "decimal",
				Description: "Cost based on contracted rates",
				Required:    false,
				Nullable:    true,
			},
			"ContractedUnitPrice": {
				Name:        "ContractedUnitPrice",
				Type:        "decimal",
				Description: "Unit price based on contracted rates",
				Required:    false,
				Nullable:    true,
			},
		},
		Dimensions: []string{
			"BillingAccountId", "SubAccountId", "ResourceId", "ServiceName",
			"ServiceCategory", "ResourceType", "RegionId", "AvailabilityZone",
		},
		Metrics: []string{
			"EffectiveCost", "ListCost", "BilledCost", "ContractedCost",
			"PricingQuantity", "ListUnitPrice", "ContractedUnitPrice",
		},
		Metadata: SchemaMetadata{
			Title:       "FOCUS Schema v1.2",
			Description: "FinOps Open Cost and Usage Specification version 1.2",
			Version:     "1.2",
			LastUpdated: "2024-01-01",
			Source:      "https://github.com/FinOps-Open-Cost-and-Usage-Spec/FOCUS_Spec",
		},
	}

	m.schemas["focus-1.2"] = focus12

	// FOCUS v1.1 schema (simplified)
	focus11 := &FOCUSSchema{
		Version: "1.1",
		RequiredColumns: map[string]ColumnSpec{
			"BillingAccountId":   focus12.RequiredColumns["BillingAccountId"],
			"BillingAccountName": focus12.RequiredColumns["BillingAccountName"],
			"BillingCurrency":    focus12.RequiredColumns["BillingCurrency"],
			"EffectiveCost":      focus12.RequiredColumns["EffectiveCost"],
			"ListCost":           focus12.RequiredColumns["ListCost"],
			"ProviderName":       focus12.RequiredColumns["ProviderName"],
			"ServiceName":        focus12.RequiredColumns["ServiceName"],
			"ServiceCategory":    focus12.RequiredColumns["ServiceCategory"],
		},
		OptionalColumns: focus12.OptionalColumns,
		Dimensions:      []string{"BillingAccountId", "ServiceName", "ServiceCategory"},
		Metrics:         []string{"EffectiveCost", "ListCost"},
		Metadata: SchemaMetadata{
			Title:       "FOCUS Schema v1.1",
			Description: "FinOps Open Cost and Usage Specification version 1.1",
			Version:     "1.1",
			LastUpdated: "2023-06-01",
			Source:      "https://github.com/FinOps-Open-Cost-and-Usage-Spec/FOCUS_Spec",
		},
	}

	m.schemas["focus-1.1"] = focus11

	// FOCUS v1.0 schema (basic)
	focus10 := &FOCUSSchema{
		Version: "1.0",
		RequiredColumns: map[string]ColumnSpec{
			"BillingAccountId": focus12.RequiredColumns["BillingAccountId"],
			"EffectiveCost":    focus12.RequiredColumns["EffectiveCost"],
			"ListCost":         focus12.RequiredColumns["ListCost"],
			"ProviderName":     focus12.RequiredColumns["ProviderName"],
			"ServiceName":      focus12.RequiredColumns["ServiceName"],
		},
		OptionalColumns: map[string]ColumnSpec{},
		Dimensions:      []string{"BillingAccountId", "ServiceName"},
		Metrics:         []string{"EffectiveCost", "ListCost"},
		Metadata: SchemaMetadata{
			Title:       "FOCUS Schema v1.0",
			Description: "FinOps Open Cost and Usage Specification version 1.0",
			Version:     "1.0",
			LastUpdated: "2023-01-01",
			Source:      "https://github.com/FinOps-Open-Cost-and-Usage-Spec/FOCUS_Spec",
		},
	}

	m.schemas["focus-1.0"] = focus10
}

// GetSchema returns a schema by name
func (m *Manager) GetSchema(name string) (*FOCUSSchema, error) {
	schema, exists := m.schemas[name]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", name)
	}
	return schema, nil
}

// GetAvailableSchemas returns list of available schema names
func (m *Manager) GetAvailableSchemas() []string {
	var names []string
	for name := range m.schemas {
		names = append(names, name)
	}
	return names
}

// RegisterSchema registers a custom schema
func (m *Manager) RegisterSchema(name string, schema *FOCUSSchema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	m.schemas[name] = schema
	return nil
}

// ValidateSchemaCompatibility validates if a schema is compatible with FOCUS
func (m *Manager) ValidateSchemaCompatibility(schema *FOCUSSchema, baseSchema string) ([]string, error) {
	base, err := m.GetSchema(baseSchema)
	if err != nil {
		return nil, err
	}

	var issues []string

	// Check required columns
	for colName, colSpec := range base.RequiredColumns {
		if userCol, exists := schema.RequiredColumns[colName]; exists {
			if userCol.Type != colSpec.Type {
				issues = append(issues, fmt.Sprintf("Column %s type mismatch: expected %s, got %s",
					colName, colSpec.Type, userCol.Type))
			}
		} else {
			issues = append(issues, fmt.Sprintf("Missing required column: %s", colName))
		}
	}

	return issues, nil
}
