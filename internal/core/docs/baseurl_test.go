package docs

import (
	"os"
	"testing"
)

func TestGetWSBaseURLVariants(t *testing.T) {
	cases := []struct {
		env  string
		base string
		ws   string
	}{
		{"http://example.com:9000", "http://example.com:9000", "ws://example.com:9000"},
		{"https://api.company.io", "https://api.company.io", "wss://api.company.io"},
		{"dev.internal.local", "dev.internal.local", "ws://dev.internal.local"}, // bare host kept as-is for base, ws assumes ws://
		{"", "http://localhost:8080", "ws://localhost:8080"},                    // default
	}
	for _, c := range cases {
		if err := os.Setenv("DOCS_BASE_URL", c.env); err != nil {
			t.Fatalf("failed to set env: %v", err)
		}
		b := GetBaseURL()
		if b != c.base {
			t.Fatalf("base mismatch for %q: got %s want %s", c.env, b, c.base)
		}
		w := GetWSBaseURL()
		if w != c.ws {
			t.Fatalf("ws mismatch for %q: got %s want %s", c.env, w, c.ws)
		}
	}
}
