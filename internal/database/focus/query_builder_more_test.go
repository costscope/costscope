//go:build qb_extended

package focus

import (
	"strings"
	"testing"

	"local/costscope/internal/database"
)

// Test CTE, Join/LeftJoin, Having, Limit/Offset paths and overall SQL shape.
func TestQueryBuilder_CTE_Joins_Having_LimitOffset(t *testing.T) {
	qb := NewFOCUSQueryBuilder()
	// Use concrete type qb for chaining; explicit type assertions only when needed.
	qb = qb.WithCTE("base", "SELECT * FROM focus_cost_data WHERE effective_cost > 0").(*FOCUSQueryBuilder)
	qb = qb.Select("service_name", "SUM(effective_cost) as total_cost").(*FOCUSQueryBuilder)
	qb = qb.From("base").(*FOCUSQueryBuilder)
	qb = qb.Join("services s", "s.name = service_name").(*FOCUSQueryBuilder)
	qb = qb.LeftJoin("regions r", "r.id = region").(*FOCUSQueryBuilder)
	qb = qb.Where("provider_id = '%s'", "aws").(*FOCUSQueryBuilder)
	qb = qb.GroupBy("service_name").(*FOCUSQueryBuilder)
	qb = qb.Having("SUM(effective_cost) > %d", 100).(*FOCUSQueryBuilder)
	qb = qb.OrderBy("total_cost", database.SortDirectionDesc).(*FOCUSQueryBuilder)
	qb = qb.Limit(10).(*FOCUSQueryBuilder)
	qb = qb.Offset(5).(*FOCUSQueryBuilder)

	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	mustContain(t, sql, "WITH base AS (SELECT * FROM focus_cost_data WHERE effective_cost > 0)")
	mustContain(t, sql, "SELECT service_name, SUM(effective_cost) as total_cost FROM base")
	mustContain(t, sql, "JOIN services s ON s.name = service_name")
	mustContain(t, sql, "LEFT JOIN regions r ON r.id = region")
	mustContain(t, sql, "WHERE provider_id = 'aws'")
	mustContain(t, sql, "GROUP BY service_name")
	mustContain(t, sql, "HAVING SUM(effective_cost) > 100")
	mustContain(t, sql, "ORDER BY total_cost DESC")
	mustContain(t, sql, "LIMIT 10")
	mustContain(t, sql, "OFFSET 5")
}

// Test BuildCount and BuildExplain helpers are wired and produce expected prefixes.
func TestQueryBuilder_BuildCount_And_Explain(t *testing.T) {
	qb := NewFOCUSQueryBuilder()
	qb = qb.From("focus_cost_data").(*FOCUSQueryBuilder)
	qb = qb.Where("region = '%s'", "us-east-1").(*FOCUSQueryBuilder)

	countSQL, _, err := qb.BuildCount()
	if err != nil {
		t.Fatalf("build count: %v", err)
	}
	mustContain(t, countSQL, "SELECT COUNT(*) as total_count FROM focus_cost_data WHERE region = 'us-east-1'")

	explainSQL, _, err := qb.BuildExplain()
	if err != nil {
		t.Fatalf("build explain: %v", err)
	}
	if !strings.HasPrefix(explainSQL, "EXPLAIN ") {
		t.Fatalf("expected EXPLAIN prefix, got: %s", explainSQL)
	}
}
