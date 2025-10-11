package gcp

import (
	"bytes"
	"testing"
)

func TestCSVReader_ReadHeadersAndChunk(t *testing.T) {
	t.Parallel()
	data := "h1,h2\na,b\nc,d\n"
	r := NewCSVReader(bytes.NewBufferString(data))
	h, err := r.ReadHeaders()
	if err != nil {
		t.Fatalf("headers: %v", err)
	}
	if len(h) != 2 || h[0] != "h1" || h[1] != "h2" {
		t.Fatalf("bad headers: %v", h)
	}
	chunk, read, err := r.ReadChunk(1)
	if err != nil {
		t.Fatalf("chunk1: %v", err)
	}
	if read != 1 || len(chunk) != 1 || len(chunk[0]) != 2 || chunk[0][0] != "a" || chunk[0][1] != "b" {
		t.Fatalf("bad chunk1: %v read=%d", chunk, read)
	}
	chunk, read, err = r.ReadChunk(10)
	if err != nil {
		t.Fatalf("chunk2: %v", err)
	}
	if read != 1 || len(chunk) != 1 || chunk[0][0] != "c" || chunk[0][1] != "d" {
		t.Fatalf("bad chunk2: %v read=%d", chunk, read)
	}
}
