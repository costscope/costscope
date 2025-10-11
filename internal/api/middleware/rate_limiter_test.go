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

// Test the custom RateLimiter (in auth.go) middleware behavior
func TestRateLimiter_Middleware_BasicLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Use a deterministic key so we don't rely on ClientIP in tests
	rl := NewRateLimiter(nil, 2, time.Hour, func(c *gin.Context) string { return "test-client" })
	r.Use(rl.Middleware())
	r.GET("/p", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// 1st request: allowed
	req1, _ := http.NewRequest("GET", "/p", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("1st status = %d, want 200", w1.Code)
	}
	if w1.Header().Get("X-RateLimit-Limit") != "2" {
		t.Fatalf("limit header = %q, want 2", w1.Header().Get("X-RateLimit-Limit"))
	}

	// 2nd request: allowed, remaining should be 0 afterward
	req2, _ := http.NewRequest("GET", "/p", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("2nd status = %d, want 200", w2.Code)
	}
	if w2.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("remaining header = %q, want 0", w2.Header().Get("X-RateLimit-Remaining"))
	}

	// 3rd request: blocked (429)
	req3, _ := http.NewRequest("GET", "/p", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("3rd status = %d, want 429", w3.Code)
	}
	if w3.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("remaining header after block = %q, want 0", w3.Header().Get("X-RateLimit-Remaining"))
	}
}
