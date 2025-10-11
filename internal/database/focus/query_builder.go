package focus

import (
	"fmt"
	"strings"
	"time"

	"local/costscope/internal/core/monitoring/telemetry"
	"local/costscope/internal/database"
	dbcommon "local/costscope/internal/database/common"
)

// FOCUSQueryBuilder provides specialized query building for FOCUS data
type FOCUSQueryBuilder struct {
	selectCols []string
	fromTable  string
	joins      []string // extended when tag enabled (still stored for compatibility)
	conditions []string
	groupBy    []string
	orderBy    []string
	having     []string // extended
	limit      *int
	offset     *int
	ctes       []string // extended
}

// internal helpers (keep unexported; consolidate repeated clause builders)
const (
	focusTableName     = "focus_cost_data"
	selTotalCost       = "SUM(effective_cost) as total_cost"
	selResourceCount   = "COUNT(*) as resource_count"
	orderTotalCostDesc = "total_cost DESC"
	orderTimePeriodAsc = "time_period ASC"
	truncHour          = "DATE_TRUNC('hour', charge_period_start)"
	truncDay           = "DATE_TRUNC('day', charge_period_start)" //nolint:goconst // used intentionally
	truncWeek          = "DATE_TRUNC('week', charge_period_start)"
	truncMonth         = "DATE_TRUNC('month', charge_period_start)"
	truncYear          = "DATE_TRUNC('year', charge_period_start)"
)

// timeGroupExpr returns the DATE_TRUNC(...) expression for a given granularity.
func timeGroupExpr(granularity database.TimeGranularity) string {
	switch granularity {
	case database.TimeGranularityHour:
		return truncHour
	case database.TimeGranularityDay:
		return truncDay //nolint:goconst // SQL time grouping constant
	case database.TimeGranularityWeek:
		return truncWeek
	case database.TimeGranularityMonth:
		return truncMonth
	case database.TimeGranularityYear:
		return truncYear
	default:
		return truncDay //nolint:goconst // SQL time grouping constant
	}
}

// NOTE: Removed local eqOrInCondition duplicate; we now delegate to shared dbcommon.EqOrIn
// for consistency across builders (focus & duckdb) and single test surface.

// NewFOCUSQueryBuilder creates a new FOCUS query builder
func NewFOCUSQueryBuilder() *FOCUSQueryBuilder {
	return &FOCUSQueryBuilder{
		selectCols: []string{},
		conditions: []string{},
		joins:      []string{},
		groupBy:    []string{},
		orderBy:    []string{},
		having:     []string{},
		ctes:       []string{},
	}
}

// Select adds columns to SELECT clause
func (qb *FOCUSQueryBuilder) Select(columns ...string) database.QueryBuilder {
	qb.selectCols = append(qb.selectCols, columns...)
	return qb
}

// From sets the main table
func (qb *FOCUSQueryBuilder) From(table string) database.QueryBuilder {
	qb.fromTable = table
	return qb
}

// Where adds a WHERE condition
func (qb *FOCUSQueryBuilder) Where(condition string, args ...interface{}) database.QueryBuilder {
	if len(args) > 0 {
		// Handle parameterized queries
		qb.conditions = append(qb.conditions, fmt.Sprintf(condition, args...))
	} else {
		qb.conditions = append(qb.conditions, condition)
	}
	return qb
}

// GroupBy adds GROUP BY columns
func (qb *FOCUSQueryBuilder) GroupBy(columns ...string) database.QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// OrderBy adds ORDER BY clause
// nolint:dupl // Intentional duplication across builders; delegates to shared dbcommon helper
func (qb *FOCUSQueryBuilder) OrderBy(column string, direction database.SortDirection) database.QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// Limit sets the LIMIT clause
func (qb *FOCUSQueryBuilder) Limit(count int) database.QueryBuilder {
	qb.limit = &count
	return qb
}

// Offset sets the OFFSET clause
func (qb *FOCUSQueryBuilder) Offset(count int) database.QueryBuilder {
	qb.offset = &count
	return qb
}

// Join adds a JOIN clause
// Extended operations (Join/LeftJoin/Having/WithCTE) live in qb_extended build tag file.

// FOCUS-specific methods

// SelectCostMetrics adds common cost metric columns
func (qb *FOCUSQueryBuilder) SelectCostMetrics() database.QueryBuilder {
	metrics := []string{
		"SUM(effective_cost) as total_cost",
		"AVG(effective_cost) as avg_cost",
		"MIN(effective_cost) as min_cost",
		"MAX(effective_cost) as max_cost",
		"COUNT(*) as record_count",
		"STDDEV(effective_cost) as cost_stddev",
		"PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY effective_cost) as median_cost",
		"PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY effective_cost) as p95_cost",
	}
	qb.selectCols = append(qb.selectCols, metrics...)
	return qb
}

// FilterByProvider adds provider filter
func (qb *FOCUSQueryBuilder) FilterByProvider(provider string) database.QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("provider_id = '%s'", provider))
	return qb
}

// FilterByDateRange adds date range filter
func (qb *FOCUSQueryBuilder) FilterByDateRange(start, end time.Time) database.QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf(
		"charge_period_start >= '%s' AND charge_period_end <= '%s'",
		start.Format("2006-01-02"), end.Format("2006-01-02"),
	))
	return qb
}

// FilterByCostThreshold adds cost threshold filter
// nolint:dupl // Intentional duplication across builders; delegates to shared dbcommon helper
func (qb *FOCUSQueryBuilder) FilterByCostThreshold(threshold float64) database.QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("effective_cost >= %f", threshold))
	return qb
}

// FilterByService adds service filter
func (qb *FOCUSQueryBuilder) FilterByService(services ...string) database.QueryBuilder {
	if cond := dbcommon.EqOrIn("service_name", services); cond != "" {
		qb.conditions = append(qb.conditions, cond)
	}
	return qb
}

// FilterByRegion adds region filter
func (qb *FOCUSQueryBuilder) FilterByRegion(regions ...string) database.QueryBuilder {
	if cond := dbcommon.EqOrIn("region", regions); cond != "" {
		qb.conditions = append(qb.conditions, cond)
	}
	return qb
}

// FilterByAccount adds account filter
func (qb *FOCUSQueryBuilder) FilterByAccount(accounts ...string) database.QueryBuilder {
	if cond := dbcommon.EqOrIn("billing_account_id", accounts); cond != "" {
		qb.conditions = append(qb.conditions, cond)
	}
	return qb
}

// FilterByTenant adds tenant scoping filter (expects column tenant_id). Empty tenant ignored.
func (qb *FOCUSQueryBuilder) FilterByTenant(tenantID string) database.QueryBuilder {
	if tenantID != "" {
		qb.conditions = append(qb.conditions, fmt.Sprintf("tenant_id = '%s'", tenantID))
	}
	return qb
}

// GroupByTime previously existed as a convenience method but was unused; prefer
// prebuilt helpers like BuildCostTrendByTime which apply deterministic ordering
// and shape. Intentionally removed to keep the builder surface minimal.

// Build generates the final SQL query
func (qb *FOCUSQueryBuilder) Build() (string, []interface{}, error) {
	// Increment build counter (focus variant). Safe even if telemetry not registered yet.
	// We call WithLabelValues lazily; if Register() not invoked, collector registration will happen later.
	telemetry.QueryBuilderBuilds.WithLabelValues("focus").Inc()
	var sql strings.Builder

	// WITH clauses (CTEs)
	if len(qb.ctes) > 0 {
		sql.WriteString("WITH ")
		sql.WriteString(strings.Join(qb.ctes, ", "))
		sql.WriteString(" ")
	}

	// SELECT clause
	sql.WriteString("SELECT ")
	if len(qb.selectCols) == 0 {
		sql.WriteString("*")
	} else {
		sql.WriteString(strings.Join(qb.selectCols, ", "))
	}

	// FROM clause
	if qb.fromTable == "" {
		return "", nil, fmt.Errorf("FROM table is required")
	}
	sql.WriteString(fmt.Sprintf(" FROM %s", qb.fromTable))

	// JOIN clauses
	if len(qb.joins) > 0 {
		sql.WriteString(" ")
		sql.WriteString(strings.Join(qb.joins, " "))
	}

	// WHERE clause
	if len(qb.conditions) > 0 {
		sql.WriteString(" WHERE ")
		sql.WriteString(strings.Join(qb.conditions, " AND "))
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING clause
	if len(qb.having) > 0 {
		sql.WriteString(" HAVING ")
		sql.WriteString(strings.Join(qb.having, " AND "))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		sql.WriteString(" ORDER BY ")
		sql.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT clause
	if qb.limit != nil {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", *qb.limit))
	}

	// OFFSET clause
	if qb.offset != nil {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", *qb.offset))
	}

	return sql.String(), []interface{}{}, nil
}

// BuildCount generates a count query
// Extended variants BuildCount / BuildExplain live under qb_extended tag.

// Pre-built FOCUS queries

// BuildTopServicesByCost creates a query for top services by cost
func BuildTopServicesByCost(limit int, filters *database.AnalyticsFilters) (*FOCUSQueryBuilder, error) {
	qb := NewFOCUSQueryBuilder()
	qb.fromTable = focusTableName //nolint:goconst // FOCUS table name
	qb.selectCols = append(qb.selectCols, "service_name")
	qb.selectCols = append(qb.selectCols, selTotalCost)
	qb.selectCols = append(qb.selectCols, selResourceCount)
	qb.groupBy = append(qb.groupBy, "service_name")
	qb.orderBy = append(qb.orderBy, orderTotalCostDesc)
	qb.limit = &limit

	if filters != nil {
		qb = applyFilters(qb, filters)
	}

	return qb, nil
}

// BuildCostTrendByTime creates a query for cost trends over time
func BuildCostTrendByTime(granularity database.TimeGranularity, filters *database.AnalyticsFilters) (*FOCUSQueryBuilder, error) {
	qb := NewFOCUSQueryBuilder()
	qb.fromTable = focusTableName //nolint:goconst // FOCUS table name
	timeGroup := timeGroupExpr(granularity)

	qb.selectCols = append(qb.selectCols, fmt.Sprintf("%s as time_period", timeGroup))
	qb.selectCols = append(qb.selectCols, selTotalCost)
	qb.groupBy = append(qb.groupBy, timeGroup)
	qb.orderBy = append(qb.orderBy, orderTimePeriodAsc)

	if filters != nil {
		qb = applyFilters(qb, filters)
	}

	return qb, nil
}

// Deprecated prebuilt helpers (BuildCostByProvider/Region) removed as unused. Use the builder directly when needed.

// applyFilters applies analytics filters to query builder
func applyFilters(qb *FOCUSQueryBuilder, filters *database.AnalyticsFilters) *FOCUSQueryBuilder {
	if filters.StartDate != nil && filters.EndDate != nil {
		qb.FilterByDateRange(*filters.StartDate, *filters.EndDate)
	}

	if len(filters.Providers) > 0 {
		for _, provider := range filters.Providers {
			qb.FilterByProvider(provider)
		}
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

	if filters.MinCost != nil {
		qb.FilterByCostThreshold(*filters.MinCost)
	}

	if filters.MaxCost != nil {
		qb.Where("effective_cost <= %f", *filters.MaxCost)
	}

	// Tenant scoping (only if filter contains tenant id; higher layers ensure flag gating)
	if filters.TenantID != "" {
		qb.FilterByTenant(filters.TenantID)
	}

	return qb
}
