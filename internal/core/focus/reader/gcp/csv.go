package gcp

import (
	"encoding/csv"
	"io"
)

type CSVReader struct{ cr *csv.Reader }

func NewCSVReader(r io.Reader) *CSVReader {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.ReuseRecord = true
	return &CSVReader{cr: cr}
}

func (gr *CSVReader) ReadHeaders() ([]string, error) {
	h, err := gr.cr.Read()
	if err != nil {
		return nil, err
	}
	cp := make([]string, len(h))
	copy(cp, h)
	return cp, nil
}

func (gr *CSVReader) ReadChunk(chunkSize int) ([][]string, int64, error) {
	if chunkSize <= 0 {
		chunkSize = 10000
	}
	chunk := make([][]string, 0, chunkSize)
	var read int64
	for i := 0; i < chunkSize; i++ {
		rec, err := gr.cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return chunk, read, err
		}
		cp := make([]string, len(rec))
		copy(cp, rec)
		chunk = append(chunk, cp)
		read++
	}
	return chunk, read, nil
}

// (Removed deprecated JSONTopLevel helper; detection now in conversion/gcp/process_json.go)
