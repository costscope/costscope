package focus

import "testing"

// Tests for FilterByTenant ensuring empty tenant is ignored and non-empty applied.

func TestFilterByTenant_IgnoresEmpty(t *testing.T) {
	qb := NewFOCUSQueryBuilder()
	qb.Select("*")
	qb.From("focus_cost_data")
	qb.FilterByTenant("")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("unexpected error building SQL: %v", err)
	}
	if contains(sql, "tenant_id =") {
		t.Fatalf("expected no tenant filter in SQL, got: %s", sql)
	}
}

func TestFilterByTenant_AddsCondition(t *testing.T) {
	qb := NewFOCUSQueryBuilder()
	qb.Select("*")
	qb.From("focus_cost_data")
	qb.FilterByTenant("tenant-abc")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("unexpected error building SQL: %v", err)
	}
	if !contains(sql, "tenant_id = 'tenant-abc'") {
		t.Fatalf("expected tenant filter in SQL, got: %s", sql)
	}
}

// local helper (avoid importing strings just for Contains to keep test tiny)
func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
