package api

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/costscope/costscope/internal/core/logging"
)

// RouteSpec declares an HTTP route exposed by the API server.
// FeatureGate is an optional env var name; when set to one of (0,false,off,disabled) the route is skipped.
// Tags allow lightweight grouping / future documentation export.
type RouteSpec struct {
	Method      string
	Path        string
	Handler     http.Handler
	Middleware  []HTTPMiddleware
	FeatureGate string   // optional env var controlling enablement
	Tags        []string // arbitrary labels (e.g. focus, analytics, system)
}

// BuildRouteSpecs returns the complete static list of API routes.
// Handlers are defined in handlers.go.
func BuildRouteSpecs(logger *logging.Logger) []RouteSpec { // logger kept for future dynamic handlers
	return []RouteSpec{
		// System
		{Method: http.MethodGet, Path: "/healthz", Handler: healthHandler(logger), Tags: []string{"system"}},
		{Method: http.MethodGet, Path: "/health", Handler: healthHandler(logger), Tags: []string{"system"}},
		{Method: http.MethodGet, Path: "/metrics", Handler: promhttp.Handler(), Tags: []string{"system", "metrics"}},
		{Method: http.MethodGet, Path: "/api/v1/info", Handler: infoHandler(logger), Tags: []string{"system"}},
		{Method: http.MethodGet, Path: "/api/v1/routes", Handler: routesSummaryHandler(logger), Tags: []string{"system", "docs"}},
		{Method: http.MethodGet, Path: "/docs", Handler: docsHandler(logger), Tags: []string{"system", "docs"}},
		// Legacy analytics aliases (unversioned) for client compatibility
		{Method: http.MethodGet, Path: "/costs/summary", Handler: analyticsSummaryHandler(logger), Tags: []string{"compat", "analytics"}},
		{Method: http.MethodGet, Path: "/breakdown", Handler: analyticsBreakdownHandler(logger), Tags: []string{"compat", "analytics"}},
		{Method: http.MethodGet, Path: "/costs/daily", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Default granularity to 'day' if not provided, then delegate to trends handler
			q := r.URL.Query()
			if strings.TrimSpace(q.Get("granularity")) == "" {
				q.Set("granularity", "day")
				r.URL.RawQuery = q.Encode()
			}
			analyticsTrendsHandler(logger).ServeHTTP(w, r)
		})},
		// WebSocket (prefix match)
		{Method: http.MethodGet, Path: "/ws/jobs/", Handler: wsJobsHandler(logger), Tags: []string{"ws", "realtime"}},

		// FOCUS
		{Method: http.MethodPost, Path: "/api/v1/focus/convert", Handler: focusConvertHandler(logger), Tags: []string{"focus", "jobs"}},
		{Method: http.MethodPost, Path: "/api/v1/focus/analyze", Handler: focusAnalyzeHandler(logger), Tags: []string{"focus", "jobs"}},
		{Method: http.MethodPost, Path: "/api/v1/focus/validate", Handler: focusValidateHandler(logger), Tags: []string{"focus", "validation"}},
		{Method: http.MethodGet, Path: "/api/v1/focus/schemas", Handler: focusSchemasHandler(logger), Tags: []string{"focus", "validation", "schemas"}},
		{Method: http.MethodGet, Path: "/api/v1/focus/datasets", Handler: focusDatasetsHandler(logger), Tags: []string{"focus"}},
		{Method: http.MethodGet, Path: "/api/v1/focus/jobs/", Handler: focusJobStatusHandler(logger), Tags: []string{"focus", "jobs"}}, // prefix match

		// Analytics
		{Method: http.MethodPost, Path: "/api/v1/analytics/analyze", Handler: analyticsAnalyzeHandler(logger), Tags: []string{"analytics", "jobs"}},
		{Method: http.MethodPost, Path: "/api/v1/analytics/forecast", Handler: analyticsForecastHandler(logger), Tags: []string{"analytics", "jobs"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/summary", Handler: analyticsSummaryHandler(logger), Tags: []string{"analytics"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/top-services", Handler: analyticsTopServicesHandler(logger), Tags: []string{"analytics"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/trends", Handler: analyticsTrendsHandler(logger), Tags: []string{"analytics"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/breakdown", Handler: analyticsBreakdownHandler(logger), Tags: []string{"analytics"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/anomalies", Handler: analyticsAnomaliesHandler(logger), Tags: []string{"analytics"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/optimize", Handler: analyticsOptimizeHandler(logger), Tags: []string{"analytics"}},
		{Method: http.MethodGet, Path: "/api/v1/analytics/metrics", Handler: analyticsMetricsHandler(logger), Tags: []string{"analytics", "metrics"}},

		// Multicloud (preview stub endpoints; enterprise server wires full service)
		{Method: http.MethodPost, Path: "/api/v1/multicloud/recommendations", Handler: multicloudRecommendationsHandler(logger), Tags: []string{"multicloud", "preview"}},
		{Method: http.MethodGet, Path: "/api/v1/multicloud/inventory", Handler: multicloudInventoryHandler(logger), Tags: []string{"multicloud", "preview"}},
		{Method: http.MethodPost, Path: "/api/v1/multicloud/migration/plan", Handler: multicloudMigrationPlanHandler(logger), Tags: []string{"multicloud", "preview"}},
		{Method: http.MethodPost, Path: "/api/v1/multicloud/migration/feasibility", Handler: multicloudMigrationFeasibilityHandler(logger), Tags: []string{"multicloud", "preview"}},
	}
}

// registerRouteSpecs attaches routes to the mux applying common + per-route middleware and logs a summary.
func registerRouteSpecs(mux *http.ServeMux, specs []RouteSpec, common []HTTPMiddleware, logger *logging.Logger) {
	totalRouteSpecificMW := 0
	skipped := 0
	for _, rs := range specs {
		if !gateEnabled(rs.FeatureGate) { // gate disabled
			skipped++
			continue
		}
		mws := append([]HTTPMiddleware{}, common...)
		if len(rs.Middleware) > 0 {
			mws = append(mws, rs.Middleware...)
			totalRouteSpecificMW += len(rs.Middleware)
		}
		// Enforce method first
		mws = append([]HTTPMiddleware{requireMethod(rs.Method)}, mws...)
		mux.Handle(rs.Path, ChainHTTP(rs.Handler, mws...))
	}
	logger.InfoWithFields("api_router_initialized", map[string]interface{}{
		"routes_total":            len(specs),
		"routes_skipped":          skipped,
		"routes_registered":       len(specs) - skipped,
		"common_middleware":       len(common),
		"route_specific_mw_total": totalRouteSpecificMW,
	})
}

// buildCommonMiddleware returns default middleware applied to every route.
func buildCommonMiddleware() []HTTPMiddleware {
	return []HTTPMiddleware{corsMiddleware, rateLimitMiddleware(rateLimitRequests, rateLimitWindow)}
}

// Simple CORS middleware (allow list defined by global corsOrigins).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use configured origins, default * if none
		origin := r.Header.Get("Origin")
		allowed := "*"
		// When specific origins configured (not just a solitary wildcard) and request has an Origin header, check allow list.
		if len(corsOrigins) > 0 && (len(corsOrigins) != 1 || corsOrigins[0] != "*") && origin != "" {
			for _, o := range corsOrigins {
				if o == origin {
					allowed = origin
					break
				}
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware provides a simple per-IP in-memory limiter for the net/http server.
// For production-grade limiting, prefer a distributed limiter; this is adequate for the basic serve mode.
func rateLimitMiddleware(max int, window time.Duration) HTTPMiddleware {
	type bucket struct{ times []time.Time }
	var (
		mu   sync.Mutex
		byIP = map[string]*bucket{}
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if max <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil || ip == "" {
				ip = r.RemoteAddr
			}
			now := time.Now()
			cutoff := now.Add(-window)
			mu.Lock()
			b := byIP[ip]
			if b == nil {
				b = &bucket{}
				byIP[ip] = b
			}
			// drop old
			filtered := b.times[:0]
			for _, t := range b.times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			b.times = filtered
			if len(b.times) >= max {
				mu.Unlock()
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(max))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(now.Add(window).Unix(), 10))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"Rate limit exceeded"}`))
				return
			}
			b.times = append(b.times, now)
			remaining := max - len(b.times)
			mu.Unlock()
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(max))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(now.Add(window).Unix(), 10))
			next.ServeHTTP(w, r)
		})
	}
}

func gateEnabled(gate string) bool {
	if gate == "" { // no gate specified
		return true
	}
	val := strings.ToLower(strings.TrimSpace(os.Getenv(gate)))
	if val == "" { // unset => enabled
		return true
	}
	switch val {
	case "0", "false", "off", "disable", "disabled":
		return false
	default:
		return true
	}
}

// local method enforcement reused in registration
// method enforcement via requireMethod from router_registry.go
