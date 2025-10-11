package aws

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTempCSV writes content to a temp file.
func writeTempCSV(t *testing.T, dir, name, content string) string {
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp csv: %v", err)
	}
	return p
}

func TestAWSCSVRowSourceEOF(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeTempCSV(t, dir, "sample.csv", "a,b\n1,2\n3,4\n")
	src, headers, err := NewCSVRowSource(p)
	if err != nil {
		t.Fatalf("NewCSVRowSource: %v", err)
	}
	if len(headers) != 2 {
		t.Fatalf("expected 2 headers got %d", len(headers))
	}
	defer func() { _ = src.Close() }()
	ctx := context.Background()
	rows := 0
	for {
		row, err := src.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("row read: %v", err)
		}
		if len(row) != 2 {
			t.Fatalf("expected 2 cols got %d", len(row))
		}
		rows++
	}
	if rows != 2 {
		t.Fatalf("expected 2 data rows got %d", rows)
	}
}

func TestAWSCSVRowSourceContextCancel(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeTempCSV(t, dir, "cancel.csv", "a,b\n1,2\n")
	src, _, err := NewCSVRowSource(p)
	if err != nil {
		t.Fatalf("NewCSVRowSource: %v", err)
	}
	defer func() { _ = src.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, rerr := src.Next(ctx)
	if rerr == nil {
		time.Sleep(5 * time.Millisecond)
		_, rerr = src.Next(ctx)
	}
	if rerr == nil || rerr == io.EOF {
		t.Fatalf("expected cancellation error got %v", rerr)
	}
}

func TestAWSCSVRowSourceVariableColumns(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeTempCSV(t, dir, "variable.csv", "a,b\n1,2,3\n4,5\n")
	src, headers, err := NewCSVRowSource(p)
	if err != nil {
		t.Fatalf("NewCSVRowSource: %v", err)
	}
	if len(headers) != 2 {
		t.Fatalf("expected 2 headers got %d", len(headers))
	}
	defer func() { _ = src.Close() }()
	ctx := context.Background()
	row1, err := src.Next(ctx)
	if err != nil {
		t.Fatalf("row1: %v", err)
	}
	if len(row1) != 3 {
		t.Fatalf("expected 3 cols row1 got %d", len(row1))
	}
	row2, err := src.Next(ctx)
	if err != nil {
		t.Fatalf("row2: %v", err)
	}
	if len(row2) != 2 {
		t.Fatalf("expected 2 cols row2 got %d", len(row2))
	}
	if _, err := src.Next(ctx); err != io.EOF {
		t.Fatalf("expected EOF got %v", err)
	}
	if err := src.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := src.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
