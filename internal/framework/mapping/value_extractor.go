package mapping

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// UniversalValueExtractor implements ValueExtractor for various data sources
type UniversalValueExtractor struct{}

// NewUniversalValueExtractor creates a new universal value extractor
func NewUniversalValueExtractor() *UniversalValueExtractor {
	return &UniversalValueExtractor{}
}

// ExtractString extracts a string value from various source types
func (uve *UniversalValueExtractor) ExtractString(source interface{}, field string) (string, bool, error) {
	switch src := source.(type) {
	case map[string]interface{}:
		return uve.extractFromMap(src, field)
	case map[string]string:
		return uve.extractFromStringMap(src, field)
	case CSVRowSource:
		return uve.extractFromCSVRow(src, field)
	default:
		return uve.extractFromStruct(source, field)
	}
}

// ExtractFloat extracts a float64 value from various source types
func (uve *UniversalValueExtractor) ExtractFloat(source interface{}, field string) (float64, bool, error) {
	// First try to get as string and convert
	strValue, exists, err := uve.ExtractString(source, field)
	if err != nil || !exists {
		return 0, exists, err
	}

	if strValue == "" {
		return 0, false, nil
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(strValue), 64)
	if err != nil {
		return 0, true, fmt.Errorf("failed to parse float from '%s': %w", strValue, err)
	}

	return value, true, nil
}

// ExtractInt extracts an int64 value from various source types
func (uve *UniversalValueExtractor) ExtractInt(source interface{}, field string) (int64, bool, error) {
	// First try to get as string and convert
	strValue, exists, err := uve.ExtractString(source, field)
	if err != nil || !exists {
		return 0, exists, err
	}

	if strValue == "" {
		return 0, false, nil
	}

	value, err := strconv.ParseInt(strings.TrimSpace(strValue), 10, 64)
	if err != nil {
		return 0, true, fmt.Errorf("failed to parse int from '%s': %w", strValue, err)
	}

	return value, true, nil
}

// ExtractBool extracts a boolean value from various source types
func (uve *UniversalValueExtractor) ExtractBool(source interface{}, field string) (bool, bool, error) {
	strValue, exists, err := uve.ExtractString(source, field)
	if err != nil || !exists {
		return false, exists, err
	}

	if strValue == "" {
		return false, false, nil
	}

	lowerValue := strings.ToLower(strings.TrimSpace(strValue))
	switch lowerValue {
	case "true", "1", "yes", "on", "enabled":
		return true, true, nil
	case "false", "0", "no", "off", "disabled":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("cannot parse bool from '%s'", strValue)
	}
}

// GetAvailableFields returns a list of available fields in the source
func (uve *UniversalValueExtractor) GetAvailableFields(source interface{}) []string {
	switch src := source.(type) {
	case map[string]interface{}:
		return getMapKeys(src)
	case map[string]string:
		return getMapKeys(src)
	case CSVRowSource:
		return src.Header
	default:
		return uve.getStructFields(source)
	}
}

// CSVRowSource represents a single CSV row plus its header slice.
// NOTE: This is a lightweight value container used by UniversalValueExtractor.
// It is NOT a streaming reader (providers have their own streaming constructors
// like aws.NewCSVRowSource / azure.NewCSVRowSourceFromReader). Avoid adding
// I/O logic here to keep separation between pure extraction helpers and
// provider ingestion pipelines.
type CSVRowSource struct {
	Header []string
	Row    []string
}

// extractFromMap extracts value from map[string]interface{}
func (uve *UniversalValueExtractor) extractFromMap(source map[string]interface{}, field string) (string, bool, error) {
	value, exists := source[field]
	if !exists {
		return "", false, nil
	}

	if value == nil {
		return "", true, nil
	}

	return fmt.Sprintf("%v", value), true, nil
}

// extractFromStringMap extracts value from map[string]string
func (uve *UniversalValueExtractor) extractFromStringMap(source map[string]string, field string) (string, bool, error) {
	value, exists := source[field]
	return value, exists, nil
}

// extractFromCSVRow extracts value from CSV row data
func (uve *UniversalValueExtractor) extractFromCSVRow(source CSVRowSource, field string) (string, bool, error) {
	// Find field index in header
	fieldIndex := -1
	for i, header := range source.Header {
		if header == field {
			fieldIndex = i
			break
		}
	}
	// Case-insensitive fallback if exact match not found (handles provider header casing / subtle normalization)
	if fieldIndex == -1 {
		lf := strings.ToLower(field)
		for i, header := range source.Header {
			if strings.ToLower(header) == lf {
				fieldIndex = i
				break
			}
		}
	}

	if fieldIndex == -1 {
		return "", false, nil // Field not found in header
	}

	if fieldIndex >= len(source.Row) {
		return "", false, fmt.Errorf("row has fewer columns (%d) than header (%d)", len(source.Row), len(source.Header))
	}

	value := strings.TrimSpace(source.Row[fieldIndex])
	return value, true, nil
}

// extractFromStruct extracts value from struct using reflection
func (uve *UniversalValueExtractor) extractFromStruct(source interface{}, field string) (string, bool, error) {
	if source == nil {
		return "", false, nil
	}

	value := reflect.ValueOf(source)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return "", false, nil
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return "", false, fmt.Errorf("source is not a struct, got %T", source)
	}

	// Try direct field name
	fieldValue := value.FieldByName(field)
	if fieldValue.IsValid() {
		return uve.convertFieldToString(fieldValue), true, nil
	}

	// Try case-insensitive search
	structType := value.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		if strings.EqualFold(structField.Name, field) {
			fieldValue = value.Field(i)
			return uve.convertFieldToString(fieldValue), true, nil
		}

		// Check struct tag for CSV field name
		csvTag := structField.Tag.Get("csv")
		if csvTag != "" {
			csvName := strings.Split(csvTag, ",")[0] // Handle tags like "name,omitempty"
			if csvName == field {
				fieldValue = value.Field(i)
				return uve.convertFieldToString(fieldValue), true, nil
			}
		}

		// Check struct tag for JSON field name
		jsonTag := structField.Tag.Get("json")
		if jsonTag != "" {
			jsonName := strings.Split(jsonTag, ",")[0]
			if jsonName == field {
				fieldValue = value.Field(i)
				return uve.convertFieldToString(fieldValue), true, nil
			}
		}
	}

	return "", false, nil // Field not found
}

// convertFieldToString converts a reflected field value to string
func (uve *UniversalValueExtractor) convertFieldToString(fieldValue reflect.Value) string {
	if !fieldValue.IsValid() || (fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil()) {
		return ""
	}

	// Dereference pointers
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	switch fieldValue.Kind() {
	case reflect.String:
		return fieldValue.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(fieldValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(fieldValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(fieldValue.Bool())
	default:
		return fmt.Sprintf("%v", fieldValue.Interface())
	}
}

// getMapKeys returns keys from any map[string]T (T can be any type)
func getMapKeys[M ~map[string]V, V any](source M) []string { //nolint:revive // generic helper for key extraction
	keys := make([]string, 0, len(source))
	for k := range source {
		keys = append(keys, k)
	}
	return keys
}

// getStructFields returns field names from struct
func (uve *UniversalValueExtractor) getStructFields(source interface{}) []string {
	if source == nil {
		return nil
	}

	value := reflect.ValueOf(source)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	structType := value.Type()
	fields := make([]string, 0, structType.NumField())

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.IsExported() {
			fields = append(fields, field.Name)
		}
	}

	return fields
}

// (Former helper NewCSVRowSource removed as deadcode: struct literals are clearer
// and avoid name collision with provider streaming constructors.)
