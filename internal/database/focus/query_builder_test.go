package focus

import (
	"strings"
	"testing"
	"time"

	"local/costscope/internal/database"
)

// helper
func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	dt, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return dt
}

func TestBuildTopServicesByCost_SQLShape(t *testing.T) {
	limit := 5
	min := 10.0
	start := mustParseDate(t, "2025-01-01")
	end := mustParseDate(t, "2025-01-31")
	filters := &database.AnalyticsFilters{
		StartDate: &start,
		EndDate:   &end,
		Services:  []string{"EC2", "S3"},
		Regions:   []string{"us-east-1"},
		MinCost:   &min,
		Accounts:  []string{"111111111111", "222222222222"},
		TenantID:  "t-001",
	}

	qb, err := BuildTopServicesByCost(limit, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// core shape
	mustContain(t, sql, "SELECT service_name, SUM(effective_cost) as total_cost, COUNT(*) as resource_count FROM focus_cost_data")
	mustContain(t, sql, "GROUP BY service_name")
	mustContain(t, sql, "ORDER BY total_cost DESC")
	mustContain(t, sql, "LIMIT 5")
	// filters
	mustContain(t, sql, "charge_period_start >= '2025-01-01' AND charge_period_end <= '2025-01-31'")
	mustContain(t, sql, "service_name IN ('EC2','S3')")
	mustContain(t, sql, "region = 'us-east-1'")
	mustContain(t, sql, "effective_cost >= 10.000000")
	mustContain(t, sql, "billing_account_id IN ('111111111111','222222222222')")
	mustContain(t, sql, "tenant_id = 't-001'")
}

func TestBuildCostTrendByTime_AllGranularities(t *testing.T) {
	cases := []struct {
		name        string
		g           database.TimeGranularity
		expectExpr  string
		expectOrder string
	}{
		{"hour", database.TimeGranularityHour, "DATE_TRUNC('hour', charge_period_start) as time_period", "ORDER BY time_period ASC"},
		{"day", database.TimeGranularityDay, "DATE_TRUNC('day', charge_period_start) as time_period", "ORDER BY time_period ASC"},
		{"week", database.TimeGranularityWeek, "DATE_TRUNC('week', charge_period_start) as time_period", "ORDER BY time_period ASC"},
		{"month", database.TimeGranularityMonth, "DATE_TRUNC('month', charge_period_start) as time_period", "ORDER BY time_period ASC"},
		{"year", database.TimeGranularityYear, "DATE_TRUNC('year', charge_period_start) as time_period", "ORDER BY time_period ASC"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qb, err := BuildCostTrendByTime(tc.g, nil)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			sql, _, err := qb.Build()
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			mustContain(t, sql, "FROM focus_cost_data")
			mustContain(t, sql, tc.expectExpr)
			mustContain(t, sql, "SUM(effective_cost) as total_cost")
			mustContain(t, sql, "GROUP BY ")
			mustContain(t, sql, tc.expectOrder)
		})
	}
}

func TestFilterPredicates_EQ_IN(t *testing.T) {
	qb := NewFOCUSQueryBuilder()
	qb.Select("*").From("focus_cost_data")
	qb.FilterByService("only-one")
	qb.FilterByRegion("r1", "r2")
	qb.FilterByAccount("a1")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	mustContain(t, sql, "service_name = 'only-one'")
	mustContain(t, sql, "region IN ('r1','r2')")
	mustContain(t, sql, "billing_account_id = 'a1'")
}

// Prebuilt provider/region helpers were removed; test coverage retained for other builders.

// small assert helper
func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected SQL to contain %q, got: %s", sub, s)
	}
}
