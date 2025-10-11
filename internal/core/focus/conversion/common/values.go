package common

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Truthy string constants reused across helpers
const (
	StrTrue  = "true"
	StrFalse = "false"
)

// ParseTimeAny parses common timestamp formats; returns zero time on failure.
func ParseTimeAny(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// GetStringPath returns a string from nested map path like "a.b.c".
func GetStringPath(m map[string]interface{}, path string) string {
	v, ok := GetPath(m, path)
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case bool:
		if t {
			return StrTrue
		}
		return StrFalse
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// GetFloatPath reads a float64 from a nested map path.
func GetFloatPath(m map[string]interface{}, path string) float64 {
	v, ok := GetPath(m, path)
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return f
		}
	}
	return 0
}

// GetPath navigates a nested map using dot-separated keys.
func GetPath(m map[string]interface{}, path string) (interface{}, bool) {
	cur := interface{}(m)
	for _, part := range strings.Split(path, ".") {
		mm, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		v, exists := mm[part]
		if !exists {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

// ExtractLabels reads labels from either array of {key,value} or object map.
func ExtractLabels(obj map[string]interface{}, path string) map[string]string {
	out := map[string]string{}
	v, ok := GetPath(obj, path)
	if !ok || v == nil {
		return out
	}
	switch t := v.(type) {
	case []interface{}:
		for k, v := range normalizeLabelArray(t) {
			out[k] = v
		}
	case map[string]interface{}:
		for k, v := range normalizeLabelMap(t) {
			out[k] = v
		}
	case string:
		// labels may be stringified JSON in CSV exports
		for k, v := range ParseLabelsJSON(t) {
			out[NormalizeTagKey(k)] = v
		}
	}
	return out
}

// ParseLabelsJSON parses labels from a string that may contain JSON array/object.
func ParseLabelsJSON(s string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(s) == "" {
		return out
	}
	var any interface{}
	if err := json.Unmarshal([]byte(s), &any); err != nil {
		return out
	}
	switch t := any.(type) {
	case []interface{}:
		for k, v := range normalizeLabelArray(t) {
			out[k] = v
		}
	case map[string]interface{}:
		for k, v := range normalizeLabelMap(t) {
			out[k] = v
		}
	}
	return out
}

// normalizeLabelArray converts an array of objects with key/name and value fields
// to a normalized map[string]string with normalized keys.
func normalizeLabelArray(arr []interface{}) map[string]string {
	out := make(map[string]string, len(arr))
	for _, it := range arr {
		if mm, ok := it.(map[string]interface{}); ok {
			k := GetStringPath(mm, "key")
			if k == "" {
				k = GetStringPath(mm, "name")
			}
			val := GetStringPath(mm, "value")
			if k != "" {
				out[NormalizeTagKey(k)] = val
			}
		}
	}
	return out
}

// normalizeLabelMap converts a generic object map to string labels with normalized keys.
func normalizeLabelMap(m map[string]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, vv := range m {
		switch tv := vv.(type) {
		case string:
			out[NormalizeTagKey(k)] = tv
		case json.Number:
			out[NormalizeTagKey(k)] = tv.String()
		case float64:
			out[NormalizeTagKey(k)] = strconv.FormatFloat(tv, 'f', -1, 64)
		case bool:
			if tv {
				out[NormalizeTagKey(k)] = StrTrue
			} else {
				out[NormalizeTagKey(k)] = StrFalse
			}
		default:
			b, _ := json.Marshal(tv)
			out[NormalizeTagKey(k)] = string(b)
		}
	}
	return out
}

// MergeLabels copies all src key/values into dst, overwriting on conflict.
func MergeLabels(dst map[string]string, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

// NormalizeTagKey lower-cases, trims and replaces spaces with underscores.
func NormalizeTagKey(k string) string {
	k = strings.TrimSpace(strings.ToLower(k))
	k = strings.ReplaceAll(k, " ", "_")
	return k
}
