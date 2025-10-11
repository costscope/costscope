package focus

import "testing"

// TestFluentChain_LimitOffset ensures a single fluent chain exercising
// Select -> From -> multiple domain filters -> GroupBy -> OrderBy -> Limit -> Offset
// assembles correctly and includes both LIMIT and OFFSET clauses. This smoke test
// closes the small remaining gap (explicit Limit+Offset assertion) not covered
// by other shape tests, guarding against future regressions where one of the
// terminal clauses might be skipped during refactors.
func TestFluentChain_LimitOffset(t *testing.T) {
	qb := NewFOCUSQueryBuilder().
		Select("service_name").
		SelectCostMetrics().
		From("focus_cost_data").
		FilterByProvider("aws").
		FilterByService("EC2", "S3").
		FilterByRegion("us-east-1").
		GroupBy("service_name").
		OrderBy("total_cost", "DESC").
		Limit(25).
		Offset(50)

	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	mustContain := func(sub string) {
		if !contains(sql, sub) { // reuse lightweight local contains from tenant test if present
			t.Fatalf("expected SQL to contain %q, got: %s", sub, sql)
		}
	}

	mustContain("SELECT service_name, SUM(effective_cost) as total_cost")
	mustContain("FROM focus_cost_data")
	mustContain("provider_id = 'aws'")
	mustContain("service_name IN ('EC2','S3')")
	mustContain("region = 'us-east-1'")
	mustContain("GROUP BY service_name")
	mustContain("ORDER BY total_cost DESC")
	mustContain("LIMIT 25")
	mustContain("OFFSET 50")
}
