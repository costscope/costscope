package mapping

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// UniversalTypeConverter implements TypeConverter with common transformations
type UniversalTypeConverter struct {
	stringTransforms map[string]StringTransformFunc
	numberTransforms map[string]NumberTransformFunc
	timeFormats      []string
}

// StringTransformFunc represents a string transformation function
type StringTransformFunc func(string) string

// NumberTransformFunc represents a number transformation function
type NumberTransformFunc func(float64) float64

// NewUniversalTypeConverter creates a new universal type converter
func NewUniversalTypeConverter() *UniversalTypeConverter {
	converter := &UniversalTypeConverter{
		stringTransforms: make(map[string]StringTransformFunc),
		numberTransforms: make(map[string]NumberTransformFunc),
		timeFormats: []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"01/02/2006",
			"01/02/2006 15:04:05",
			"2006/01/02",
			"2006/01/02 15:04:05",
		},
	}

	// Register default transforms
	converter.registerDefaultStringTransforms()
	converter.registerDefaultNumberTransforms()

	return converter
}

// ConvertString converts and transforms string values
func (utc *UniversalTypeConverter) ConvertString(value string, mapping FieldMapping) (interface{}, error) {
	result := strings.TrimSpace(value)

	if mapping.Transform != "" {
		if transform, exists := utc.stringTransforms[mapping.Transform]; exists {
			result = transform(result)
		}
	}
	return result, nil
}

// ConvertFloat converts and transforms float64 values
func (utc *UniversalTypeConverter) ConvertFloat(value float64, mapping FieldMapping) (interface{}, error) {
	result := value
	if mapping.Transform != "" {
		if transform, exists := utc.numberTransforms[mapping.Transform]; exists {
			result = transform(result)
		}
	}
	return result, nil
}

// ConvertInt converts and transforms int64 values
func (utc *UniversalTypeConverter) ConvertInt(value int64, mapping FieldMapping) (interface{}, error) {
	return value, nil
}

// ConvertTime converts time strings to time.Time (UTC)
func (utc *UniversalTypeConverter) ConvertTime(value string, format string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}
	// Try the specified format first
	if format != "" {
		if t, err := time.Parse(format, value); err == nil {
			return t.UTC(), nil
		}
	}
	// Try common formats
	for _, tf := range utc.timeFormats {
		if t, err := time.Parse(tf, value); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse time '%s' with format '%s'", value, format)
}

// ConvertEnum converts enum values using the provided mapping
func (utc *UniversalTypeConverter) ConvertEnum(value string, enumMap map[string]string) (string, error) {
	if enumMap == nil {
		return value, nil
	}
	// Try exact match first
	if mapped, exists := enumMap[value]; exists {
		return mapped, nil
	}
	lower := strings.ToLower(value)
	for k, v := range enumMap {
		if strings.ToLower(k) == lower {
			return v, nil
		}
	}
	return value, nil
}

// registerDefaultStringTransforms registers common string transformation functions
func (utc *UniversalTypeConverter) registerDefaultStringTransforms() {
	utc.stringTransforms["lowercase"] = strings.ToLower
	utc.stringTransforms["uppercase"] = strings.ToUpper
	utc.stringTransforms["title"] = func(s string) string { return strings.ToTitle(s) }
	utc.stringTransforms["trim"] = strings.TrimSpace
	utc.stringTransforms["normalize_whitespace"] = func(s string) string {
		re := regexp.MustCompile(`\s+`)
		return strings.TrimSpace(re.ReplaceAllString(s, " "))
	}
	utc.stringTransforms["remove_prefix_slash"] = func(s string) string { return strings.TrimPrefix(s, "/") }
	utc.stringTransforms["remove_suffix_slash"] = func(s string) string { return strings.TrimSuffix(s, "/") }
}

// registerDefaultNumberTransforms registers common number transformation functions
func (utc *UniversalTypeConverter) registerDefaultNumberTransforms() {
	utc.numberTransforms["round_to_cents"] = func(f float64) float64 {
		// Use standard rounding that handles negatives correctly
		return math.Round(f*100) / 100
	}
	utc.numberTransforms["absolute"] = func(f float64) float64 {
		if f < 0 {
			return -f
		}
		return f
	}
}

// RegisterStringTransform registers a custom string transformation function
func (utc *UniversalTypeConverter) RegisterStringTransform(name string, transform StringTransformFunc) {
	utc.stringTransforms[name] = transform
}

// RegisterNumberTransform registers a custom number transformation function
func (utc *UniversalTypeConverter) RegisterNumberTransform(name string, transform NumberTransformFunc) {
	utc.numberTransforms[name] = transform
}

// AddTimeFormat adds a custom time format to try when parsing time strings
func (utc *UniversalTypeConverter) AddTimeFormat(format string) {
	utc.timeFormats = append(utc.timeFormats, format)
}

// GetAvailableTransforms returns lists of available transformation functions
func (utc *UniversalTypeConverter) GetAvailableTransforms() ([]string, []string) {
	s := make([]string, 0, len(utc.stringTransforms))
	for name := range utc.stringTransforms {
		s = append(s, name)
	}
	n := make([]string, 0, len(utc.numberTransforms))
	for name := range utc.numberTransforms {
		n = append(n, name)
	}
	return s, n
}
