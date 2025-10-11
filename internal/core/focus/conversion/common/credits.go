package common

import (
	"encoding/json"
	"strings"
)

// ParseCredits parses a credits JSON string from GCP exports.
// Accepts either an array of objects or a single object. Returns first credit's
// id, type, name and ok=true if parsed.
func ParseCredits(s string) (string, string, string, bool) {
	if strings.TrimSpace(s) == "" {
		return "", "", "", false
	}
	var any interface{}
	if err := json.Unmarshal([]byte(s), &any); err != nil {
		return "", "", "", false
	}
	switch t := any.(type) {
	case []interface{}:
		for _, it := range t {
			if m, ok := it.(map[string]interface{}); ok {
				cID := GetStringPath(m, "id")
				cType := GetStringPath(m, "type")
				cName := GetStringPath(m, "name")
				return cID, cType, cName, cID != "" || cType != "" || cName != ""
			}
		}
	case map[string]interface{}:
		cID := GetStringPath(t, "id")
		cType := GetStringPath(t, "type")
		cName := GetStringPath(t, "name")
		return cID, cType, cName, cID != "" || cType != "" || cName != ""
	}
	return "", "", "", false
}

// ParseCreditsUnified handles credits value provided as stringified JSON, array, or map.
// Returns (id, type, name, isCredit, isSpot)
func ParseCreditsUnified(v interface{}) (string, string, string, bool, bool) {
	// classify returns whether credit looks like spot/preempt
	classify := func(typ, name string) bool {
		lt := strings.ToLower(typ)
		ln := strings.ToLower(name)
		if strings.Contains(lt, "spot") || strings.Contains(ln, "spot") {
			return true
		}
		if strings.Contains(lt, "preempt") || strings.Contains(ln, "preempt") {
			return true
		}
		return false
	}

	switch t := v.(type) {
	case string:
		if id, typ, name, ok := ParseCredits(t); ok {
			return id, typ, name, true, classify(typ, name)
		}
		return "", "", "", false, false
	case []interface{}:
		var firstID, firstType, firstName string
		isSpot := false
		found := false
		for _, it := range t {
			if m, ok := it.(map[string]interface{}); ok {
				id := GetStringPath(m, "id")
				typ := GetStringPath(m, "type")
				name := GetStringPath(m, "name")
				if !found && (id != "" || typ != "" || name != "") {
					firstID, firstType, firstName = id, typ, name
					found = true
				}
				if classify(typ, name) {
					isSpot = true
				}
			}
		}
		return firstID, firstType, firstName, found, isSpot
	case map[string]interface{}:
		id := GetStringPath(t, "id")
		typ := GetStringPath(t, "type")
		name := GetStringPath(t, "name")
		if id != "" || typ != "" || name != "" {
			return id, typ, name, true, classify(typ, name)
		}
		return "", "", "", false, false
	default:
		return "", "", "", false, false
	}
}
