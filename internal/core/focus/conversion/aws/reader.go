package aws

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// RowSource abstracts sequential row access for AWS CUR data (CSV or CSV.GZ).
// Implementations MUST return io.EOF when exhausted. Returned row slice MUST NOT be
// reused (caller may retain). Implementations should honor context cancellation.
// Note: A legacy reader lived at internal/core/focus/reader/aws/csv.go (CURCSVReader)
// and was removed as deadcode (no production usages; superseded by this implementation).
// If future refactors reintroduce multiple reader implementations prefer adding
// them here or behind provider-agnostic abstractions under conversion/universal/.
type RowSource interface {
	Next(context.Context) ([]string, error)
	Close() error
}

// csvRowSource implements RowSource for local CSV / CSV.GZ files.
type csvRowSource struct {
	f        *os.File
	gz       *gzip.Reader
	csv      *csv.Reader
	closed   bool
	filePath string
}

// NewCSVRowSource opens a CUR CSV (optionally .gz) and returns the row source plus headers.
func NewCSVRowSource(path string) (RowSource, []string, error) { // #nosec G304 - validated earlier
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open input: %w", err)
	}
	var rdr io.Reader = f
	var gz *gzip.Reader
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".gz") {
		gr, gerr := gzip.NewReader(f)
		if gerr != nil {
			_ = f.Close()
			return nil, nil, fmt.Errorf("gzip reader: %w", gerr)
		}
		gz = gr
		rdr = gz
	}
	csvReader := csv.NewReader(rdr)
	// Allow variable number of fields per data row (CUR exports sometimes append extra columns).
	csvReader.FieldsPerRecord = -1
	headers, err := csvReader.Read()
	if err != nil {
		if gz != nil {
			_ = gz.Close()
		}
		_ = f.Close()
		return nil, nil, fmt.Errorf("read headers: %w", err)
	}
	return &csvRowSource{f: f, gz: gz, csv: csvReader, filePath: path}, headers, nil
}

func (s *csvRowSource) Next(ctx context.Context) ([]string, error) {
	if s.closed {
		return nil, io.EOF
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	row, err := s.csv.Read()
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (s *csvRowSource) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	if s.gz != nil {
		_ = s.gz.Close()
	}
	if s.f != nil {
		return s.f.Close()
	}
	return nil
}
