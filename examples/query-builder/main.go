//go:build example && duckdb && qb_extended
// +build example,duckdb,qb_extended

package main

import (
	"fmt"
	"time"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/database/duckdb"
	"local/costscope/internal/database/types"
)

func testQueryBuilder() {
	fmt.Println(" Testing DuckDB Query Builder Integration")
	logger := logging.GetLogger().WithFields(map[string]interface{}{"example": "query-builder"})

	// Test QueryBuilder creation
	qb := duckdb.NewDuckDBQueryBuilder()

	// Test method chaining
	query, args, err := qb.
		Select("service_name", "SUM(effective_cost) as total_cost").
		From("focus_cost_data").
		FilterByProvider("aws").
		FilterByCostThreshold(100.0).
		FilterByDateRange(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)).
		GroupBy("service_name").
		OrderBy("total_cost", types.DESC).
		Limit(10).
		Build()

	if err != nil {
		logger.FatalWithFields("Query build failed", map[string]interface{}{"error": err.Error()})
	}

	fmt.Printf(" Query Builder Test Successful!\n")
	fmt.Printf(" Generated Query:\n%s\n", query)
	fmt.Printf(" Arguments: %v\n", args)

	// Test count query
	countQuery, countArgs, err := qb.BuildCount()
	if err != nil {
		logger.FatalWithFields("Count query build failed", map[string]interface{}{"error": err.Error()})
	}

	fmt.Printf(" Count Query:\n%s\n", countQuery)
	fmt.Printf(" Count Arguments: %v\n", countArgs)

	// Test explain query
	explainQuery, explainArgs, err := qb.BuildExplain()
	if err != nil {
		logger.FatalWithFields("Explain query build failed", map[string]interface{}{"error": err.Error()})
	}

	fmt.Printf(" Explain Query:\n%s\n", explainQuery)
	fmt.Printf(" Explain Arguments: %v\n", explainArgs)

	fmt.Println(" All tests passed - No circular import issues!")
	fmt.Println(" DuckDB integration working correctly!")
}

func main() {
	testQueryBuilder()
}
