package common

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNormalizeLabelArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected map[string]string
	}{
		{
			name: "basic key/value and key normalization",
			input: []interface{}{
				map[string]interface{}{"key": "Env", "value": "Prod"},
				map[string]interface{}{"key": "Team Name", "value": "Alpha"},
			},
			expected: map[string]string{
				"env":       "Prod",
				"team_name": "Alpha",
			},
		},
		{
			name: "fallback to name when key missing",
			input: []interface{}{
				map[string]interface{}{"name": "Project", "value": "Phoenix"},
			},
			expected: map[string]string{
				"project": "Phoenix",
			},
		},
		{
			name: "mixed value types and skip empty name",
			input: []interface{}{
				map[string]interface{}{"key": "NumF", "value": 1.5},
				map[string]interface{}{"key": "Flag", "value": true},
				map[string]interface{}{"key": "NumJ", "value": json.Number("42")},
				map[string]interface{}{"name": "", "value": "noop"}, // skipped
			},
			expected: map[string]string{
				"numf": "1.5",
				"flag": "true",
				"numj": "42",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeLabelArray(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("normalizeLabelArray() mismatch\nexpected: %#v\n     got: %#v", tc.expected, got)
			}
		})
	}
}

func TestNormalizeLabelMap(t *testing.T) {
	// prepare a complex object to validate default JSON marshaling branch
	complexObj := map[string]int{"a": 1}
	complexJSON, _ := json.Marshal(complexObj)

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name: "basic types and key normalization",
			input: map[string]interface{}{
				"Team Name": "Alpha",
				"Count":     float64(7),
				"ok":        true,
				"num":       json.Number("3.14"),
				"obj":       complexObj,
				" spaced ":  "trim_me",
			},
			expected: map[string]string{
				"team_name": "Alpha",
				"count":     "7",
				"ok":        "true",
				"num":       "3.14",
				"obj":       string(complexJSON),
				"spaced":    "trim_me",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeLabelMap(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("normalizeLabelMap() mismatch\nexpected: %#v\n     got: %#v", tc.expected, got)
			}
		})
	}
}
