package focus

import (
	"strings"
	"testing"
	"time"

	"local/costscope/internal/database"
)

// local helper (duplicated intentionally; kept private to avoid exporting test-only utils)
func mustContain2(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected SQL to contain %q, got: %s", sub, s)
	}
}

// BasicSelectFilters ensures minimal SELECT * + provider filter works.
func TestBasicSelectFilters(t *testing.T) {
	qb := NewFOCUSQueryBuilder().Select("*").From("focus_cost_data").FilterByProvider("aws")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	mustContain2(t, sql, "SELECT * FROM focus_cost_data")
	mustContain2(t, sql, "provider_id = 'aws'")
}

// AggregationGroupOrder validates cost metrics selection, grouping and ordering.
func TestAggregationGroupOrder(t *testing.T) {
	qb := NewFOCUSQueryBuilder().From("focus_cost_data").Select("service_name").SelectCostMetrics().GroupBy("service_name").OrderBy("total_cost", database.SortDirectionDesc)
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	mustContain2(t, sql, "SUM(effective_cost) as total_cost")
	mustContain2(t, sql, "GROUP BY service_name")
	mustContain2(t, sql, "ORDER BY total_cost DESC")
}

// DateRangeAndProviderComposite validates both date range and provider predicates appear together.
func TestDateRangeAndProviderComposite(t *testing.T) {
	start, _ := time.Parse("2006-01-02", "2025-02-01")
	end, _ := time.Parse("2006-01-02", "2025-02-28")
	qb := NewFOCUSQueryBuilder().From("focus_cost_data").FilterByDateRange(start, end).FilterByProvider("azure")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	mustContain2(t, sql, "charge_period_start >= '2025-02-01' AND charge_period_end <= '2025-02-28'")
	mustContain2(t, sql, "provider_id = 'azure'")
}

// MultiValueServiceRegion checks IN predicate generation for services & regions.
func TestMultiValueServiceRegion(t *testing.T) {
	qb := NewFOCUSQueryBuilder().From("focus_cost_data").FilterByService("EC2", "S3", "RDS").FilterByRegion("us-east-1", "us-west-2")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	mustContain2(t, sql, "service_name IN ('EC2','S3','RDS')")
	mustContain2(t, sql, "region IN ('us-east-1','us-west-2')")
}

// CostThresholdStacking ensures min & max cost constraints coexist.
func TestCostThresholdStacking(t *testing.T) {
	qb := NewFOCUSQueryBuilder().From("focus_cost_data").FilterByCostThreshold(15.5).Where("effective_cost <= %f", 200.0)
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	mustContain2(t, sql, "effective_cost >= 15.500000")
	mustContain2(t, sql, "effective_cost <= 200.000000")
}

// IdempotentDomainFilters documents current non-idempotent behavior (duplicate predicates) when applying the same filter twice.
// This test intentionally asserts duplication to capture current behavior; if future refactor makes filters idempotent,
// adjust expectation accordingly (and possibly remove duplication at build time).
func TestIdempotentDomainFilters_CurrentDuplicateBehavior(t *testing.T) {
	qb := NewFOCUSQueryBuilder().From("focus_cost_data").FilterByService("Lambda").FilterByService("Lambda")
	sql, _, err := qb.Build()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	// Expect service_name = 'Lambda' to appear twice
	occurrences := strings.Count(sql, "service_name = 'Lambda'")
	if occurrences != 2 { // document current state
		t.Fatalf("expected duplicate service filter (2 occurrences), got %d. SQL: %s", occurrences, sql)
	}
}

// ErrorPaths validates missing FROM triggers error.
func TestErrorPaths_MissingFrom(t *testing.T) {
	qb := NewFOCUSQueryBuilder().Select("*") // no From
	_, _, err := qb.Build()
	if err == nil || !strings.Contains(err.Error(), "FROM table is required") {
		t.Fatalf("expected FROM table error, got: %v", err)
	}
}
