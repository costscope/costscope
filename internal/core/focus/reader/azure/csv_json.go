package azure

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type CSVRowSource interface {
	Next(context.Context) ([]string, error)
	Close() error
}

type csvRowSource struct{ csv *csv.Reader }

func NewCSVRowSourceFromReader(r io.Reader) (CSVRowSource, []string, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.ReuseRecord = false
	cr.FieldsPerRecord = -1
	h, err := cr.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	return &csvRowSource{csv: cr}, h, nil
}

func (s *csvRowSource) Next(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	row, err := s.csv.Read()
	if err != nil {
		return nil, err
	}
	cp := make([]string, len(row))
	copy(cp, row)
	return cp, nil
}

func (s *csvRowSource) Close() error { return nil }

type JSONStream interface {
	NextObject(context.Context) (map[string]interface{}, error)
}

type jsonArrayStream struct {
	dec     *json.Decoder
	started bool
	ended   bool
}

type ndjsonStream struct{ br *bufio.Reader }

func NewJSONStreamFromReader(r io.Reader) (JSONStream, error) {
	br := bufio.NewReader(r)
	for {
		b, err := br.Peek(1)
		if err != nil {
			if err == io.EOF {
				return &ndjsonStream{br: br}, nil
			}
			return nil, fmt.Errorf("failed to read JSON: %w", err)
		}
		if len(b) == 0 {
			return &ndjsonStream{br: br}, nil
		}
		switch b[0] {
		case ' ', '\n', '\r', '\t':
			_, _ = br.ReadByte()
			continue
		case '[':
			return &jsonArrayStream{dec: json.NewDecoder(br)}, nil
		default:
			return &ndjsonStream{br: br}, nil
		}
	}
}

func (s *jsonArrayStream) NextObject(_ context.Context) (map[string]interface{}, error) {
	if s.ended {
		return nil, io.EOF
	}
	if !s.started {
		if tok, err := s.dec.Token(); err != nil {
			return nil, err
		} else if d, ok := tok.(json.Delim); !ok || d != '[' {
			return nil, fmt.Errorf("invalid JSON array")
		}
		s.started = true
	}
	if !s.dec.More() {
		_, _ = s.dec.Token()
		s.ended = true
		return nil, io.EOF
	}
	var obj map[string]interface{}
	if err := s.dec.Decode(&obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *ndjsonStream) NextObject(_ context.Context) (map[string]interface{}, error) {
	for {
		line, err := s.br.ReadBytes('\n')
		if len(strings.TrimSpace(string(line))) == 0 {
			if err != nil {
				return nil, err
			}
			continue
		}
		var obj map[string]interface{}
		if jerr := json.Unmarshal(line, &obj); jerr != nil {
			return nil, jerr
		}
		return obj, nil
	}
}
