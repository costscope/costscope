package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRegisterGinRouteGroups_AllMethods ensures the registration switch covers all method
// branches (including default Any) and that middleware wrapping path is exercised.
func TestRegisterGinRouteGroups_AllMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	parent := router.Group("/g")

	// A simple handler that writes the method so we can assert the route was called.
	makeHandler := func(want string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(200, want)
		}
	}

	// A middleware that appends a header and continues.
	mw := func(c *gin.Context) {
		c.Writer.Header().Set("X-MW", "ok")
		c.Next()
	}

	groups := []GinRouteGroup{{
		BasePath: "/test",
		Routes: []GinRoute{
			{Method: http.MethodGet, Path: "/get", Handler: makeHandler("get")},
			{Method: http.MethodPost, Path: "/post", Handler: makeHandler("post")},
			{Method: http.MethodPut, Path: "/put", Handler: makeHandler("put")},
			{Method: http.MethodPatch, Path: "/patch", Handler: makeHandler("patch")},
			{Method: http.MethodDelete, Path: "/delete", Handler: makeHandler("delete")},
			{Method: http.MethodHead, Path: "/head", Handler: makeHandler("head")},
			{Method: http.MethodOptions, Path: "/options", Handler: makeHandler("options")},
			// Empty method triggers default (Any) branch; will register GET among others.
			{Method: "", Path: "/custom", Handler: makeHandler("custom")},
			// Route with middleware to ensure wrapGin path is used
			{Method: http.MethodGet, Path: "/withmw", Handler: makeHandler("withmw"), Middleware: []gin.HandlerFunc{mw}},
		},
	}}

	// Should register without panic
	RegisterGinRouteGroups(parent, groups)

	// Now make requests for each path and assert we get the expected body and header when applicable.
	testCases := []struct {
		method     string
		path       string
		want       string
		wantHeader bool
	}{
		{http.MethodGet, "/g/test/get", "get", false},
		{http.MethodPost, "/g/test/post", "post", false},
		{http.MethodPut, "/g/test/put", "put", false},
		{http.MethodPatch, "/g/test/patch", "patch", false},
		{http.MethodDelete, "/g/test/delete", "delete", false},
		{http.MethodHead, "/g/test/head", "", false},
		{http.MethodOptions, "/g/test/options", "options", false},
		{http.MethodGet, "/g/test/custom", "custom", false},
		{http.MethodGet, "/g/test/withmw", "withmw", true},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if tc.method == http.MethodHead {
			// HEAD responses may have empty body but should be 200
			if w.Code != 200 {
				t.Fatalf("HEAD %s returned code %d", tc.path, w.Code)
			}
			continue
		}
		if w.Code != 200 {
			t.Fatalf("%s %s returned code %d (body=%q)", tc.method, tc.path, w.Code, w.Body.String())
		}
		if tc.want != "" && !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("%s %s body mismatch: want %q got %q", tc.method, tc.path, tc.want, w.Body.String())
		}
		if tc.wantHeader {
			if w.Header().Get("X-MW") != "ok" {
				t.Fatalf("expected middleware header on %s %s", tc.method, tc.path)
			}
		}
	}
}
