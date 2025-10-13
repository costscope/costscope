//go:build example && duckdb
// +build example,duckdb

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"database/sql"

	"github.com/costscope/costscope/internal/core/logging"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/spf13/cobra"
)

func main() {
	var cmd = &cobra.Command{
		Use:   "test-focus-duckdb [parquet-file]",
		Short: "Simple DuckDB FOCUS test",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			parquetFile := args[0]

			fmt.Printf(" Testing DuckDB with FOCUS file: %s\n", parquetFile)
			logger := logging.GetLogger().WithFields(map[string]interface{}{"example": "simple-duckdb"})

			// Check if file exists
			if _, err := os.Stat(parquetFile); os.IsNotExist(err) {
				logger.FatalWithFields("File does not exist", map[string]interface{}{"file": parquetFile})
			}

			// Open DuckDB connection
			db, err := sql.Open("duckdb", ":memory:")
			if err != nil {
				logger.FatalWithFields("Failed to open DuckDB", map[string]interface{}{"error": err.Error()})
			}
			defer func() { _ = db.Close() }()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			fmt.Println(" Loading Parquet file into DuckDB...")

			// Load parquet file using parameterized query for security
			query := "CREATE TABLE focus_data AS SELECT * FROM read_parquet(?)"
			if _, err := db.ExecContext(ctx, query, parquetFile); err != nil {
				logger.FatalWithFields("Failed to load parquet", map[string]interface{}{"error": err.Error()})
			}

			fmt.Println(" Parquet file loaded successfully!")

			// Get basic stats
			fmt.Println("\n Basic Statistics:")
			row := db.QueryRowContext(ctx, "SELECT COUNT(*), COUNT(DISTINCT service_name), MIN(effective_cost), MAX(effective_cost), SUM(effective_cost) FROM focus_data")

			var count, services int64
			var minCost, maxCost, totalCost float64
			if err := row.Scan(&count, &services, &minCost, &maxCost, &totalCost); err != nil {
				logger.WarnWithFields("Failed to get stats", map[string]interface{}{"error": err.Error()})
			} else {
				fmt.Printf("  Records: %d\n", count)
				fmt.Printf("  Services: %d\n", services)
				fmt.Printf("  Min Cost: $%.2f\n", minCost)
				fmt.Printf("  Max Cost: $%.2f\n", maxCost)
				fmt.Printf("  Total Cost: $%.2f\n", totalCost)
			}

			// Top 5 services by cost
			fmt.Println("\n Top 5 Services by Cost:")
			rows, err := db.QueryContext(ctx, "SELECT service_name, SUM(effective_cost) as total_cost FROM focus_data GROUP BY service_name ORDER BY total_cost DESC LIMIT 5")
			if err != nil {
				logger.WarnWithFields("Failed to get top services", map[string]interface{}{"error": err.Error()})
			} else {
				defer func() { _ = rows.Close() }()
				for rows.Next() {
					var service string
					var cost float64
					if err := rows.Scan(&service, &cost); err == nil {
						fmt.Printf("  %s: $%.2f\n", service, cost)
					}
				}
			}

			fmt.Println("\n DuckDB FOCUS test completed successfully!")
		},
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
