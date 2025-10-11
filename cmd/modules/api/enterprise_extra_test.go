package api

import (
	"os"
	"testing"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"
)

// 1: containsWildcardOrigin returns true when '*' present
func TestContainsWildcardOrigin_True(t *testing.T) {
	if !containsWildcardOrigin([]string{"https://a.example", "*"}) {
		t.Fatal("expected wildcard origin to be detected")
	}
}

// 2: containsWildcardOrigin returns false when no '*'
func TestContainsWildcardOrigin_False(t *testing.T) {
	if containsWildcardOrigin([]string{"https://a.example"}) {
		t.Fatal("did not expect wildcard origin")
	}
}

// 3: isProductionEnv reads COSTSCOPE_ENVIRONMENT
func TestIsProductionEnv_FromEnv(t *testing.T) {
	prev := os.Getenv("COSTSCOPE_ENVIRONMENT")
	defer func() {
		if err := os.Setenv("COSTSCOPE_ENVIRONMENT", prev); err != nil {
			t.Fatalf("defer Setenv failed: %v", err)
		}
	}()
	if err := os.Setenv("COSTSCOPE_ENVIRONMENT", "production"); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	if !isProductionEnv(nil) {
		t.Fatal("expected production env to be detected from COSTSCOPE_ENVIRONMENT")
	}
}

// 4: checkAndLogCorsWarning does not panic when called in production with wildcard origins
func TestCheckAndLogCorsWarning_DoesNotPanic(t *testing.T) {
	prevEnv := os.Getenv("COSTSCOPE_ENVIRONMENT")
	defer func() {
		if err := os.Setenv("COSTSCOPE_ENVIRONMENT", prevEnv); err != nil {
			t.Fatalf("defer Setenv failed: %v", err)
		}
	}()
	if err := os.Setenv("COSTSCOPE_ENVIRONMENT", "production"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	logger := logging.NewLogger(logging.LevelInfo)
	// Should not panic; just exercise the path that logs a warning
	checkAndLogCorsWarning(logger, nil, []string{"*"})
}

// 5: buildTLSConfig recognizes known cipher names and maps to numeric IDs
func TestBuildTLSConfig_RecognizedCipherName(t *testing.T) {
	name := "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	cfg := buildTLSConfig("1.2", []string{name}, false)
	if cfg == nil {
		t.Fatal("expected non-nil tls.Config")
	}
	if len(cfg.CipherSuites) == 0 {
		t.Fatal("expected cipher suites to be populated when a known name is provided")
	}
	// ensure the mapped ID is present
	want, ok := tlsCipherNameToID[name]
	if !ok {
		t.Fatalf("test misconfigured; tlsCipherNameToID missing %s", name)
	}
	found := false
	for _, id := range cfg.CipherSuites {
		if id == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cipher suite id %v to be present in cfg.CipherSuites", want)
	}
}

// 6: buildModuleRouteGroups contains analytics /models route
func TestBuildModuleRouteGroups_HasAnalyticsModelsRoute(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	groups := buildModuleRouteGroups(reg, logger, nil)
	got := false
	for _, g := range groups {
		if g.BasePath == "/analytics" {
			for _, r := range g.Routes {
				if r.Path == "/models" {
					got = true
					break
				}
			}
		}
	}
	if !got {
		t.Fatal("expected /analytics base to contain /models route")
	}
}
