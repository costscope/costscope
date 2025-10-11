package gcp

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	extGZ   = ".gz"
	extCSV  = ".csv"
	extJSON = ".json"
)

// OpenInput opens the input file (CSV or JSON), handles gzip transparently, and returns
// a ReadCloser along with the detected inner extension (".csv" or ".json").
func OpenInput(inputPath string) (io.ReadCloser, string, error) {
	// #nosec G304 - validated by ValidateInput before call
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, "", err
	}
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext == extGZ {
		base := strings.TrimSuffix(inputPath, extGZ)
		inner := strings.ToLower(filepath.Ext(base))
		gr, gerr := gzip.NewReader(f)
		if gerr != nil {
			_ = f.Close()
			return nil, "", gerr
		}
		return &multiReadCloser{Reader: gr, closers: []io.Closer{gr, f}}, inner, nil
	}
	return &multiReadCloser{Reader: f, closers: []io.Closer{f}}, ext, nil
}

// multiReadCloser mirrors the helper from the root package, locally scoped to avoid import cycles.
// It closes closer(s) in order after the reader is closed.

type multiReadCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiReadCloser) Close() error {
	var first error
	for i := len(m.closers) - 1; i >= 0; i-- {
		if err := m.closers[i].Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
