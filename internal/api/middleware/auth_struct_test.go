//go:build enterprise
// +build enterprise

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// helper to build a router for tests
func newGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware_RequireAuth_Table(t *testing.T) {
	// middleware.NewAuthMiddleware expects a *logging.Logger; nil is acceptable for these tests.
	am := NewAuthMiddleware(nil, "test-secret-very-long-for-jwt", "issuer")

	// generate a valid token
	validToken, err := am.GenerateToken("u1", "alice", "a@example", []string{"user"}, []string{"read"}, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	// generate an expired token
	expiredToken, err := am.GenerateToken("u1", "alice", "a@example", []string{"user"}, []string{"read"}, -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken(expired) error: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"missing header", "", http.StatusUnauthorized},
		{"bad format", "Token " + validToken, http.StatusUnauthorized},
		{"expired", "Bearer " + expiredToken, http.StatusUnauthorized},
		{"valid", "Bearer " + validToken, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newGin()
			r.Use(am.RequireAuth())
			r.GET("/p", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

			req, _ := http.NewRequest("GET", "/p", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthMiddleware_RequireRoleAndScope(t *testing.T) {
	am := NewAuthMiddleware(nil, "test-secret-very-long-for-jwt", "issuer")
	mkToken := func(roles, scopes []string) string {
		tok, err := am.GenerateToken("u1", "alice", "a@example", roles, scopes, time.Minute)
		if err != nil {
			t.Fatalf("GenerateToken error: %v", err)
		}
		return tok
	}

	tests := []struct {
		name       string
		roles      []string
		scopes     []string
		wantStatus int
	}{
		{"role denied", []string{"user"}, []string{"read:focus"}, http.StatusForbidden},
		{"scope denied", []string{"admin"}, []string{"other"}, http.StatusForbidden},
		{"admin role allows", []string{"admin"}, []string{"read:focus"}, http.StatusOK},
		// admin:all satisfies scope, but missing admin role still fails when RequireRole is chained
		{"scope admin:all without admin role denied", []string{"user"}, []string{"admin:all"}, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newGin()
			r.Use(am.RequireAuth())
			r.Use(am.RequireRole("admin"))
			r.Use(am.RequireScope("read:focus"))
			r.GET("/p", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

			token := mkToken(tt.roles, tt.scopes)
			req, _ := http.NewRequest("GET", "/p", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	am := NewAuthMiddleware(nil, "test-secret-very-long-for-jwt", "issuer")
	tok, err := am.GenerateToken("u1", "alice", "a@example", []string{"user"}, []string{"r"}, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	r := newGin()
	r.Use(am.OptionalAuth())
	r.GET("/p", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// no header
	req1, _ := http.NewRequest("GET", "/p", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("status: %d", w1.Code)
	}

	// with valid token
	req2, _ := http.NewRequest("GET", "/p", nil)
	req2.Header.Set("Authorization", "Bearer "+tok)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("status: %d", w2.Code)
	}
}

func TestCORS_PreflightAndHeaders(t *testing.T) {
	r := newGin()
	cfg := DefaultCORSConfig()
	r.Use(CORS(cfg))
	r.OPTIONS("/p", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.GET("/p", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// Preflight
	pre, _ := http.NewRequest("OPTIONS", "/p", nil)
	pre.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, pre)
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing CORS allow methods header")
	}

	// Actual request with origin
	req, _ := http.NewRequest("GET", "/p", nil)
	req.Header.Set("Origin", "https://example.com")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w2.Code)
	}
}

func TestCORS_SpecificOrigin(t *testing.T) {
	r := newGin()
	cfg := CORSConfig{AllowOrigins: []string{"https://allowed.example"}, AllowMethods: []string{"GET"}, AllowHeaders: []string{"Authorization"}}
	r.Use(CORS(cfg))
	r.GET("/p", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// Allowed origin
	req, _ := http.NewRequest("GET", "/p", nil)
	req.Header.Set("Origin", "https://allowed.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example" {
		t.Fatalf("allow-origin = %q, want https://allowed.example", got)
	}

	// Disallowed origin
	req2, _ := http.NewRequest("GET", "/p", nil)
	req2.Header.Set("Origin", "https://denied.example")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow-origin for denied origin: %q", got)
	}
}
