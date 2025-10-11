package types

import (
	"time"
)

// =====================================================================================
// FOCUS v1.2 Data Types - FinOps Open Cloud Usage Specification
// =====================================================================================
// Complete implementation of FOCUS v1.2 specification for multi-cloud billing data

// FocusRecord represents a single FOCUS v1.2 billing record
type FocusRecord struct {
	// Core Required Fields
	BillingAccountId   string    `parquet:"name=billing_account_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"billing_account_id"`
	BillingAccountName string    `parquet:"name=billing_account_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"billing_account_name"`
	BillingCurrency    string    `parquet:"name=billing_currency, type=BYTE_ARRAY, convertedtype=UTF8" json:"billing_currency"`
	BillingPeriodEnd   time.Time `parquet:"name=billing_period_end, type=INT64, convertedtype=TIMESTAMP_MILLIS" json:"billing_period_end"`
	BillingPeriodStart time.Time `parquet:"name=billing_period_start, type=INT64, convertedtype=TIMESTAMP_MILLIS" json:"billing_period_start"`
	ChargeCategory     string    `parquet:"name=charge_category, type=BYTE_ARRAY, convertedtype=UTF8" json:"charge_category"`
	ChargeClass        string    `parquet:"name=charge_class, type=BYTE_ARRAY, convertedtype=UTF8" json:"charge_class"`
	ChargeDescription  string    `parquet:"name=charge_description, type=BYTE_ARRAY, convertedtype=UTF8" json:"charge_description"`
	ChargeFrequency    string    `parquet:"name=charge_frequency, type=BYTE_ARRAY, convertedtype=UTF8" json:"charge_frequency"`
	ChargePeriodEnd    time.Time `parquet:"name=charge_period_end, type=INT64, convertedtype=TIMESTAMP_MILLIS" json:"charge_period_end"`
	ChargePeriodStart  time.Time `parquet:"name=charge_period_start, type=INT64, convertedtype=TIMESTAMP_MILLIS" json:"charge_period_start"`
	ChargeSubcategory  string    `parquet:"name=charge_subcategory, type=BYTE_ARRAY, convertedtype=UTF8" json:"charge_subcategory"`
	EffectiveCost      float64   `parquet:"name=effective_cost, type=DOUBLE" json:"effective_cost"`
	InvoiceIssuerName  string    `parquet:"name=invoice_issuer_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"invoice_issuer_name"`
	ListCost           float64   `parquet:"name=list_cost, type=DOUBLE" json:"list_cost"`
	ListUnitPrice      float64   `parquet:"name=list_unit_price, type=DOUBLE" json:"list_unit_price"`
	PricingCategory    string    `parquet:"name=pricing_category, type=BYTE_ARRAY, convertedtype=UTF8" json:"pricing_category"`
	PricingQuantity    float64   `parquet:"name=pricing_quantity, type=DOUBLE" json:"pricing_quantity"`
	PricingUnit        string    `parquet:"name=pricing_unit, type=BYTE_ARRAY, convertedtype=UTF8" json:"pricing_unit"`
	ProviderName       string    `parquet:"name=provider_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"provider_name"`
	PublisherName      string    `parquet:"name=publisher_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"publisher_name"`
	ResourceId         string    `parquet:"name=resource_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"resource_id"`
	ResourceName       string    `parquet:"name=resource_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"resource_name"`
	ResourceType       string    `parquet:"name=resource_type, type=BYTE_ARRAY, convertedtype=UTF8" json:"resource_type"`
	ServiceCategory    string    `parquet:"name=service_category, type=BYTE_ARRAY, convertedtype=UTF8" json:"service_category"`
	ServiceName        string    `parquet:"name=service_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"service_name"`
	SkuId              string    `parquet:"name=sku_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"sku_id"`
	SkuPriceId         string    `parquet:"name=sku_price_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"sku_price_id"`
	SubAccountId       string    `parquet:"name=sub_account_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"sub_account_id"`
	SubAccountName     string    `parquet:"name=sub_account_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"sub_account_name"`
	UsageQuantity      float64   `parquet:"name=usage_quantity, type=DOUBLE" json:"usage_quantity"`
	UsageUnit          string    `parquet:"name=usage_unit, type=BYTE_ARRAY, convertedtype=UTF8" json:"usage_unit"`

	// Optional Extended Fields
	AvailabilityZone       *string  `parquet:"name=availability_zone, type=BYTE_ARRAY, convertedtype=UTF8" json:"availability_zone,omitempty"`
	BilledCost             *float64 `parquet:"name=billed_cost, type=DOUBLE" json:"billed_cost,omitempty"`
	CommitmentDiscountId   *string  `parquet:"name=commitment_discount_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"commitment_discount_id,omitempty"`
	CommitmentDiscountName *string  `parquet:"name=commitment_discount_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"commitment_discount_name,omitempty"`
	CommitmentDiscountType *string  `parquet:"name=commitment_discount_type, type=BYTE_ARRAY, convertedtype=UTF8" json:"commitment_discount_type,omitempty"`
	ConsumedQuantity       *float64 `parquet:"name=consumed_quantity, type=DOUBLE" json:"consumed_quantity,omitempty"`
	ConsumedUnit           *string  `parquet:"name=consumed_unit, type=BYTE_ARRAY, convertedtype=UTF8" json:"consumed_unit,omitempty"`
	ContractedCost         *float64 `parquet:"name=contracted_cost, type=DOUBLE" json:"contracted_cost,omitempty"`
	ContractedUnitPrice    *float64 `parquet:"name=contracted_unit_price, type=DOUBLE" json:"contracted_unit_price,omitempty"`
	InstanceId             *string  `parquet:"name=instance_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"instance_id,omitempty"`
	InstanceName           *string  `parquet:"name=instance_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"instance_name,omitempty"`
	InstanceType           *string  `parquet:"name=instance_type, type=BYTE_ARRAY, convertedtype=UTF8" json:"instance_type,omitempty"`
	InvoiceId              *string  `parquet:"name=invoice_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"invoice_id,omitempty"`
	Region                 *string  `parquet:"name=region, type=BYTE_ARRAY, convertedtype=UTF8" json:"region,omitempty"`
	Tags                   Tags     `parquet:"name=tags" json:"tags,omitempty"`

	// Metadata for tracking
	SourceProvider      string    `parquet:"name=source_provider, type=BYTE_ARRAY, convertedtype=UTF8" json:"source_provider"`
	SourceFileName      string    `parquet:"name=source_file_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"source_file_name"`
	ConversionTimestamp time.Time `parquet:"name=conversion_timestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS" json:"conversion_timestamp"`

	// Multi-tenant placeholder (optional). Only populated when multi-tenancy feature is enabled.
	// TASK-MULTITENANT-SKEL: Not yet enforced in queries or isolation.
	TenantID *string `parquet:"name=tenant_id, type=BYTE_ARRAY, convertedtype=UTF8" json:"tenant_id,omitempty"`
}

// Tags represents resource tags as key-value pairs
type Tags map[string]string

// FocusSchema represents the complete FOCUS v1.2 schema metadata
type FocusSchema struct {
	Version       string              `json:"version"`
	SchemaVersion string              `json:"schema_version"`
	Fields        []FocusFieldSchema  `json:"fields"`
	Metadata      FocusSchemaMetadata `json:"metadata"`
}

// FocusFieldSchema represents a single field in the FOCUS schema
type FocusFieldSchema struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Constraints []string `json:"constraints,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// FocusSchemaMetadata contains metadata about the schema
type FocusSchemaMetadata struct {
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	Source       string    `json:"source"`
	TotalRecords int64     `json:"total_records"`
	FileSizeMB   float64   `json:"file_size_mb"`
}

// ChargeCategories defines standard FOCUS charge categories
var ChargeCategories = struct {
	Usage      string
	Purchase   string
	Tax        string
	Adjustment string
	Credit     string
}{
	Usage:      "Usage",
	Purchase:   "Purchase",
	Tax:        "Tax",
	Adjustment: "Adjustment",
	Credit:     "Credit",
}

// ChargeClasses defines standard FOCUS charge classes
var ChargeClasses = struct {
	OnDemand   string
	Commitment string
	Correction string
}{
	OnDemand:   "On-Demand",
	Commitment: "Commitment",
	Correction: "Correction",
}

// PricingCategories defines standard FOCUS pricing categories
var PricingCategories = struct {
	Standard string
	Spot     string
	Reserved string
}{
	Standard: "Standard",
	Spot:     "Spot",
	Reserved: "Reserved",
}

// ProviderNames defines standard provider names
var ProviderNames = struct {
	AWS   string
	Azure string
	GCP   string
}{
	AWS:   "Amazon Web Services",
	Azure: "Microsoft Azure",
	GCP:   "Google Cloud Platform",
}

// GetFocusV12Schema returns the complete FOCUS v1.2 schema definition
func GetFocusV12Schema() *FocusSchema {
	return &FocusSchema{
		Version:       "1.2",
		SchemaVersion: "FOCUS_1.2",
		Fields: []FocusFieldSchema{
			{Name: "billing_account_id", Type: "string", Required: true, Description: "The identifier for the billing account"},
			{Name: "billing_account_name", Type: "string", Required: true, Description: "The display name for the billing account"},
			{Name: "billing_currency", Type: "string", Required: true, Description: "The currency used for billing"},
			{Name: "billing_period_end", Type: "datetime", Required: true, Description: "The end date and time of the billing period"},
			{Name: "billing_period_start", Type: "datetime", Required: true, Description: "The start date and time of the billing period"},
			{Name: "charge_category", Type: "string", Required: true, Description: "High-level category of the charge"},
			{Name: "charge_class", Type: "string", Required: true, Description: "Indicates whether the row represents an on-demand or amortized commitment"},
			{Name: "charge_description", Type: "string", Required: true, Description: "Description of the charge"},
			{Name: "charge_frequency", Type: "string", Required: true, Description: "How often the charge occurs"},
			{Name: "charge_period_end", Type: "datetime", Required: true, Description: "The end date and time of the charge period"},
			{Name: "charge_period_start", Type: "datetime", Required: true, Description: "The start date and time of the charge period"},
			{Name: "charge_subcategory", Type: "string", Required: true, Description: "Additional subcategory for the charge"},
			{Name: "effective_cost", Type: "decimal", Required: true, Description: "The effective cost after discounts and credits"},
			{Name: "invoice_issuer_name", Type: "string", Required: true, Description: "The name of the entity issuing the invoice"},
			{Name: "list_cost", Type: "decimal", Required: true, Description: "The cost based on public list prices"},
			{Name: "list_unit_price", Type: "decimal", Required: true, Description: "The unit price based on public list prices"},
			{Name: "pricing_category", Type: "string", Required: true, Description: "The category of pricing model used"},
			{Name: "pricing_quantity", Type: "decimal", Required: true, Description: "The quantity used for pricing"},
			{Name: "pricing_unit", Type: "string", Required: true, Description: "The unit of measure for pricing"},
			{Name: "provider_name", Type: "string", Required: true, Description: "The name of the cloud provider"},
			{Name: "publisher_name", Type: "string", Required: true, Description: "The name of the service publisher"},
			{Name: "resource_id", Type: "string", Required: true, Description: "The unique identifier for the resource"},
			{Name: "resource_name", Type: "string", Required: true, Description: "The display name for the resource"},
			{Name: "resource_type", Type: "string", Required: true, Description: "The type of the resource"},
			{Name: "service_category", Type: "string", Required: true, Description: "The high-level category of the service"},
			{Name: "service_name", Type: "string", Required: true, Description: "The name of the service"},
			{Name: "sku_id", Type: "string", Required: true, Description: "The unique identifier for the SKU"},
			{Name: "sku_price_id", Type: "string", Required: true, Description: "The unique identifier for the SKU price"},
			{Name: "sub_account_id", Type: "string", Required: true, Description: "The identifier for the sub-account"},
			{Name: "sub_account_name", Type: "string", Required: true, Description: "The display name for the sub-account"},
			{Name: "usage_quantity", Type: "decimal", Required: true, Description: "The quantity of usage"},
			{Name: "usage_unit", Type: "string", Required: true, Description: "The unit of measure for usage"},
		},
		Metadata: FocusSchemaMetadata{
			CreatedAt: time.Now(),
			CreatedBy: "CostScope FOCUS Converter",
			Source:    "FOCUS v1.2 Specification",
		},
	}
}

// ValidateFocusRecord validates a FOCUS record against the v1.2 specification
func (record *FocusRecord) Validate() error {
	// TODO: Implement comprehensive validation
	return nil
}

// ToMap converts a FOCUS record to a map for JSON serialization
func (record *FocusRecord) ToMap() map[string]interface{} {
	// TODO: Implement map conversion
	return make(map[string]interface{})
}
