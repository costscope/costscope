package api

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"
)

// 1: tlsCipherNameToID contains at least one known mapping
func TestTLSCipherNameMap_HasKnownKey(t *testing.T) {
	key := "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	id, ok := tlsCipherNameToID[key]
	if !ok {
		t.Fatalf("expected tlsCipherNameToID to contain key %s", key)
	}
	if id == 0 {
		t.Fatalf("unexpected zero id for cipher %s", key)
	}
}

// 2: registerHealthAndDocs when docs disabled only registers health endpoints
func TestRegisterHealthAndDocs_NoDocs(t *testing.T) {
	prev := docsEnabled
	docsEnabled = false
	defer func() { docsEnabled = prev }()

	r := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	// use a unique health path to avoid collisions
	prevHealth := healthPath
	healthPath = "/health-no-docs"
	defer func() { healthPath = prevHealth }()

	registerHealthAndDocs(r, reg)
	routes := r.Routes()
	haveHealth := false
	haveDocs := false
	for _, rt := range routes {
		if rt.Path == "/health-no-docs" {
			haveHealth = true
		}
		if rt.Path == docsPath || rt.Path == docsPath+"/*filepath" {
			haveDocs = true
		}
	}
	if !haveHealth {
		t.Fatal("expected health path to be registered")
	}
	if haveDocs {
		t.Fatal("did not expect docs to be registered when docsEnabled=false")
	}
}

// 3: buildModuleRouteGroups contains providers /:provider/services route
func TestBuildModuleRouteGroups_ProvidersServicesPresent(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	groups := buildModuleRouteGroups(reg, logger, nil)
	found := false
	for _, g := range groups {
		if g.BasePath == "/providers" {
			for _, r := range g.Routes {
				if r.Path == "/:provider/services" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Fatal("expected /providers to include /:provider/services route")
	}
}

// 4: buildTenantMiddleware with nil cfg (multitenant disabled) should call Next and not abort
func TestBuildTenantMiddleware_Disabled_Allows(t *testing.T) {
	mw := buildTenantMiddleware(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// create a dummy request context
	c.Request = httptest.NewRequest("GET", "/", nil)
	mw(c)
	if c.IsAborted() {
		t.Fatal("expected middleware not to abort when multitenant disabled")
	}
}

// 5: newEnterpriseGinRouter sets gin mode to release when GIN_MODE is empty
func TestNewEnterpriseGinRouter_SetsReleaseMode(t *testing.T) {
	prev := os.Getenv("GIN_MODE")
	if err := os.Unsetenv("GIN_MODE"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}
	defer func() {
		if err := os.Setenv("GIN_MODE", prev); err != nil {
			t.Fatalf("defer Setenv failed: %v", err)
		}
	}()

	// call function (it calls gin.SetMode internally)
	_ = newEnterpriseGinRouter()
	if gin.Mode() != gin.ReleaseMode {
		t.Fatalf("expected gin mode to be ReleaseMode, got %s", gin.Mode())
	}
}

// 6: buildTLSConfig default suites include at least one CHACHA20 entry in fallback
func TestBuildTLSConfig_DefaultIncludesChaCha(t *testing.T) {
	cfg := buildTLSConfig("1.2", nil, false)
	if cfg == nil {
		t.Fatal("expected non-nil tls.Config")
	}
	found := false
	for _, id := range cfg.CipherSuites {
		if id == tlsCipherNameToID["TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"] || id == tlsCipherNameToID["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"] {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one CHACHA20_POLY1305 cipher in defaults")
	}
}
