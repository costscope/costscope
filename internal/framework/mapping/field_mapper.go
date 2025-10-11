package mapping

import (
	"fmt"
	"reflect"
	"time"

	ftypes "local/costscope/internal/core/focus/types"
)

// FieldMapper provides unified field mapping abstraction for all cloud providers
type FieldMapper struct {
	config         *FieldMappingConfig
	valueExtractor ValueExtractor
	typeConverter  TypeConverter
	defaultHandler DefaultHandler
	ordered        []orderedMapping
}

type orderedMapping struct {
	target  string
	mapping FieldMapping
}

// FieldMappingConfig holds configuration for field mapping
type FieldMappingConfig struct {
	ProviderName    string                           `json:"provider_name"`
	FieldMappings   map[string]FieldMapping          `json:"field_mappings"`
	EnumMappings    map[string]map[string]string     `json:"enum_mappings"`
	DefaultValues   map[string]interface{}           `json:"default_values"`
	ValidationRules map[string]ValidationRule        `json:"validation_rules"`
	TimeFormats     map[string]string                `json:"time_formats"`
	CustomMappings  map[string]CustomMappingFunction `json:"-"` // Not serializable
}

// FieldMapping defines how a source field maps to a FOCUS field
type FieldMapping struct {
	SourceField  string      `json:"source_field"`
	TargetField  string      `json:"target_field"`
	FieldType    FieldType   `json:"field_type"`
	IsRequired   bool        `json:"is_required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Transform    string      `json:"transform,omitempty"`
	Validation   string      `json:"validation,omitempty"`
	TimeFormat   string      `json:"time_format,omitempty"`
	EnumMapping  string      `json:"enum_mapping,omitempty"`
}

// FieldType represents the data type of a field
type FieldType string

const (
	FieldTypeString   FieldType = "string"
	FieldTypeFloat64  FieldType = "float64"
	FieldTypeInt64    FieldType = "int64"
	FieldTypeBool     FieldType = "bool"
	FieldTypeTime     FieldType = "time"
	FieldTypeEnum     FieldType = "enum"
	FieldTypeOptional FieldType = "optional"
)

// ValidationRule mirrors a simple numeric/string validation definition
type ValidationRule struct {
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	MinValue      *float64 `json:"min_value,omitempty"`
	MaxValue      *float64 `json:"max_value,omitempty"`
}

// ValueExtractor defines interface for extracting values from different data sources
type ValueExtractor interface {
	ExtractString(source interface{}, field string) (string, bool, error)
	ExtractFloat(source interface{}, field string) (float64, bool, error)
	ExtractInt(source interface{}, field string) (int64, bool, error)
	ExtractBool(source interface{}, field string) (bool, bool, error)
	GetAvailableFields(source interface{}) []string
}

// TypeConverter handles type conversions and transformations
type TypeConverter interface {
	ConvertString(value string, mapping FieldMapping) (interface{}, error)
	ConvertFloat(value float64, mapping FieldMapping) (interface{}, error)
	ConvertInt(value int64, mapping FieldMapping) (interface{}, error)
	ConvertTime(value string, format string) (time.Time, error)
	ConvertEnum(value string, enumMap map[string]string) (string, error)
}

// DefaultHandler manages default values and missing field handling
type DefaultHandler interface {
	GetDefaultValue(field string, fieldType FieldType) interface{}
	ApplyDefaults(record *ftypes.FocusRecord) error
}

// CustomMappingFunction allows custom field mapping logic
type CustomMappingFunction func(source interface{}, mapping FieldMapping, extractor ValueExtractor) (interface{}, error)

// NewFieldMapper creates a new field mapper with the specified configuration
func NewFieldMapper(config *FieldMappingConfig) (*FieldMapper, error) {
	if config == nil {
		return nil, fmt.Errorf("field mapping config cannot be nil")
	}

	// Initialize maps
	if config.FieldMappings == nil {
		config.FieldMappings = make(map[string]FieldMapping)
	}
	if config.EnumMappings == nil {
		config.EnumMappings = make(map[string]map[string]string)
	}
	if config.DefaultValues == nil {
		config.DefaultValues = make(map[string]interface{})
	}
	if config.ValidationRules == nil {
		config.ValidationRules = make(map[string]ValidationRule)
	}
	if config.TimeFormats == nil {
		config.TimeFormats = make(map[string]string)
	}
	if config.CustomMappings == nil {
		config.CustomMappings = make(map[string]CustomMappingFunction)
	}

	mapper := &FieldMapper{
		config:         config,
		valueExtractor: NewUniversalValueExtractor(),
		typeConverter:  NewUniversalTypeConverter(),
		defaultHandler: NewUniversalDefaultHandler(config),
	}
	// Precompute stable ordered slice of mappings to avoid per-record map iteration allocations
	mapper.ordered = make([]orderedMapping, 0, len(config.FieldMappings))
	for k, v := range config.FieldMappings {
		mapper.ordered = append(mapper.ordered, orderedMapping{target: k, mapping: v})
	}

	return mapper, nil
}

// MapToFOCUS maps a source record to a FOCUS record using the configured mappings
func (fm *FieldMapper) MapToFOCUS(source interface{}) (*ftypes.FocusRecord, error) {
	if source == nil {
		return nil, fmt.Errorf("source record cannot be nil")
	}

	// Create new FOCUS record with provider-specific defaults
	record := &ftypes.FocusRecord{}

	// Apply default values first
	if err := fm.defaultHandler.ApplyDefaults(record); err != nil {
		return nil, fmt.Errorf("failed to apply default values: %w", err)
	}

	// Process each field mapping in stable order (slice) to reduce allocations
	for _, om := range fm.ordered {
		value, err := fm.mapSingleField(source, om.mapping)
		if err != nil {
			if om.mapping.IsRequired {
				return nil, fmt.Errorf("failed to map required field %s: %w", om.target, err)
			}
			// Skip optional fields that fail to map
			continue
		}

		// Set the value in the FOCUS record
		if err := fm.setFieldValue(record, om.target, value); err != nil {
			return nil, fmt.Errorf("failed to set field %s in FOCUS record: %w", om.target, err)
		}
	}

	// Validate the final record (lightweight)
	if err := fm.validateRecord(record); err != nil {
		return nil, fmt.Errorf("record validation failed: %w", err)
	}

	return record, nil
}

// mapSingleField maps a single field from source to target using the field mapping
func (fm *FieldMapper) mapSingleField(source interface{}, mapping FieldMapping) (interface{}, error) {
	// Check for custom mapping function first
	if customFunc, exists := fm.config.CustomMappings[mapping.Transform]; exists {
		return customFunc(source, mapping, fm.valueExtractor)
	}

	// Extract raw value from source
	rawValue, exists, err := fm.extractRawValue(source, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to extract value for field %s: %w", mapping.SourceField, err)
	}

	// Handle missing values
	if !exists {
		if mapping.IsRequired {
			return nil, fmt.Errorf("required field %s is missing", mapping.SourceField)
		}
		return fm.defaultHandler.GetDefaultValue(mapping.TargetField, mapping.FieldType), nil
	}

	// Convert and transform the value
	return fm.convertValue(rawValue, mapping)
}

// extractRawValue extracts the raw value from source based on field type
func (fm *FieldMapper) extractRawValue(source interface{}, mapping FieldMapping) (interface{}, bool, error) {
	switch mapping.FieldType {
	case FieldTypeString, FieldTypeEnum:
		return fm.valueExtractor.ExtractString(source, mapping.SourceField)
	case FieldTypeFloat64:
		return fm.valueExtractor.ExtractFloat(source, mapping.SourceField)
	case FieldTypeInt64:
		return fm.valueExtractor.ExtractInt(source, mapping.SourceField)
	case FieldTypeBool:
		return fm.valueExtractor.ExtractBool(source, mapping.SourceField)
	case FieldTypeTime:
		// Time fields are extracted as strings and then converted
		return fm.valueExtractor.ExtractString(source, mapping.SourceField)
	case FieldTypeOptional:
		// Optional fields can be any type, try string first
		return fm.valueExtractor.ExtractString(source, mapping.SourceField)
	default:
		return nil, false, fmt.Errorf("unsupported field type: %s", mapping.FieldType)
	}
}

// convertValue converts the raw value to the target type
func (fm *FieldMapper) convertValue(rawValue interface{}, mapping FieldMapping) (interface{}, error) {
	switch mapping.FieldType {
	case FieldTypeString:
		if str, ok := rawValue.(string); ok {
			return fm.typeConverter.ConvertString(str, mapping)
		}
		return fmt.Sprintf("%v", rawValue), nil

	case FieldTypeFloat64:
		if f, ok := rawValue.(float64); ok {
			return fm.typeConverter.ConvertFloat(f, mapping)
		}
		return nil, fmt.Errorf("expected float64, got %T", rawValue)

	case FieldTypeInt64:
		if i, ok := rawValue.(int64); ok {
			return fm.typeConverter.ConvertInt(i, mapping)
		}
		return nil, fmt.Errorf("expected int64, got %T", rawValue)

	case FieldTypeBool:
		if b, ok := rawValue.(bool); ok {
			return b, nil
		}
		return nil, fmt.Errorf("expected bool, got %T", rawValue)

	case FieldTypeTime:
		if str, ok := rawValue.(string); ok {
			format := mapping.TimeFormat
			if format == "" {
				format = time.RFC3339
			}
			return fm.typeConverter.ConvertTime(str, format)
		}
		return nil, fmt.Errorf("expected string for time field, got %T", rawValue)

	case FieldTypeEnum:
		if str, ok := rawValue.(string); ok {
			if enumMap, exists := fm.config.EnumMappings[mapping.EnumMapping]; exists {
				return fm.typeConverter.ConvertEnum(str, enumMap)
			}
			return str, nil // Return as-is if no enum mapping
		}
		return nil, fmt.Errorf("expected string for enum field, got %T", rawValue)

	case FieldTypeOptional:
		// For optional fields, wrap in pointer
		return fm.wrapOptionalValue(rawValue)

	default:
		return nil, fmt.Errorf("unsupported field type: %s", mapping.FieldType)
	}
}

// wrapOptionalValue wraps a value in a pointer for optional fields
func (fm *FieldMapper) wrapOptionalValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		return &v, nil
	case float64:
		vv := v
		return &vv, nil
	case int64:
		vv := v
		return &vv, nil
	case bool:
		vv := v
		return &vv, nil
	case time.Time:
		if v.IsZero() {
			return nil, nil
		}
		vv := v
		return &vv, nil
	default:
		return nil, fmt.Errorf("unsupported optional field type: %T", value)
	}
}

// setFieldValue sets a value in the FOCUS record using reflection
func (fm *FieldMapper) setFieldValue(record *ftypes.FocusRecord, fieldName string, value interface{}) error {
	if value == nil {
		return nil // Skip nil values
	}

	recordValue := reflect.ValueOf(record).Elem()
	fieldValue := recordValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return fmt.Errorf("field %s does not exist in FocusRecord", fieldName)
	}

	if !fieldValue.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	valueToSet := reflect.ValueOf(value)

	// Handle exact assignable types
	if valueToSet.Type().AssignableTo(fieldValue.Type()) {
		fieldValue.Set(valueToSet)
		return nil
	}

	// Support assigning non-pointer to pointer fields when element types match
	if fieldValue.Kind() == reflect.Ptr && valueToSet.Type().AssignableTo(fieldValue.Type().Elem()) {
		ptr := reflect.New(fieldValue.Type().Elem())
		ptr.Elem().Set(valueToSet)
		fieldValue.Set(ptr)
		return nil
	}

	// Convertible types
	if valueToSet.Type().ConvertibleTo(fieldValue.Type()) {
		fieldValue.Set(valueToSet.Convert(fieldValue.Type()))
		return nil
	}

	// Convertible to pointer element
	if fieldValue.Kind() == reflect.Ptr && valueToSet.Type().ConvertibleTo(fieldValue.Type().Elem()) {
		ptr := reflect.New(fieldValue.Type().Elem())
		ptr.Elem().Set(valueToSet.Convert(fieldValue.Type().Elem()))
		fieldValue.Set(ptr)
		return nil
	}

	return fmt.Errorf("cannot convert %T to %s for field %s", value, fieldValue.Type(), fieldName)
}

// validateRecord validates the final FOCUS record against simple rules
func (fm *FieldMapper) validateRecord(record *ftypes.FocusRecord) error {
	recordValue := reflect.ValueOf(record).Elem()

	for fieldName, rule := range fm.config.ValidationRules {
		fieldValue := recordValue.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			continue
		}

		if err := fm.validateFieldValue(fieldValue.Interface(), rule); err != nil {
			return fmt.Errorf("validation failed for field %s: %w", fieldName, err)
		}
	}

	return nil
}

// validateFieldValue validates a single field value against its validation rule
func (fm *FieldMapper) validateFieldValue(value interface{}, rule ValidationRule) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		return fm.validateStringField(v, rule)
	case float64:
		return fm.validateNumericField(v, rule)
	case int64:
		return fm.validateNumericField(float64(v), rule)
	case *string:
		if v != nil {
			return fm.validateStringField(*v, rule)
		}
	case *float64:
		if v != nil {
			return fm.validateNumericField(*v, rule)
		}
	case *int64:
		if v != nil {
			return fm.validateNumericField(float64(*v), rule)
		}
	}

	return nil
}

func (fm *FieldMapper) validateStringField(value string, rule ValidationRule) error {
	if rule.MinLength != nil && len(value) < *rule.MinLength {
		return fmt.Errorf("string length %d is less than minimum %d", len(value), *rule.MinLength)
	}
	if rule.MaxLength != nil && len(value) > *rule.MaxLength {
		return fmt.Errorf("string length %d exceeds maximum %d", len(value), *rule.MaxLength)
	}
	if len(rule.AllowedValues) > 0 {
		ok := false
		for _, allowed := range rule.AllowedValues {
			if value == allowed {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("value '%s' is not in allowed values", value)
		}
	}
	return nil
}

func (fm *FieldMapper) validateNumericField(value float64, rule ValidationRule) error {
	if rule.MinValue != nil && value < *rule.MinValue {
		return fmt.Errorf("value %f is less than minimum %f", value, *rule.MinValue)
	}
	if rule.MaxValue != nil && value > *rule.MaxValue {
		return fmt.Errorf("value %f exceeds maximum %f", value, *rule.MaxValue)
	}
	return nil
}

// ValidateConfig performs basic validation of the mapping config
// NOTE: historical helpers MapToFOCUSInto and ValidateConfig removed as unused in active unified mapper paths.
