package azure

import (
	"strconv"
	"strings"
	"time"
)

// ParseFloat safely parses a float, returning 0 on error.
func ParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// FirstNonEmptyField returns first non-empty field among keys using case-insensitive lookup.
func FirstNonEmptyField(rec map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := GetField(rec, k); strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// GetField performs case-insensitive lookup for a key in the record map.
func GetField(rec map[string]string, key string) string {
	if v, ok := rec[key]; ok {
		return v
	}
	for k, v := range rec {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

// FirstNonEmptyValue returns the first non-empty value.
func FirstNonEmptyValue(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// ParseTimeFlexible parses RFC3339 and common time layouts; falls back to date-only parsing.
func ParseTimeFlexible(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05Z07:00", "2006-01-02 15:04:05"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC()
		}
	}
	// Fallback to date only
	return ParseDateOnlyFlexible(s)
}

// ParseDateOnlyFlexible parses common date-only layouts.
func ParseDateOnlyFlexible(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{"2006-01-02", "20060102"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// NormalizeRegion returns a normalized region/location string (lowercase, trimmed).
func NormalizeRegion(s string) string {
	if strings.TrimSpace(s) == "" {
		return s
	}
	return strings.ToLower(strings.TrimSpace(s))
}
