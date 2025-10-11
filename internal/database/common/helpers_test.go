package common

import "testing"

func TestEqOrIn(t *testing.T) {
	cases := []struct {
		name   string
		input  []string
		expect string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"aws"}, "provider = 'aws'"},
		{"multi", []string{"a", "b"}, "provider IN ('a','b')"},
	}
	for _, c := range cases {
		got := EqOrIn("provider", c.input)
		if got != c.expect {
			t.Errorf("%s: expected %q got %q", c.name, c.expect, got)
		}
	}
}

// Removed tests for trivial helpers (order by / left join / cost threshold) after
// inlining into builders. EqOrIn retained as the only shared helper requiring
// dedicated branching tests.
