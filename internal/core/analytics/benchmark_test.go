package analytics

import (
	"testing"

	analyticsTypes "github.com/costscope/costscope/cmd/modules/analytics/types"
	"github.com/costscope/costscope/internal/core/logging"
)

func BenchmarkAnalyzeBasic(b *testing.B) {
	cfg := &Config{MLEnabled: true}
	logger := logging.NewLogger("error")
	svc := NewBasicService(cfg, logger)

	opts := &analyticsTypes.AnalyticsOptions{
		TableName:     "focus_costs",
		Currency:      "USD",
		EnableML:      true,
		Filters:       map[string]interface{}{"region": "us-east-1"},
		GroupByFields: []string{"service", "region"},
		SortOrder:     "desc",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Analyze(opts); err != nil {
			b.Fatalf("Analyze failed: %v", err)
		}
	}
}
