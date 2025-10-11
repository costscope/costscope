package azure

import (
	"bytes"
	"context"
	"testing"
)

func TestCSVRowSource_HeadersNextClose(t *testing.T) {
	t.Parallel()
	data := "h1,h2\na,b\nc,d\n"
	src, headers, err := NewCSVRowSourceFromReader(bytes.NewBufferString(data))
	if err != nil {
		t.Fatalf("new csv source: %v", err)
	}
	t.Cleanup(func() { _ = src.Close() })
	if len(headers) != 2 || headers[0] != "h1" || headers[1] != "h2" {
		t.Fatalf("bad headers: %v", headers)
	}
	row, err := src.Next(context.Background())
	if err != nil || len(row) != 2 || row[0] != "a" || row[1] != "b" {
		t.Fatalf("row1: %v err=%v", row, err)
	}
	row, err = src.Next(context.Background())
	if err != nil || row[0] != "c" || row[1] != "d" {
		t.Fatalf("row2: %v err=%v", row, err)
	}
}

func TestNewJSONStreamFromReader_TopLevelDetection(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		typeName string
	}{
		{"array", "[{}]", "array"},
		{"ndjson", "{\"a\":1}\n{\"b\":2}\n", "ndjson"},
		{"whitespace_ndjson", "  \n{\"a\":1}\n", "ndjson"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			s, err := NewJSONStreamFromReader(bytes.NewBufferString(c.input))
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			switch st := s.(type) {
			case *jsonArrayStream:
				if c.typeName != "array" {
					t.Fatalf("expected array, got %T", st)
				}
			case *ndjsonStream:
				if c.typeName != "ndjson" {
					t.Fatalf("expected ndjson, got %T", st)
				}
			default:
				t.Fatalf("unexpected type: %T", st)
			}
		})
	}
}

func TestNDJSONStream_NextObject(t *testing.T) {
	t.Parallel()
	data := "{\"x\":1}\n{\"y\":2}\n\n{\"z\":3}\n"
	s, err := NewJSONStreamFromReader(bytes.NewBufferString(data))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	nd, ok := s.(*ndjsonStream)
	if !ok {
		t.Fatalf("expected ndjsonStream, got %T", s)
	}
	obj, err := nd.NextObject(context.Background())
	if err != nil || obj["x"].(float64) != 1 {
		t.Fatalf("obj1: %v err=%v", obj, err)
	}
	obj, err = nd.NextObject(context.Background())
	if err != nil || obj["y"].(float64) != 2 {
		t.Fatalf("obj2: %v err=%v", obj, err)
	}
	obj, err = nd.NextObject(context.Background())
	if err != nil || obj["z"].(float64) != 3 {
		t.Fatalf("obj3: %v err=%v", obj, err)
	}
}
