//go:build !duckdb

package commands

import "fmt"

func loadValues(path string) ([]float64, []float64, error) {
	return nil, nil, fmt.Errorf("drift dataset loading requires a binary built with -tags duckdb")
}
