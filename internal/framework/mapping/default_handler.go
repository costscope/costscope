package mapping

import (
	"fmt"
	"reflect"

	ftypes "github.com/costscope/costscope/internal/core/focus/types"
)

const defaultCurrencyUSD = "USD"

// UniversalDefaultHandler implements DefaultHandler with configurable defaults
type UniversalDefaultHandler struct {
	config           *FieldMappingConfig
	providerDefaults map[string]interface{}
}

// NewUniversalDefaultHandler creates a new universal default handler
func NewUniversalDefaultHandler(config *FieldMappingConfig) *UniversalDefaultHandler {
	handler := &UniversalDefaultHandler{
		config:           config,
		providerDefaults: make(map[string]interface{}),
	}
	handler.initializeProviderDefaults()
	return handler
}

// GetDefaultValue returns the default value for a field and type
func (udh *UniversalDefaultHandler) GetDefaultValue(field string, fieldType FieldType) interface{} {
	if defaultValue, exists := udh.config.DefaultValues[field]; exists {
		return defaultValue
	}
	if defaultValue, exists := udh.providerDefaults[field]; exists {
		return defaultValue
	}
	return udh.getTypeDefault(fieldType)
}

// HandleMissingField handles missing fields based on whether they are required
// (Removed unused HandleMissingField method – missing field logic handled inline in mapper.)

// ApplyDefaults applies default values to a FOCUS record
func (udh *UniversalDefaultHandler) ApplyDefaults(record *ftypes.FocusRecord) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}
	// Provider-specific defaults via field names in FocusRecord
	udh.applyProviderDefaults(record)
	return nil
}

// initializeProviderDefaults sets up provider-specific default values
func (udh *UniversalDefaultHandler) initializeProviderDefaults() {
	providerName := udh.config.ProviderName
	switch providerName {
	case ftypes.ProviderNames.AWS, "aws":
		udh.initializeAWSDefaults()
	case ftypes.ProviderNames.Azure, "azure":
		udh.initializeAzureDefaults()
	case ftypes.ProviderNames.GCP, "gcp":
		udh.initializeGCPDefaults()
	default:
		udh.initializeGenericDefaults()
	}
}

func (udh *UniversalDefaultHandler) initializeAWSDefaults() {
	udh.providerDefaults["ProviderName"] = ftypes.ProviderNames.AWS
	udh.providerDefaults["PublisherName"] = ftypes.ProviderNames.AWS
	udh.providerDefaults["ChargeClass"] = ftypes.ChargeClasses.OnDemand
	udh.providerDefaults["PricingCategory"] = ftypes.PricingCategories.Standard
	udh.providerDefaults["ChargeCategory"] = ftypes.ChargeCategories.Usage
	udh.providerDefaults["BillingCurrency"] = defaultCurrencyUSD
}

func (udh *UniversalDefaultHandler) initializeAzureDefaults() {
	udh.providerDefaults["ProviderName"] = ftypes.ProviderNames.Azure
	// PublisherName for Azure should be the company name to match legacy converters
	// Legacy fast path uses "Microsoft"; keep parity here.
	udh.providerDefaults["PublisherName"] = "Microsoft"
	udh.providerDefaults["ChargeClass"] = ftypes.ChargeClasses.OnDemand
	udh.providerDefaults["PricingCategory"] = ftypes.PricingCategories.Standard
	udh.providerDefaults["ChargeCategory"] = ftypes.ChargeCategories.Usage
	udh.providerDefaults["BillingCurrency"] = defaultCurrencyUSD
}

func (udh *UniversalDefaultHandler) initializeGCPDefaults() {
	udh.providerDefaults["ProviderName"] = ftypes.ProviderNames.GCP
	// PublisherName for GCP is "Google" in legacy fast path; mirror it for unified parity.
	udh.providerDefaults["PublisherName"] = "Google"
	udh.providerDefaults["ChargeClass"] = ftypes.ChargeClasses.OnDemand
	udh.providerDefaults["PricingCategory"] = ftypes.PricingCategories.Standard
	udh.providerDefaults["ChargeCategory"] = ftypes.ChargeCategories.Usage
	udh.providerDefaults["BillingCurrency"] = defaultCurrencyUSD
}

func (udh *UniversalDefaultHandler) initializeGenericDefaults() {
	udh.providerDefaults["ChargeClass"] = ftypes.ChargeClasses.OnDemand
	udh.providerDefaults["PricingCategory"] = ftypes.PricingCategories.Standard
	udh.providerDefaults["ChargeCategory"] = ftypes.ChargeCategories.Usage
	udh.providerDefaults["BillingCurrency"] = defaultCurrencyUSD
}

// getTypeDefault returns default values based on field type
func (udh *UniversalDefaultHandler) getTypeDefault(fieldType FieldType) interface{} {
	switch fieldType {
	case FieldTypeString:
		return ""
	case FieldTypeFloat64:
		return 0.0
	case FieldTypeInt64:
		return int64(0)
	case FieldTypeBool:
		return false
	case FieldTypeTime:
		return ftypes.FocusRecord{}.BillingPeriodStart // zero time
	case FieldTypeEnum:
		return ""
	case FieldTypeOptional:
		return nil
	default:
		return nil
	}
}

// applyProviderDefaults applies provider-specific defaults using reflection
func (udh *UniversalDefaultHandler) applyProviderDefaults(record *ftypes.FocusRecord) {
	recordValue := reflect.ValueOf(record).Elem()
	for fieldName, defaultValue := range udh.providerDefaults {
		fieldValue := recordValue.FieldByName(fieldName)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}
		if udh.isZeroValue(fieldValue) {
			dv := reflect.ValueOf(defaultValue)
			if dv.Type().AssignableTo(fieldValue.Type()) {
				fieldValue.Set(dv)
			}
		}
	}
}

// isZeroValue checks if a reflected value is zero/empty
func (udh *UniversalDefaultHandler) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0.0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}
