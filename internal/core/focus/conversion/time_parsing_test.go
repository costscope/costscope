package conversion

import (
	"testing"
	"time"

	conv "local/costscope/internal/core/focus/conversion/common"
)

// TestParseTimeAnyFormats ensures multiple timestamp formats are parsed consistently.
func TestParseTimeAnyFormats(t *testing.T) {
	cases := []struct {
		in     string
		expect string
	}{
		{"2025-01-02T03:04:05Z", "2025-01-02T03:04:05Z"},      // RFC3339 Z
		{"2025-01-02T03:04:05+00:00", "2025-01-02T03:04:05Z"}, // RFC3339 offset
		{"2025-01-02 03:04:05", "2025-01-02T03:04:05Z"},       // space separated
		{"2025-01-02", "2025-01-02T00:00:00Z"},                // date only
		{"", ""},                                              // empty
		{"invalid", ""},                                       // invalid
	}
	for i, tc := range cases {
		got := conv.ParseTimeAny(tc.in)
		if tc.expect == "" {
			if !got.IsZero() {
				t.Fatalf("case %d expected zero time got %v", i, got)
			}
			continue
		}
		exp, _ := time.Parse(time.RFC3339, tc.expect)
		if !got.Equal(exp) {
			t.Fatalf("case %d mismatch: in=%s got=%s exp=%s", i, tc.in, got.UTC().Format(time.RFC3339), exp.Format(time.RFC3339))
		}
	}
}
