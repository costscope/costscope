package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// 1: wrapGin composes middleware before handler and all run in order
func TestWrapGin_MiddlewareExecutesInOrder(t *testing.T) {
	handler := func(c *gin.Context) { c.Set("h", "ok") }
	mw1 := func(c *gin.Context) { c.Set("m1", "ok"); c.Next() }
	mw2 := func(c *gin.Context) { c.Set("m2", "ok"); c.Next() }

	wrapped := wrapGin(handler, mw1, mw2)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	wrapped(c)
	if _, ok := c.Get("m1"); !ok {
		t.Fatal("expected m1 to run")
	}
	if _, ok := c.Get("m2"); !ok {
		t.Fatal("expected m2 to run")
	}
	if _, ok := c.Get("h"); !ok {
		t.Fatal("expected handler to run")
	}
}

// 2: wrapGin respects Abort and does not run handler when middleware aborts
func TestWrapGin_AbortPreventsHandler(t *testing.T) {
	handler := func(c *gin.Context) { c.Set("h", "ok") }
	mw := func(c *gin.Context) { c.Abort() }
	wrapped := wrapGin(handler, mw)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	wrapped(c)
	if _, ok := c.Get("h"); ok {
		t.Fatal("expected handler NOT to run after Abort")
	}
}

// 3: ChainHTTP composes HTTP middlewares in provided order
func TestChainHTTP_Order(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("h")); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	})
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("1")); err != nil {
				t.Fatalf("write failed: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("2")); err != nil {
				t.Fatalf("write failed: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	}
	final := ChainHTTP(h, mw1, mw2)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	final.ServeHTTP(rr, req)
	body, _ := io.ReadAll(rr.Body)
	if string(body) != "12h" {
		t.Fatalf("unexpected body sequence: %s", string(body))
	}
}

// 4: requireMethod returns 405 for non-matching methods and allows OPTIONS
func TestRequireMethod_DisallowAndOptions(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	})
	mw := requireMethod(http.MethodPost)
	wrapped := mw(h)

	// GET should be 405
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", rr.Code)
	}

	// OPTIONS should be allowed
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodOptions, "/", nil)
	wrapped.ServeHTTP(rr2, req2)
	if rr2.Code != 200 {
		t.Fatalf("expected 200 for OPTIONS, got %d", rr2.Code)
	}
}

// 5: RegisterGinRouteGroups registers provided routes under the base path
func TestRegisterGinRouteGroups_Registers(t *testing.T) {
	r := gin.New()
	parent := r.Group("")
	groups := []GinRouteGroup{{
		BasePath: "/x",
		Routes:   []GinRoute{{Method: http.MethodGet, Path: "/ping", Handler: func(c *gin.Context) { c.String(200, "pong") }}},
	}}
	RegisterGinRouteGroups(parent, groups)
	found := false
	for _, rt := range r.Routes() {
		if rt.Method == http.MethodGet && rt.Path == "/x/ping" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected /x/ping route to be registered")
	}
}

// 6: requireMethod allows correct method through
func TestRequireMethod_AllowsMatching(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	wrapped := requireMethod(http.MethodPut)(h)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	wrapped.ServeHTTP(rr, req)
	if rr.Code != 204 {
		t.Fatalf("expected 204 for matching method, got %d", rr.Code)
	}
}

// Test that enabling Casbin but leaving model/policy empty logs a warning and does not panic
func TestWrapServerWithCasbinIfEnabled_NoModelPolicy_NoPanic(t *testing.T) {
	// set global flags and restore
	prev := enterpriseCasbinEnabled
	prevModel := enterpriseCasbinModelPath
	prevPolicy := enterpriseCasbinPolicyPath
	defer func() {
		enterpriseCasbinEnabled = prev
		enterpriseCasbinModelPath = prevModel
		enterpriseCasbinPolicyPath = prevPolicy
	}()
	enterpriseCasbinEnabled = true
	enterpriseCasbinModelPath = ""
	enterpriseCasbinPolicyPath = ""

	// Build a trivial handler
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	var handler http.Handler = h

	// Use a logger; ensure function returns without panic
	logger := logging.NewLogger(logging.LevelError)
	wrapServerWithCasbinIfEnabled(&handler, logger)

	// Handler should still be the original (stub will not wrap since model/policy empty)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 from original handler, got %d", rr.Code)
	}
}

// Small sanity: ensure tlsCipherNameToID contains at least one known key uppercase lookup
func TestTLSCipherNameToIDLookup(t *testing.T) {
	if _, ok := tlsCipherNameToID["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"]; !ok {
		t.Fatalf("expected tlsCipherNameToID to contain TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
	}
}
