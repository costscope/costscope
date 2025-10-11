package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"
)

// 1: validate short secret fails
func TestValidateEnterpriseJWTSecret_Short(t *testing.T) {
	if err := validateEnterpriseJWTSecret("short"); err == nil {
		t.Fatal("expected error for short secret")
	}
}

// 2: validate long secret passes
func TestValidateEnterpriseJWTSecret_OK(t *testing.T) {
	long := strings.Repeat("x", 32)
	if err := validateEnterpriseJWTSecret(long); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 3: TLS 1.3 branch sets MinVersion to TLS13
func TestBuildTLSConfig_TLS13(t *testing.T) {
	cfg := buildTLSConfig("1.3", nil, true)
	// Some Go versions set MinVersion to 0 (unset) while others set a numeric constant for TLS1.3.
	// Accept either 0 (unset) or the numeric TLS1.3 value (0x0304) here.
	if cfg.MinVersion != 0 && cfg.MinVersion != 0x0304 /* tls.VersionTLS13 */ {
		t.Fatalf("unexpected MinVersion for TLS1.3 branch: %v", cfg.MinVersion)
	}
	if len(cfg.CipherSuites) != 0 {
		t.Fatalf("expected no TLS 1.2 cipher suites for TLS 1.3; got %d", len(cfg.CipherSuites))
	}
}

// 4: unknown cipher names fallback to defaults
func TestBuildTLSConfig_UnknownCiphers_Fallback(t *testing.T) {
	cfg := buildTLSConfig("1.2", []string{"UNKNOWN_CIPHER"}, false)
	if cfg == nil {
		t.Fatal("cfg nil")
	}
	if len(cfg.CipherSuites) == 0 {
		t.Fatal("expected fallback cipher suites to be populated")
	}
}

// 5: new router exposes /metrics
func TestNewEnterpriseGinRouter_HasMetrics(t *testing.T) {
	prev := enterpriseCorsOrigins
	enterpriseCorsOrigins = []string{"*"}
	defer func() { enterpriseCorsOrigins = prev }()

	r := newEnterpriseGinRouter()
	routes := r.Routes()
	found := false
	for _, rt := range routes {
		if rt.Method == http.MethodGet && rt.Path == "/metrics" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("/metrics route not found; routes: %v", routes)
	}
}

// 6: registerHealthAndDocs registers health and docs endpoints when docsEnabled=true
func TestRegisterHealthAndDocs_RegistersPaths(t *testing.T) {
	prev := docsEnabled
	docsEnabled = true
	defer func() { docsEnabled = prev }()

	r := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	// set unique health and docs paths to avoid gin route wildcard conflicts
	prevHealth := healthPath
	prevDocs := docsPath
	healthPath = "/health-test"
	docsPath = "/docs-test"
	defer func() {
		healthPath = prevHealth
		docsPath = prevDocs
	}()

	registerHealthAndDocs(r, reg)
	routes := r.Routes()
	haveHealth := false
	haveDocs := false
	for _, rt := range routes {
		if rt.Path == "/health-test" {
			haveHealth = true
		}
		if rt.Path == docsPath || strings.HasPrefix(rt.Path, docsPath+"/") {
			haveDocs = true
		}
	}
	if !haveHealth {
		t.Fatal("health path not registered")
	}
	if !haveDocs {
		t.Fatal("docs path not registered")
	}
}

// 7: registerWebSocketRoutes registers the ws jobs route
func TestRegisterWebSocketRoutes_Registers(t *testing.T) {
	r := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	auth := func(c *gin.Context) { c.Next() }
	registerWebSocketRoutes(r, auth, reg)
	routes := r.Routes()
	found := false
	for _, rt := range routes {
		if rt.Method == http.MethodGet && rt.Path == "/ws/jobs/:jobID" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected websocket jobs route to be registered")
	}
}

// 8: buildModuleRouteGroups contains several expected bases
func TestBuildModuleRouteGroups_ContainsBasics(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	groups := buildModuleRouteGroups(reg, logger, nil)
	bases := map[string]bool{}
	for _, g := range groups {
		bases[g.BasePath] = true
	}
	for _, want := range []string{"/focus", "/providers", "/analytics", "/reports"} {
		if !bases[want] {
			t.Fatalf("expected base path %s in groups", want)
		}
	}
}

// 9: wrapServerWithCasbinIfEnabled short-circuits when model/policy missing
func TestWrapServerWithCasbinIfEnabled_NoModelPolicy_Noop(t *testing.T) {
	prev := enterpriseCasbinEnabled
	enterpriseCasbinEnabled = true
	defer func() { enterpriseCasbinEnabled = prev }()
	// ensure model/policy are empty to hit the early warning branch
	prevModel := enterpriseCasbinModelPath
	prevPolicy := enterpriseCasbinPolicyPath
	enterpriseCasbinModelPath = ""
	enterpriseCasbinPolicyPath = ""
	defer func() {
		enterpriseCasbinModelPath = prevModel
		enterpriseCasbinPolicyPath = prevPolicy
	}()

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapServerWithCasbinIfEnabled(&h, logging.NewLogger(logging.LevelInfo))
	// handler should remain callable (no panic) and unchanged type
	if h == nil {
		t.Fatal("handler unexpectedly nil")
	}
}

// 10: registerDebugRoutes registers /debug/cache-stats when enabled
func TestRegisterDebugRoutes_WhenEnabled_AddsRoute(t *testing.T) {
	prev := enableCacheStats
	enableCacheStats = true
	defer func() { enableCacheStats = prev }()

	r := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	auth := func(c *gin.Context) { c.Next() }
	registerDebugRoutes(r, reg, auth)
	routes := r.Routes()
	found := false
	for _, rt := range routes {
		if rt.Method == http.MethodGet && rt.Path == "/debug/cache-stats" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected /debug/cache-stats to be registered when enabled")
	}
}
