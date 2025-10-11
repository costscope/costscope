package paths

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeOutput(t *testing.T) {
	baseAbs, _ := filepath.Abs(BaseOutputDir)

	tests := []struct {
		name       string
		in         string
		wantErr    bool
		wantInside bool
	}{
		{name: "valid relative filename", in: "focus.parquet", wantErr: false, wantInside: true},
		{name: "absolute inside", in: filepath.Join(baseAbs, "nested", "ok.parquet"), wantErr: false, wantInside: true},
		{name: "traversal attempt", in: "../evil.parquet", wantErr: true, wantInside: false},
		{name: "absolute root outside", in: filepath.Clean("/tmp/outside.parquet"), wantErr: true, wantInside: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := SanitizeOutput(tc.in)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got none (out=%s)", out)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr {
				if !strings.Contains(err.Error(), ErrOutsideDataDir.Error()) {
					t.Fatalf("error does not contain sentinel: %v", err)
				}
				return
			}
			// Success path should be inside base
			if !strings.HasPrefix(out, baseAbs) {
				t.Fatalf("expected output inside %s, got %s", baseAbs, out)
			}
		})
	}
}
