package azure

import (
	"io"
	raz "local/costscope/internal/core/focus/reader/azure"
)

// CSVRowSource and JSONStream are forwarded types from reader/azure to provide
// a stable surface within the conversion/azure subpackage during migration.
type CSVRowSource = raz.CSVRowSource
type JSONStream = raz.JSONStream

// NewCSVRowSourceFromReader wraps reader/azure constructor to keep imports stable.
func NewCSVRowSourceFromReader(r io.Reader) (CSVRowSource, []string, error) {
	return raz.NewCSVRowSourceFromReader(r)
}

// NewJSONStreamFromReader wraps reader/azure constructor to keep imports stable.
func NewJSONStreamFromReader(r io.Reader) (JSONStream, error) {
	return raz.NewJSONStreamFromReader(r)
}
