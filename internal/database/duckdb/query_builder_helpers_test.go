//go:build duckdb && qb_extended

package duckdb

import (
	"strings"
	"testing"

	"local/costscope/internal/database"
)

// TestDuckDBQueryBuilder_HelperDelegation ensures the shared common helpers (ORDER BY, LEFT JOIN, cost threshold)
// integrate correctly when invoked through the DuckDBQueryBuilder surface so future refactors that inline or
// remove helpers will break in a single obvious place.
func TestDuckDBQueryBuilder_HelperDelegation(t *testing.T) {
	qb := NewDuckDBQueryBuilder()
	qb = qb.Select("service_name").(*DuckDBQueryBuilder)
	qb = qb.From("focus_cost_data").(*DuckDBQueryBuilder)
	qb = qb.LeftJoin("regions r", "r.id = region").(*DuckDBQueryBuilder)
	qb = qb.Where("provider_name = '%s'", "aws").(*DuckDBQueryBuilder)
	qb = qb.FilterByCostThreshold(25.0).(*DuckDBQueryBuilder)
	qb = qb.GroupBy("service_name").(*DuckDBQueryBuilder)
	qb = qb.OrderBy("service_name", database.SortDirectionAsc).(*DuckDBQueryBuilder)

	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	assertContains(t, sql, "LEFT JOIN regions r ON r.id = region")
	assertContains(t, sql, "effective_cost >= 25.000000")
	assertContains(t, sql, "ORDER BY service_name ASC")
}

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected %q in SQL: %s", sub, s)
	}
}
