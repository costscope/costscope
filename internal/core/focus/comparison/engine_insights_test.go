package comparison_test

import (
	"testing"
	"time"

	comparison "github.com/costscope/costscope/internal/core/focus/comparison"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestComparisonEngine_GenerateComparisonInsights(t *testing.T) {
	eng := comparison.NewEngine(logging.NewLogger(logging.LevelWarn), nil)
	start := time.Now().AddDate(0, 0, -10)

	// baseline (service A), current (service A with growth + new service B)
	base := []comparison.FOCUSRecord{}
	for i := 0; i < 5; i++ {
		d := start.AddDate(0, 0, i)
		base = append(base, comparison.FOCUSRecord{BillingAccountID: "acct", BillingAccountName: "acct", BillingCurrency: "USD", BillingPeriodStart: d, BillingPeriodEnd: d.Add(24 * time.Hour), ServiceName: "Compute", Region: "us-east-1", UsageQuantity: 1, UsageUnit: "Hours", BilledCost: 100 + float64(i)})
	}
	cur := []comparison.FOCUSRecord{}
	for i := 0; i < 5; i++ {
		d := start.AddDate(0, 0, i)
		cur = append(cur, comparison.FOCUSRecord{BillingAccountID: "acct", BillingAccountName: "acct", BillingCurrency: "USD", BillingPeriodStart: d, BillingPeriodEnd: d.Add(24 * time.Hour), ServiceName: "Compute", Region: "us-east-1", UsageQuantity: 1, UsageUnit: "Hours", BilledCost: 150 + float64(i)})
	}
	// new service
	for i := 0; i < 5; i++ {
		d := start.AddDate(0, 0, i)
		cur = append(cur, comparison.FOCUSRecord{BillingAccountID: "acct", BillingAccountName: "acct", BillingCurrency: "USD", BillingPeriodStart: d, BillingPeriodEnd: d.Add(24 * time.Hour), ServiceName: "Storage", Region: "us-east-1", UsageQuantity: 1, UsageUnit: "GB", BilledCost: 10 + float64(i)})
	}

	opts := comparison.DiffOptions{Dimensions: []string{"service", "region"}, Threshold: 1, ShowAnomalies: true, ShowTrends: true, MLEnabled: true}
	insights, err := eng.GenerateComparisonInsights(base, cur, opts, 3)
	if err != nil {
		t.Fatalf("GenerateComparisonInsights error: %v", err)
	}
	if insights.Diff == nil || insights.Executive == nil {
		t.Fatalf("expected diff & executive populated")
	}
	if len(insights.Forecast) > 0 && len(insights.Forecast) != 3 {
		t.Fatalf("expected 3 forecast periods, got %d", len(insights.Forecast))
	}
	if insights.Executive.TotalCostImpact == 0 {
		t.Fatalf("expected non-zero cost impact in executive summary")
	}
}
