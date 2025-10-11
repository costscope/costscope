package api

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// (Duplicate of tests in other files removed during consolidation)

// If Casbin is enabled and model/policy are provided but enforcer init fails (stub), function should not panic and handler remains usable
func TestWrapServerWithCasbinIfEnabled_WithPaths_ErrorNoPanic(t *testing.T) {
	prevEnabled := enterpriseCasbinEnabled
	prevModel := enterpriseCasbinModelPath
	prevPolicy := enterpriseCasbinPolicyPath
	defer func() {
		enterpriseCasbinEnabled = prevEnabled
		enterpriseCasbinModelPath = prevModel
		enterpriseCasbinPolicyPath = prevPolicy
	}()
	enterpriseCasbinEnabled = true
	enterpriseCasbinModelPath = "some-model.conf"
	enterpriseCasbinPolicyPath = "some-policy.csv"

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	var handler http.Handler = h
	logger := logging.NewLogger(logging.LevelError)
	wrapServerWithCasbinIfEnabled(&handler, logger)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected original handler to return 200, got %d", rr.Code)
	}
}

// buildTLSConfig should use provided TLS 1.2 cipher names when recognized
func TestBuildTLSConfig_AppliesProvidedCipherNames(t *testing.T) {
	cfg := buildTLSConfig("1.2", []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}, true)
	if cfg == nil {
		t.Fatal("expected non-nil tls.Config")
	}
	want := tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	found := false
	for _, c := range cfg.CipherSuites {
		if c == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cipher suite %v in cfg.CipherSuites", want)
	}
}

// Ensure reports routes are present under /api/v1/reports
func TestBuildModuleRouteGroups_ReportsRoutesPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelError)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	e := gin.New()
	v1 := e.Group("/api/v1")
	RegisterGinRouteGroups(v1, buildModuleRouteGroups(reg, logger, nil))

	want := map[string]struct{}{
		"/api/v1/reports/exports":      {},
		"/api/v1/reports/:id/download": {},
	}
	found := map[string]bool{}
	for _, rt := range e.Routes() {
		if _, ok := want[rt.Path]; ok && rt.Method == http.MethodGet {
			found[rt.Path] = true
		}
	}
	for p := range want {
		if !found[p] {
			t.Fatalf("expected route %s to be registered", p)
		}
	}
}

// Ensure streaming job routes exist
func TestBuildModuleRouteGroups_StreamingRoutesPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelError)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	e := gin.New()
	v1 := e.Group("/api/v1")
	RegisterGinRouteGroups(v1, buildModuleRouteGroups(reg, logger, nil))

	want := []string{"/api/v1/streaming/jobs", "/api/v1/streaming/jobs/:id/start"}
	have := map[string]bool{}
	for _, rt := range e.Routes() {
		for _, p := range want {
			if rt.Path == p {
				have[p] = true
			}
		}
	}
	for _, p := range want {
		if !have[p] {
			t.Fatalf("missing streaming route: %s", p)
		}
	}
}

// Ensure monitoring routes are present
func TestBuildModuleRouteGroups_MonitoringRoutesPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelError)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	e := gin.New()
	v1 := e.Group("/api/v1")
	RegisterGinRouteGroups(v1, buildModuleRouteGroups(reg, logger, nil))

	want := []string{"/api/v1/monitoring/metrics", "/api/v1/monitoring/alerts"}
	have := map[string]bool{}
	for _, rt := range e.Routes() {
		for _, p := range want {
			if rt.Path == p {
				have[p] = true
			}
		}
	}
	for _, p := range want {
		if !have[p] {
			t.Fatalf("missing monitoring route: %s", p)
		}
	}
}

// Ensure providers services route exists (sanity check for providers group)
// (Providers/services route test already exists in other file; omitted here)
