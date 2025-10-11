package azure

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// multiReadCloser is a tiny helper local to the azure package to avoid import cycles.
type multiReadCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiReadCloser) Close() error {
	var firstErr error
	for i := len(m.closers) - 1; i >= 0; i-- {
		if err := m.closers[i].Close(); firstErr == nil && err != nil {
			firstErr = err
		}
	}
	return firstErr
}

// OpenInput opens an Azure input path (CSV or JSON, with optional .gz) and returns a ReadCloser
// plus the detected inner extension (".csv" or ".json"). Behavior mirrors the legacy root helper.
func OpenInput(inputPath string) (io.ReadCloser, string, error) {
	// #nosec G304 - validated by ValidateInput before call
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, "", err
	}
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext == ".gz" { // keep parity with legacy awsExtGZ constant
		base := strings.TrimSuffix(inputPath, ".gz")
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
