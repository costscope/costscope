package api

import (
	"crypto/tls"
	"os"
	"testing"

	cfg "github.com/costscope/costscope/internal/core/config"
)

func TestIsProductionEnv_ConfigPriority(t *testing.T) {
	c := &cfg.ConsolidatedConfig{Environment: cfg.Production}
	if !isProductionEnv(c) {
		t.Fatalf("expected production when config.Environment=%v", cfg.Production)
	}
}

func TestIsProductionEnv_EnvVarsFallback(t *testing.T) {
	// Ensure env vars are considered when cfg is nil
	prev1 := os.Getenv("COSTSCOPE_ENVIRONMENT")
	prev2 := os.Getenv("ENV")
	defer func() {
		if err := os.Setenv("COSTSCOPE_ENVIRONMENT", prev1); err != nil {
			t.Fatalf("defer Setenv failed: %v", err)
		}
		if err := os.Setenv("ENV", prev2); err != nil {
			t.Fatalf("defer Setenv failed: %v", err)
		}
	}()

	if err := os.Unsetenv("COSTSCOPE_ENVIRONMENT"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}
	if err := os.Unsetenv("ENV"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}
	if isProductionEnv(nil) {
		t.Fatalf("expected false when no envs set and cfg nil")
	}

	if err := os.Setenv("COSTSCOPE_ENVIRONMENT", "production"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if !isProductionEnv(nil) {
		t.Fatalf("expected true when COSTSCOPE_ENVIRONMENT=production")
	}
	if err := os.Unsetenv("COSTSCOPE_ENVIRONMENT"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}
	if err := os.Setenv("ENV", "production"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if !isProductionEnv(nil) {
		t.Fatalf("expected true when ENV=production")
	}
}

// containsWildcardOrigin is covered by existing tests in enterprise_test.go

func TestBuildTLSConfig_MinVersionAndCipherFallback(t *testing.T) {
	// TLS 1.3 path
	c := buildTLSConfig("1.3", []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}, true)
	if c.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected TLS1.3 min version, got %v", c.MinVersion)
	}

	// TLS 1.2 with unknown cipher should fallback to defaults
	c2 := buildTLSConfig("1.2", []string{"UNKNOWN_CIPHER"}, false)
	if len(c2.CipherSuites) == 0 {
		t.Fatalf("expected default cipher suites when unknown provided")
	}
}

func TestBuildEnterpriseAPICommand_Basics(t *testing.T) {
	cmd := BuildEnterpriseAPICommand()
	if cmd == nil {
		t.Fatalf("expected non-nil command")
	}
	if cmd.Use != "enterprise" {
		t.Fatalf("unexpected use: %s", cmd.Use)
	}
	// check a few flags exist
	if cmd.Flags().Lookup("host") == nil {
		t.Fatalf("expected host flag")
	}
	if cmd.Flags().Lookup("port") == nil {
		t.Fatalf("expected port flag")
	}
	if cmd.Flags().Lookup("jwt-secret") == nil {
		t.Fatalf("expected jwt-secret flag")
	}
}
