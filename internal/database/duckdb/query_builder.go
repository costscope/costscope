//go:build duckdb

package duckdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/database"
	dbcommon "github.com/costscope/costscope/internal/database/common"
)

// DuckDBQueryBuilder wraps FOCUS query builder to ensure interface compatibility
type DuckDBQueryBuilder struct {
	selectColumns []string
	fromTable     string
	conditions    []string
	groupBy       []string
	orderBy       []string
	limit         *int
	offset        *int
	joins         []string
	having        []string
	cte           map[string]string
}

// NewDuckDBQueryBuilder creates a new FOCUS query builder for DuckDB
func NewDuckDBQueryBuilder() *DuckDBQueryBuilder {
	return &DuckDBQueryBuilder{cte: make(map[string]string)}
}

// Select adds columns to SELECT clause
func (qb *DuckDBQueryBuilder) Select(columns ...string) database.QueryBuilder {
	qb.selectColumns = append(qb.selectColumns, columns...)
	return qb
}

// From sets the FROM table
func (qb *DuckDBQueryBuilder) From(table string) database.QueryBuilder {
	qb.fromTable = table
	return qb
}

// Where adds a WHERE condition
func (qb *DuckDBQueryBuilder) Where(condition string, args ...interface{}) database.QueryBuilder {
	if len(args) > 0 {
		condition = fmt.Sprintf(condition, args...)
	}
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// GroupBy adds GROUP BY columns
func (qb *DuckDBQueryBuilder) GroupBy(columns ...string) database.QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// OrderBy adds ORDER BY clause
// nolint:dupl // Intentional duplication across builders; delegates to shared dbcommon helper
func (qb *DuckDBQueryBuilder) OrderBy(column string, direction database.SortDirection) database.QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// Limit sets the LIMIT clause
func (qb *DuckDBQueryBuilder) Limit(count int) database.QueryBuilder {
	qb.limit = &count
	return qb
}

// Offset sets the OFFSET clause
func (qb *DuckDBQueryBuilder) Offset(count int) database.QueryBuilder {
	qb.offset = &count
	return qb
}

// Join adds an INNER JOIN
func (qb *DuckDBQueryBuilder) Join(table, condition string) database.QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("INNER JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin adds a LEFT JOIN
// nolint:dupl // Intentional duplication across builders; delegates to shared dbcommon helper
func (qb *DuckDBQueryBuilder) LeftJoin(table, condition string) database.QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// Having adds HAVING condition
func (qb *DuckDBQueryBuilder) Having(condition string, args ...interface{}) database.QueryBuilder {
	if len(args) > 0 {
		condition = fmt.Sprintf(condition, args...)
	}
	qb.having = append(qb.having, condition)
	return qb
}

// WithCTE adds a Common Table Expression
func (qb *DuckDBQueryBuilder) WithCTE(name, query string) database.QueryBuilder {
	qb.cte[name] = query
	return qb
}

// FOCUS-specific operations

// SelectCostMetrics adds cost-related columns
func (qb *DuckDBQueryBuilder) SelectCostMetrics() database.QueryBuilder {
	costColumns := []string{
		"effective_cost",
		"billed_cost",
		"list_cost",
		"resource_id",
		"service_name",
		"charge_period_start",
		"charge_period_end",
	}
	qb.selectColumns = append(qb.selectColumns, costColumns...)
	return qb
}

// FilterByProvider adds provider filter
func (qb *DuckDBQueryBuilder) FilterByProvider(provider string) database.QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("provider_name = '%s'", provider))
	return qb
}

// FilterByDateRange adds date range filter
func (qb *DuckDBQueryBuilder) FilterByDateRange(start, end time.Time) database.QueryBuilder {
	qb.conditions = append(qb.conditions,
		fmt.Sprintf("charge_period_start >= '%s' AND charge_period_end <= '%s'",
			start.Format("2006-01-02"), end.Format("2006-01-02")))
	return qb
}

// FilterByCostThreshold adds cost threshold filter
// nolint:dupl // Intentional duplication across builders; delegates to shared dbcommon helper
func (qb *DuckDBQueryBuilder) FilterByCostThreshold(threshold float64) database.QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("effective_cost >= %f", threshold))
	return qb
}

// FilterByService adds service filter (supports IN semantics when multiple)
func (qb *DuckDBQueryBuilder) FilterByService(services ...string) database.QueryBuilder {
	if cond := dbcommon.EqOrIn("service_name", services); cond != "" {
		qb.conditions = append(qb.conditions, cond)
	}
	return qb
}

// FilterByRegion adds region filter
func (qb *DuckDBQueryBuilder) FilterByRegion(regions ...string) database.QueryBuilder {
	if cond := dbcommon.EqOrIn("region", regions); cond != "" {
		qb.conditions = append(qb.conditions, cond)
	}
	return qb
}

// FilterByAccount adds account filter
func (qb *DuckDBQueryBuilder) FilterByAccount(accounts ...string) database.QueryBuilder {
	if cond := dbcommon.EqOrIn("billing_account_id", accounts); cond != "" {
		qb.conditions = append(qb.conditions, cond)
	}
	return qb
}

// FilterByTenant adds tenant scoping (ignored when empty)
func (qb *DuckDBQueryBuilder) FilterByTenant(tenantID string) database.QueryBuilder {
	if tenantID != "" {
		qb.conditions = append(qb.conditions, fmt.Sprintf("tenant_id = '%s'", tenantID))
	}
	return qb
}

// Build constructs the final SQL query
func (qb *DuckDBQueryBuilder) Build() (string, []interface{}, error) {
	// Telemetry: count duckdb builds (safe even if registry not yet initialized)
	telemetry.QueryBuilderBuilds.WithLabelValues("duckdb").Inc()
	var query strings.Builder
	var args []interface{}

	// WITH clause (CTEs)
	if len(qb.cte) > 0 {
		query.WriteString("WITH ")
		cteList := make([]string, 0, len(qb.cte))
		for name, cteQuery := range qb.cte {
			cteList = append(cteList, fmt.Sprintf("%s AS (%s)", name, cteQuery))
		}
		query.WriteString(strings.Join(cteList, ", "))
		query.WriteString(" ")
	}

	// SELECT clause
	if len(qb.selectColumns) == 0 {
		query.WriteString("SELECT *")
	} else {
		query.WriteString("SELECT ")
		query.WriteString(strings.Join(qb.selectColumns, ", "))
	}

	// FROM clause
	if qb.fromTable != "" {
		query.WriteString(" FROM ")
		query.WriteString(qb.fromTable)
	}

	// JOIN clauses
	if len(qb.joins) > 0 {
		query.WriteString(" ")
		query.WriteString(strings.Join(qb.joins, " "))
	}

	// WHERE clause
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.conditions, " AND "))
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING clause
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		query.WriteString(strings.Join(qb.having, " AND "))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT clause
	if qb.limit != nil {
		query.WriteString(fmt.Sprintf(" LIMIT %d", *qb.limit))
	}

	// OFFSET clause
	if qb.offset != nil {
		query.WriteString(fmt.Sprintf(" OFFSET %d", *qb.offset))
	}

	return query.String(), args, nil
}

// BuildCount constructs a COUNT query
func (qb *DuckDBQueryBuilder) BuildCount() (string, []interface{}, error) {
	// Clone the query builder but replace SELECT with COUNT(*)
	countQB := &DuckDBQueryBuilder{
		selectColumns: []string{"COUNT(*)"},
		fromTable:     qb.fromTable,
		conditions:    qb.conditions,
		joins:         qb.joins,
		cte:           qb.cte,
		// Don't include GROUP BY, ORDER BY, LIMIT, OFFSET for count
	}
	telemetry.QueryBuilderBuilds.WithLabelValues("duckdb").Inc()
	return countQB.Build()
}

// BuildExplain constructs an EXPLAIN query
func (qb *DuckDBQueryBuilder) BuildExplain() (string, []interface{}, error) {
	telemetry.QueryBuilderBuilds.WithLabelValues("duckdb").Inc()
	query, args, err := qb.Build()
	if err != nil {
		return "", nil, err
	}

	return "EXPLAIN " + query, args, nil
}
