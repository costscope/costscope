//go:build qb_extended

package focus

import "local/costscope/internal/core/monitoring/telemetry"

// Extended build-only query forms for COUNT / EXPLAIN.

// BuildCount generates a COUNT(*) variant of the current query.
func (qb *FOCUSQueryBuilder) BuildCount() (string, []interface{}, error) {
	// Telemetry: extended variant build
	telemetry.QueryBuilderBuilds.WithLabelValues("extended").Inc()
	countBuilder := &FOCUSQueryBuilder{
		fromTable:  qb.fromTable,
		joins:      qb.joins,
		conditions: qb.conditions,
		groupBy:    qb.groupBy,
		having:     qb.having,
		ctes:       qb.ctes,
	}
	countBuilder.selectCols = []string{"COUNT(*) as total_count"}
	return countBuilder.Build()
}

// BuildExplain wraps the built SQL with EXPLAIN.
func (qb *FOCUSQueryBuilder) BuildExplain() (string, []interface{}, error) {
	telemetry.QueryBuilderBuilds.WithLabelValues("extended").Inc()
	query, params, err := qb.Build()
	if err != nil {
		return "", nil, err
	}
	return "EXPLAIN " + query, params, nil
}
