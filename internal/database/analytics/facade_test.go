package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/database"

	"github.com/stretchr/testify/require"
)

type fakeExec struct {
	res *database.QueryResult
	err error
}

func (f *fakeExec) ExecuteQuery(ctx context.Context, query string) (*database.QueryResult, error) {
	return f.res, f.err
}

func TestFacadeTopServices_basic(t *testing.T) {
	f := NewFacade(&fakeExec{res: &database.QueryResult{
		Columns: []string{"service_name", "total_cost", "resource_count"},
		Data: []map[string]interface{}{
			{"service_name": "EC2", "total_cost": 123.45, "resource_count": int64(10)},
			{"service_name": "S3", "total_cost": 67.89, "resource_count": int64(5)},
		},
	}})

	ctx := context.Background()
	items, err := f.TopServices(ctx, nil, 5)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "EC2", items[0].ServiceName)
	require.InDelta(t, 123.45, items[0].TotalCost, 1e-9)
	require.EqualValues(t, int64(10), items[0].RecordCount)
	require.False(t, items[0].LastUpdated.IsZero())
}

func TestFacadeCostTrends_basic(t *testing.T) {
	f := NewFacade(&fakeExec{res: &database.QueryResult{
		Columns: []string{"time_period", "total_cost"},
		Data: []map[string]interface{}{
			{"time_period": "2025-01-01", "total_cost": 10.0},
			{"time_period": "2025-01-02", "total_cost": 15.5},
		},
	}})
	ctx := context.Background()
	items, err := f.CostTrends(ctx, nil, database.TimeGranularityDay)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, 10.0, items[0].Value)
	require.Equal(t, database.TimeGranularityDay, items[0].Granularity)
}

func TestFacadeCostSummary_basic(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	filters := &database.AnalyticsFilters{StartDate: &start, EndDate: &end}
	f := NewFacade(&fakeExec{res: &database.QueryResult{
		Columns: []string{"total_cost", "avg_cost", "record_count", "median_cost", "min_cost", "max_cost", "cost_stddev", "p95_cost", "billing_currency"},
		Data: []map[string]interface{}{
			{"total_cost": 1000.0, "avg_cost": 10.0, "record_count": int64(100), "median_cost": 9.0, "min_cost": 1.0, "max_cost": 20.0, "cost_stddev": 2.5, "p95_cost": 18.0, "billing_currency": "USD"},
		},
	}})
	ctx := context.Background()
	sum, err := f.CostSummary(ctx, filters)
	require.NoError(t, err)
	require.InDelta(t, 1000.0, sum.TotalCost, 1e-9)
	require.Equal(t, "USD", sum.Currency)
	require.EqualValues(t, int64(100), sum.RecordCount)
	require.InDelta(t, 18.0, sum.Percentiles.P95, 1e-9)
	require.Equal(t, start.UTC(), sum.Period.Start)
	require.Equal(t, end.UTC(), sum.Period.End)
}
