package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/costscope/costscope/internal/core/config"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/multitenant"
	"github.com/costscope/costscope/internal/database"
	"github.com/costscope/costscope/internal/database/focus"
)

// QueryExecutor is a minimal contract satisfied by DuckDBEngine and fakes in tests.
// It avoids pulling the full AnalyticsEngine interface into this thin facade.
type QueryExecutor interface {
	ExecuteQuery(ctx context.Context, query string) (*database.QueryResult, error)
}

// Facade provides a small, opinionated analytics surface built on top of QueryBuilder + DuckDB.
// It is intentionally narrow and maps common analytics use cases directly to SQL via the builder.
//
// NOTE: This type is referenced by:
//   - CLI enhanced analytics commands (see cmd/analyze_enhanced.go & cmd/modules/api/handlers_analytics_duckdb.go)
//   - API handlers under internal/api/handlers/analytics_facade_duckdb.go (duckdb build tag)
//   - Unit tests in this package.
//
// Deadcode tools may mis-classify it in community builds lacking the duckdb tag; keep it.
type Facade struct {
	exec   QueryExecutor
	logger *logging.Logger
	cfg    *config.ConsolidatedConfig // optional; when nil tenant injection disabled implicitly
}

// NewFacade creates a new analytics facade.
// NewFacade retains backward compatibility (no config awareness => no automatic tenant injection).
func NewFacade(exec QueryExecutor) *Facade { return NewFacadeWithConfig(exec, nil) }

// NewFacadeWithConfig allows passing consolidated config to enable multi-tenant auto scoping.
func NewFacadeWithConfig(exec QueryExecutor, cfg *config.ConsolidatedConfig) *Facade {
	return &Facade{
		exec:   exec,
		logger: logging.GetLogger().WithFields(map[string]interface{}{"component": "analytics_facade"}),
		cfg:    cfg,
	}
}

// TopServices returns the top services by effective cost, honoring provided filters.
func (f *Facade) TopServices(ctx context.Context, filters *database.AnalyticsFilters, limit int) ([]*database.ServiceCost, error) {
	// Auto-inject tenant scope if applicable (skeleton feature flag)
	if filters != nil && f.cfg != nil {
		multitenant.InjectTenantFilter(ctx, f.cfg, filters, multitenant.ContextResolver{Key: multitenant.ContextKeyTenantID})
	}
	qb, err := focus.BuildTopServicesByCost(limit, filters)
	if err != nil {
		return nil, fmt.Errorf("build top services query: %w", err)
	}
	query, _, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("build top services SQL: %w", err)
	}

	res, err := f.exec.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute top services: %w", err)
	}

	out := make([]*database.ServiceCost, 0, len(res.Data))
	now := time.Now().UTC()
	for _, row := range res.Data {
		out = append(out, &database.ServiceCost{
			ServiceName: str(row["service_name"]),
			Provider:    str(row["provider_id"]), // may be empty if not selected
			TotalCost:   f64(row["total_cost"]),
			Currency:    defaultCurrency(str(row["billing_currency"])),
			RecordCount: i64(row["resource_count"]),
			AverageCost: f64(row["avg_cost"]),
			CostTrend:   0, // not derived here
			LastUpdated: now,
		})
	}
	return out, nil
}

// CostTrends returns cost aggregated by time period at the requested granularity.
func (f *Facade) CostTrends(ctx context.Context, filters *database.AnalyticsFilters, granularity database.TimeGranularity) ([]*database.TrendData, error) {
	if filters != nil && f.cfg != nil {
		multitenant.InjectTenantFilter(ctx, f.cfg, filters, multitenant.ContextResolver{Key: multitenant.ContextKeyTenantID})
	}
	qb, err := focus.BuildCostTrendByTime(granularity, filters)
	if err != nil {
		return nil, fmt.Errorf("build cost trend query: %w", err)
	}
	query, _, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("build cost trend SQL: %w", err)
	}

	res, err := f.exec.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute cost trend: %w", err)
	}

	out := make([]*database.TrendData, 0, len(res.Data))
	for _, row := range res.Data {
		// time_period is typically a DATE_TRUNC expression rendered by the engine; try best-effort parse.
		ts := parseTime(str(row["time_period"]))
		out = append(out, &database.TrendData{
			Timestamp:   ts,
			Value:       f64(row["total_cost"]),
			Granularity: granularity,
			Metadata:    map[string]interface{}{},
		})
	}
	return out, nil
}

// CostSummary returns a summary across the selected scope using the builder's SelectCostMetrics.
func (f *Facade) CostSummary(ctx context.Context, filters *database.AnalyticsFilters) (*database.CostSummary, error) {
	qb := focus.NewFOCUSQueryBuilder()
	// Use non-chaining calls to preserve concrete type for provider-specific helpers
	qb.SelectCostMetrics()
	qb.From("focus_cost_data")

	// Apply available filters directly using FOCUS builder helpers.
	if filters != nil {
		if f.cfg != nil { // attempt tenant injection before applying other filters (order unimportant for AND)
			multitenant.InjectTenantFilter(ctx, f.cfg, filters, multitenant.ContextResolver{Key: multitenant.ContextKeyTenantID})
		}
		if len(filters.Providers) > 0 {
			for _, p := range filters.Providers {
				qb.FilterByProvider(p)
			}
		}
		if filters.StartDate != nil && filters.EndDate != nil {
			qb.FilterByDateRange(*filters.StartDate, *filters.EndDate)
		}
		if filters.MinCost != nil {
			qb.FilterByCostThreshold(*filters.MinCost)
		}
		if len(filters.Services) > 0 {
			qb.FilterByService(filters.Services...)
		}
		if len(filters.Regions) > 0 {
			qb.FilterByRegion(filters.Regions...)
		}
		if len(filters.Accounts) > 0 {
			qb.FilterByAccount(filters.Accounts...)
		}
		if filters.TenantID != "" {
			qb.FilterByTenant(filters.TenantID)
		}
	}

	query, _, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("build cost summary SQL: %w", err)
	}

	res, err := f.exec.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute cost summary: %w", err)
	}

	// Expect a single row of aggregates; handle empty gracefully.
	var row map[string]interface{}
	if len(res.Data) > 0 {
		row = res.Data[0]
	} else {
		row = map[string]interface{}{}
	}

	var period database.TimePeriod
	if filters != nil && filters.StartDate != nil && filters.EndDate != nil {
		period = database.TimePeriod{Start: filters.StartDate.UTC(), End: filters.EndDate.UTC()}
	}

	summary := &database.CostSummary{
		TotalCost:   f64(row["total_cost"]),
		Currency:    defaultCurrency(str(row["billing_currency"])),
		Period:      period,
		RecordCount: i64(row["record_count"]),
		AverageCost: f64(row["avg_cost"]),
		MedianCost:  f64(row["median_cost"]),
		MinCost:     f64(row["min_cost"]),
		MaxCost:     f64(row["max_cost"]),
		StandardDev: f64(row["cost_stddev"]),
		Percentiles: database.Percentiles{
			P50: f64(row["median_cost"]),
			P75: 0,
			P90: 0,
			P95: f64(row["p95_cost"]),
			P99: 0,
		},
	}

	return summary, nil
}

// --- helpers ---

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func f64(v interface{}) float64 {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case int:
		return float64(t)
	case string:
		var out float64
		_, _ = fmt.Sscanf(t, "%f", &out)
		return out
	default:
		return 0
	}
}

func i64(v interface{}) int64 {
	switch t := v.(type) {
	case nil:
		return 0
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	default:
		return 0
	}
}

func parseTime(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if ts, err := time.Parse(l, v); err == nil {
			return ts
		}
	}
	return time.Time{}
}

func defaultCurrency(cur string) string {
	if cur == "" {
		return "USD"
	}
	return cur
}
