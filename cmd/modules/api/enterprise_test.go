package api

import (
	"strings"
	"testing"

	"local/costscope/internal/core/config"
	"local/costscope/internal/core/logging"
)

func TestValidateEnterpriseJWTSecret(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{"empty", "", true},
		{"short-16", strings.Repeat("a", 16), true},
		{"short-31", strings.Repeat("a", 31), true},
		{"ok-32", strings.Repeat("a", 32), false},
		{"long-48", strings.Repeat("a", 48), false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateEnterpriseJWTSecret(tc.secret)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestContainsWildcardOrigin(t *testing.T) {
	t.Parallel()
	if !containsWildcardOrigin([]string{"*"}) {
		t.Fatalf("expected true for ['*']")
	}
	if containsWildcardOrigin([]string{"https://example.com"}) {
		t.Fatalf("expected false for specific origin")
	}
	if !containsWildcardOrigin([]string{"https://example.com", "*"}) {
		t.Fatalf("expected true when list contains '*'")
	}
}

func TestCheckAndLogCorsWarning(t *testing.T) {
	t.Parallel()
	logger := logging.NewLogger(logging.LevelDebug)

	// Production with wildcard should warn (cannot capture output here; smoke test only)
	cfg := &config.ConsolidatedConfig{Environment: config.Production}
	checkAndLogCorsWarning(logger, cfg, []string{"*"})

	// Non-production should not warn even with wildcard
	cfg2 := &config.ConsolidatedConfig{Environment: config.Development}
	checkAndLogCorsWarning(logger, cfg2, []string{"*"})
}
