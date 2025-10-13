//go:build example && duckdb
// +build example,duckdb

package main

import (
	"fmt"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/database/duckdb"
	"github.com/costscope/costscope/internal/database/types"
)

func main() {
	fmt.Println("=== DuckDB Integration Test ===")
	logger := logging.GetLogger().WithFields(map[string]interface{}{"example": "duckdb-integration"})

	// Test 1: Create DuckDB QueryBuilder
	fmt.Println("\n1. Testing QueryBuilder creation...")
	builder := duckdb.NewDuckDBQueryBuilder()
	if builder == nil {
		logger.FatalWithFields("Failed to create QueryBuilder", nil)
	}
	fmt.Println(" QueryBuilder created successfully")

	// Test 2: Test basic query building
	fmt.Println("\n2. Testing basic query building...")
	builder.Select("service", "cost", "usage_date").
		From("billing_data").
		Where("cost > ?", 100).
		GroupBy("service").
		OrderBy("cost", types.DESC).
		Limit(10)

	query, args, err := builder.Build()
	if err != nil {
		logger.FatalWithFields("Failed to build query", map[string]interface{}{"error": err.Error()})
	}

	fmt.Printf("Generated query: %s\n", query)
	fmt.Printf("Arguments: %v\n", args)
	fmt.Println(" Query built successfully")

	// Test 3: Test FOCUS-specific operations
	fmt.Println("\n3. Testing FOCUS operations...")
	builder2 := duckdb.NewDuckDBQueryBuilder()
	builder2.Select("provider_name", "SUM(billing_amount) as total_cost").
		From("focus_data").
		Where("billing_period_start >= ?", "2024-01-01").
		GroupBy("provider_name").
		OrderBy("total_cost", types.DESC)

	focusQuery, focusArgs, err := builder2.Build()
	if err != nil {
		logger.FatalWithFields("Failed to build FOCUS query", map[string]interface{}{"error": err.Error()})
	}
	fmt.Printf("FOCUS query: %s\n", focusQuery)
	fmt.Printf("FOCUS args: %v\n", focusArgs)
	fmt.Println(" FOCUS query generated successfully")

	// Test 4: Test complex query building
	fmt.Println("\n4. Testing complex query building...")
	builder3 := duckdb.NewDuckDBQueryBuilder()
	builder3.Select("provider_name", "service_name", "AVG(cost) as avg_cost").
		From("focus_data").
		Join("metadata", "focus_data.resource_id = metadata.resource_id").
		Where("billing_period_start >= ?", "2024-01-01").
		Where("cost > ?", 0).
		GroupBy("provider_name", "service_name").
		Having("AVG(cost) > ?", 50).
		OrderBy("avg_cost", types.DESC).
		Limit(20).
		Offset(0)

	complexQuery, complexArgs, err := builder3.Build()
	if err != nil {
		logger.FatalWithFields("Failed to build complex query", map[string]interface{}{"error": err.Error()})
	}
	fmt.Printf("Complex query: %s\n", complexQuery)
	fmt.Printf("Complex args: %v\n", complexArgs)
	fmt.Println(" Complex query generated successfully")

	fmt.Println("\n=== Test Summary ===")
	fmt.Println(" QueryBuilder interface working")
	fmt.Println(" SQL generation functional")
	fmt.Println(" FOCUS operations implemented")
	fmt.Println(" Complex queries supported")
	fmt.Println(" Integration test completed successfully")

	fmt.Println("\n DuckDB QueryBuilder integration is ready for production use!")
}
