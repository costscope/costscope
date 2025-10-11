package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinRoute represents a single Gin route registration entry
type GinRoute struct {
	Method     string
	Path       string
	Handler    gin.HandlerFunc
	Middleware []gin.HandlerFunc
}

// GinRouteGroup describes a group of routes under a common base path
type GinRouteGroup struct {
	BasePath   string
	Middleware []gin.HandlerFunc
	Routes     []GinRoute
}

// RegisterGinRouteGroups registers groups and routes on a parent group
func RegisterGinRouteGroups(parent *gin.RouterGroup, groups []GinRouteGroup) {
	for _, g := range groups {
		grp := parent.Group(g.BasePath)
		if len(g.Middleware) > 0 {
			grp.Use(g.Middleware...)
		}
		for _, r := range g.Routes {
			h := r.Handler
			if len(r.Middleware) > 0 {
				h = wrapGin(h, r.Middleware...)
			}
			switch r.Method {
			case http.MethodGet:
				grp.GET(r.Path, h)
			case http.MethodPost:
				grp.POST(r.Path, h)
			case http.MethodPut:
				grp.PUT(r.Path, h)
			case http.MethodPatch:
				grp.PATCH(r.Path, h)
			case http.MethodDelete:
				grp.DELETE(r.Path, h)
			case http.MethodHead:
				grp.HEAD(r.Path, h)
			case http.MethodOptions:
				grp.OPTIONS(r.Path, h)
			default:
				grp.Any(r.Path, h)
			}
		}
	}
}

func wrapGin(h gin.HandlerFunc, mws ...gin.HandlerFunc) gin.HandlerFunc {
	if len(mws) == 0 {
		return h
	}
	// Compose middleware manually into a single handler
	return func(c *gin.Context) {
		// Build a chain by iterating mws and finally calling h
		idx := 0
		var next gin.HandlerFunc
		next = func(cc *gin.Context) {
			if idx < len(mws) {
				mw := mws[idx]
				idx++
				mw(cc)
				if cc.IsAborted() {
					return
				}
				next(cc)
				return
			}
			h(cc)
		}
		next(c)
	}
}

// NetHTTP middleware chain for standard library mux
type HTTPMiddleware func(http.Handler) http.Handler

// ChainHTTP composes middlewares around a final handler
func ChainHTTP(h http.Handler, mws ...HTTPMiddleware) http.Handler {
	if len(mws) == 0 {
		return h
	}
	wrapped := h
	// Apply in reverse so the first provided executes first
	for i := len(mws) - 1; i >= 0; i-- {
		wrapped = mws[i](wrapped)
	}
	return wrapped
}

// requireMethod creates middleware enforcing a single HTTP method.
func requireMethod(method string) HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow CORS preflight to pass through to CORS middleware
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if r.Method != method {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
