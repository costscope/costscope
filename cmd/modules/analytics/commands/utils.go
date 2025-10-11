package commands

import (
	"os"
	"path/filepath"
	"strings"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

const (
	formatCSV     = "csv"
	formatJSON    = "json"
	formatParquet = "parquet"
	formatExcel   = "excel"
	formatUnknown = "unknown"
)

// detectFormat detects file format based on extension
func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".parquet":
		return formatParquet
	case ".csv":
		return formatCSV
	case ".json":
		return formatJSON
	case ".xlsx":
		return formatExcel
	default:
		return formatUnknown
	}
}
