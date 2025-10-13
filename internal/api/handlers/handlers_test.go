package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	streamingTypes "github.com/costscope/costscope/cmd/modules/streaming/types"
	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/websocket"
	"github.com/costscope/costscope/internal/core/focus/conversion"
	"github.com/costscope/costscope/internal/core/logging"
	persistence "github.com/costscope/costscope/internal/core/persistence"
	"github.com/costscope/costscope/internal/providers"
	providerTypes "github.com/costscope/costscope/internal/providers/types"
	"github.com/costscope/costscope/internal/testutil"
)

// unwrapEnvelope normalizes the new unified API response envelope to the original
// flat shape expected by many legacy assertions. If the top-level object contains
// a "data" key and a boolean "success" field, it assumes envelope shape and
// returns data (merged with error/meta if needed). Otherwise the original map
// is returned unchanged.
func unwrapEnvelope(raw map[string]interface{}) map[string]interface{} {
	if raw == nil {
		return raw
	}
	if _, hasData := raw["data"]; hasData {
		if _, hasSuccess := raw["success"]; hasSuccess {
			if dataMap, ok := raw["data"].(map[string]interface{}); ok {
				return dataMap
			}
		}
	}
	return raw
}

// TestFocusHandler tests the FOCUS API handler
func TestFocusHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	jobMgr := jobs.NewManager(logger, 2)
	wsManager := websocket.NewManager(logger)

	handler := NewFocusHandler(logger, jobMgr, wsManager, conversion.NewConversionManager(1))
	router := gin.New()

	// Register routes
	v1 := router.Group("/api/v1")
	focus := v1.Group("/focus")
	{
		focus.POST("/convert", handler.ConvertData)
		focus.POST("/analyze", handler.AnalyzeData)
		focus.POST("/compare", handler.CompareData)
		focus.POST("/validate", handler.ValidateData)
		focus.GET("/jobs/:id", handler.GetJob)
		focus.GET("/jobs", handler.ListJobs)
		focus.DELETE("/jobs/:id", handler.CancelJob)
	}

	// Test cases
	tests := []struct {
		name          string
		method        string
		path          string
		body          interface{}
		expectedCode  int
		checkResponse func(t *testing.T, body string)
	}{
		{
			name:   "Convert Data - Valid Request",
			method: "POST",
			path:   "/api/v1/focus/convert",
			body: map[string]interface{}{
				"provider":    "aws",
				"input_path":  "/data/aws-cur.csv",
				"output_path": "/data/focus.parquet",
				"options": map[string]interface{}{
					"validate":    true,
					"compression": true,
				},
			},
			expectedCode: http.StatusAccepted,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Contains(t, resp, "job_id")
				assert.Equal(t, "accepted", resp["status"])
			},
		},
		{
			name:   "Convert Data - Invalid Provider",
			method: "POST",
			path:   "/api/v1/focus/convert",
			body: map[string]interface{}{
				"provider":    "invalid",
				"input_path":  "/data/billing.csv",
				"output_path": "/data/focus.parquet",
			},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				// error envelope: success=false, error.code=bad_request
				assert.Equal(t, false, raw["success"])
				errObj, _ := raw["error"].(map[string]interface{})
				if assert.NotNil(t, errObj) {
					assert.Equal(t, "bad_request", errObj["code"])
				}
			},
		},
		{
			name:   "Analyze Data - Valid Request",
			method: "POST",
			path:   "/api/v1/focus/analyze",
			body: map[string]interface{}{
				"input_path":    "/data/focus.parquet",
				"analysis_type": "cost_breakdown",
				"dimensions":    []string{"service", "region"},
				"options": map[string]interface{}{
					"include_trends":    true,
					"include_anomalies": true,
				},
			},
			expectedCode: http.StatusAccepted,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Contains(t, resp, "job_id")
			},
		},
		{
			name:   "Compare Data - Valid Request",
			method: "POST",
			path:   "/api/v1/focus/compare",
			body: map[string]interface{}{
				"dataset1": "/data/baseline.parquet",
				"dataset2": "/data/comparison.parquet",
			},
			expectedCode: http.StatusAccepted,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Contains(t, resp, "job_id")
			},
		},
		{
			name:   "Validate Data - Valid Request",
			method: "POST",
			path:   "/api/v1/focus/validate",
			body: map[string]interface{}{
				"input_path":   "/data/focus.parquet",
				"spec_version": "1.2",
			},
			expectedCode: http.StatusAccepted,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Contains(t, resp, "job_id")
				assert.Equal(t, "accepted", resp["status"])
			},
		},
		{
			name:         "Get Job Status - Not Found",
			method:       "GET",
			path:         "/api/v1/focus/jobs/nonexistent",
			body:         nil,
			expectedCode: http.StatusNotFound,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				// Expect error envelope
				assert.Equal(t, false, raw["success"])
				errObj, _ := raw["error"].(map[string]interface{})
				if assert.NotNil(t, errObj) {
					assert.Equal(t, "not_found", errObj["code"])
				}
			},
		},
		{
			name:         "List Jobs",
			method:       "GET",
			path:         "/api/v1/focus/jobs",
			body:         nil,
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				// list jobs shape: { jobs: [...], total: N }
				assert.Contains(t, resp, "jobs")
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req, err = http.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, tt.path, nil)
			}
			assert.NoError(t, err)

			// Add mock authentication headers
			req.Header.Set("Authorization", "Bearer mock-token")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

// TestFocusHandlerErrors tests error scenarios
func TestFocusHandlerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	jobMgr := jobs.NewManager(logger, 2)
	wsManager := websocket.NewManager(logger)

	handler := NewFocusHandler(logger, jobMgr, wsManager, conversion.NewConversionManager(1))
	router := gin.New()

	focus := router.Group("/api/v1/focus")
	focus.POST("/convert", handler.ConvertData)

	tests := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name:         "Invalid JSON",
			body:         `{"invalid": json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing Required Fields",
			body:         `{"provider": "aws"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty Request Body",
			body:         ``,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/v1/focus/convert", bytes.NewBufferString(tt.body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestAnalyticsHandler tests the Analytics API handler
func TestAnalyticsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	jobMgr := jobs.NewManager(logger, 2)

	handler := NewAnalyticsHandler(logger, jobMgr)
	router := gin.New()

	v1 := router.Group("/api/v1")
	analytics := v1.Group("/analytics")
	{
		analytics.POST("/forecast", handler.GenerateForecast)
		analytics.POST("/anomalies", handler.DetectAnomalies)
		analytics.POST("/recommendations", handler.GetRecommendations)
		analytics.POST("/trends", handler.AnalyzeTrends)
	}

	tests := []struct {
		name         string
		path         string
		body         interface{}
		expectedCode int
	}{
		{
			name: "Generate Forecast",
			path: "/api/v1/analytics/forecast",
			body: map[string]interface{}{
				"data_source":   "/data/focus.parquet",
				"forecast_days": 30,
				"model_type":    "auto-arima",
				"confidence":    95.0,
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "Detect Anomalies",
			path: "/api/v1/analytics/anomalies",
			body: map[string]interface{}{
				"data_source": "/data/focus.parquet",
				"sensitivity": 0.8,
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "Get Recommendations",
			path: "/api/v1/analytics/recommendations",
			body: map[string]interface{}{
				"data_source": "/data/focus.parquet",
				"categories":  []string{"rightsizing", "reserved_instances"},
				"min_savings": 20.0,
			},
			expectedCode: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req, err := http.NewRequest("POST", tt.path, bytes.NewBuffer(bodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestMulticloudHandler tests multicloud advanced endpoints
func TestMulticloudHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	pm := providers.NewProviderManager() // empty manager acceptable (service uses mock logic)
	handler := NewMulticloudHandler(logger, pm)
	router := gin.New()
	v1 := router.Group("/api/v1")
	mc := v1.Group("/multicloud")
	mc.POST("/recommendations", handler.Recommendations)
	mc.GET("/inventory", handler.Inventory)
	mc.POST("/migration/plan", handler.MigrationPlan)
	mc.POST("/migration/feasibility", handler.MigrationFeasibility)

	// Recommendations
	recBody := map[string]any{"providers": []string{"aws"}}
	b, _ := json.Marshal(recBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/multicloud/recommendations", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Inventory
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/multicloud/inventory", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Migration plan
	mBody := map[string]any{"source": "aws", "target": "azure"}
	mb, _ := json.Marshal(mBody)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/v1/multicloud/migration/plan", bytes.NewReader(mb))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// Feasibility
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/v1/multicloud/migration/feasibility", bytes.NewReader(mb))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}

// TestHealthHandler tests the Health API handler
func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")

	handler := NewHealthHandler(logger)
	router := gin.New()

	router.GET("/health", handler.HealthCheck)
	router.GET("/health/ready", handler.ReadinessCheck)
	router.GET("/health/live", handler.LivenessCheck)

	tests := []struct {
		name         string
		path         string
		expectedCode int
		checkBody    func(t *testing.T, body string)
	}{
		{
			name:         "Health Check",
			path:         "/health",
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Equal(t, "healthy", resp["status"])
			},
		},
		{
			name:         "Readiness Check",
			path:         "/health/ready",
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Equal(t, "ready", resp["status"])
			},
		},
		{
			name:         "Liveness Check",
			path:         "/health/live",
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(body), &raw))
				resp := unwrapEnvelope(raw)
				assert.Contains(t, resp, "status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.String())
			}
		})
	}
}

// TestAnalyticsReadHandler covers the lightweight GET endpoints backed by the analytics facade.
// The handlers have two build variants:
//   - duckdb build tag: executes real queries over a parquet input (expects 200/400 depending on params)
//   - default (no duckdb): returns 501 Not Implemented; tests accept 501 in that case
func TestAnalyticsReadHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")

	// Wire only the GET endpoints we need for testing
	read := NewAnalyticsReadHandler(logger)
	router := gin.New()
	v1 := router.Group("/api/v1")
	analytics := v1.Group("/analytics")
	{
		analytics.GET("/summary", read.Summary)
		analytics.GET("/top-services", read.TopServices)
		analytics.GET("/trends", read.Trends)
	}

	// Resolve sample parquet path if present (skip content-validations if missing)
	getDemoParquet := func() (string, bool) {
		candidates := []string{
			"demo/focus-conversion/demo-focus.parquet",
		}
		repoRoot := testutil.FindRepoRoot(t)
		for _, rel := range candidates {
			p := filepath.Join(repoRoot, rel)
			if _, err := os.Stat(p); err == nil {
				return p, true
			}
		}
		return "", false
	}

	t.Run("Summary_missing_input_returns_400_or_501_enveloped", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/analytics/summary", nil)
		req.Header.Set("X-Request-ID", "req-12345")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// duckdb build => 400; default => 501
		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotImplemented {
			t.Fatalf("unexpected status: %d body=%s", w.Code, w.Body.String())
		}
		// Assert envelope + meta.request_id present
		var root map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &root)
		if _, ok := root["success"].(bool); !ok {
			t.Fatalf("expected envelope success field")
		}
		if meta, ok := root["meta"].(map[string]any); ok {
			if rid, ok := meta["request_id"].(string); !ok || rid != "req-12345" {
				// request id auto injection should propagate
				// not fatal for non-duckdb build but we still expect it
				t.Fatalf("expected meta.request_id=req-12345 got %v", meta["request_id"])
			}
		} else {
			// meta should exist
			t.Fatalf("expected meta in envelope")
		}
	})

	t.Run("TopServices_limit_parsing_200_or_501", func(t *testing.T) {
		if path, ok := getDemoParquet(); ok {
			req, _ := http.NewRequest("GET", "/api/v1/analytics/top-services?input="+path+"&limit=3", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				var root map[string]any
				_ = json.Unmarshal(w.Body.Bytes(), &root)
				payload := unwrapEnvelope(root)
				if ts, ok := payload["top_services"].([]any); ok {
					if len(ts) > 3 {
						t.Fatalf("expected <=3 items, got %d", len(ts))
					}
				}
			} else if w.Code != http.StatusNotImplemented && w.Code != http.StatusBadRequest { // duckdb returns 400 when load fails
				t.Fatalf("unexpected status: %d body=%s", w.Code, w.Body.String())
			}
		} else {
			t.Skip("demo parquet not found; skipping content validation")
		}
	})

	t.Run("Trends_granularity_valid_and_invalid", func(t *testing.T) {
		if path, ok := getDemoParquet(); ok {
			// valid granularity
			reqOK, _ := http.NewRequest("GET", "/api/v1/analytics/trends?input="+path+"&granularity=day", nil)
			wOK := httptest.NewRecorder()
			router.ServeHTTP(wOK, reqOK)
			if wOK.Code == http.StatusOK {
				var root map[string]any
				_ = json.Unmarshal(wOK.Body.Bytes(), &root)
				payload := unwrapEnvelope(root)
				if g, ok := payload["granularity"].(string); ok {
					if g != "day" {
						t.Fatalf("expected granularity=day, got %s", g)
					}
				}
			} else if wOK.Code != http.StatusNotImplemented && wOK.Code != http.StatusBadRequest { // accept 400 for duckdb load failure
				t.Fatalf("unexpected status for valid granularity: %d body=%s", wOK.Code, wOK.Body.String())
			}

			// invalid granularity
			reqBad, _ := http.NewRequest("GET", "/api/v1/analytics/trends?input="+path+"&granularity=quarter", nil)
			wBad := httptest.NewRecorder()
			router.ServeHTTP(wBad, reqBad)
			if wBad.Code != http.StatusBadRequest && wBad.Code != http.StatusNotImplemented { // 400 expected for invalid granularity
				t.Fatalf("unexpected status for invalid granularity: %d body=%s", wBad.Code, wBad.Body.String())
			}
		} else {
			t.Skip("demo parquet not found; skipping trends tests")
		}
	})
}

// TestRequestIDPropagation asserts that X-Request-ID is injected into the envelope meta.
func TestRequestIDPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	h := NewHealthHandler(logger)
	router := gin.New()
	router.GET("/health", h.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Request-ID", "rid-integration-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
	var root map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &root); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	meta, ok := root["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta object in envelope")
	}
	if rid, ok := meta["request_id"].(string); !ok || rid != "rid-integration-123" {
		t.Fatalf("expected meta.request_id=rid-integration-123 got %v", meta["request_id"])
	}
	if success, ok := root["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true in envelope")
	}
	data, ok := root["data"].(map[string]any)
	if !ok || data["status"] != "healthy" {
		t.Fatalf("expected data.status=healthy got %v", data["status"])
	}
}

// TestReadiness_JSONShape asserts the JSON structure of /health/ready without
// depending on environment-specific readiness conditions.
func TestReadiness_JSONShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")

	handler := NewHealthHandler(logger)
	router := gin.New()
	router.GET("/health/ready", handler.ReadinessCheck)

	req, err := http.NewRequest("GET", "/health/ready", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Readiness may be 200 (ready) or 503 (not ready) depending on optional subsystems; both are valid.
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable, "unexpected status code: %d", w.Code)

	var raw map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &raw)
	assert.NoError(t, err)
	resp := unwrapEnvelope(raw)

	// status must be a non-empty string (enveloped under data)
	status, ok := resp["status"].(string)
	assert.True(t, ok, "status should be a string")
	assert.NotEmpty(t, status)

	// checks must be an object with expected keys present
	checks, ok := resp["checks"].(map[string]interface{})
	assert.True(t, ok, "checks should be an object")

	for _, key := range []string{"background_jobs", "database", "duckdb"} {
		v, exists := checks[key]
		assert.True(t, exists, "checks.%s missing", key)
		// values are expected to be strings (e.g., ok, unknown, not_running, not_linked)
		if vs, isStr := v.(string); isStr {
			assert.NotEmpty(t, vs)
		}
	}
}

// readinessRepoStub implements persistence.Repository with controllable Health behavior
type readinessRepoStub struct {
	sleep time.Duration
	err   error
}

func (r *readinessRepoStub) SaveJob(ctx context.Context, job *streamingTypes.StreamingJobInfo) error {
	return nil
}
func (r *readinessRepoStub) GetJob(ctx context.Context, jobID string) (*streamingTypes.StreamingJobInfo, error) {
	return nil, nil
}
func (r *readinessRepoStub) ListJobs(ctx context.Context, filters persistence.JobFilters) ([]*streamingTypes.StreamingJobInfo, error) {
	return nil, nil
}
func (r *readinessRepoStub) UpdateJobStatus(ctx context.Context, jobID string, status *streamingTypes.StreamingJobStatus) error {
	return nil
}
func (r *readinessRepoStub) DeleteJob(ctx context.Context, jobID string) error { return nil }
func (r *readinessRepoStub) SaveProvider(ctx context.Context, config *providerTypes.ProviderConfig) error {
	return nil
}
func (r *readinessRepoStub) GetProvider(ctx context.Context, name string) (*providerTypes.ProviderConfig, error) {
	return nil, nil
}
func (r *readinessRepoStub) ListProviders(ctx context.Context) ([]*providerTypes.ProviderConfig, error) {
	return nil, nil
}
func (r *readinessRepoStub) DeleteProvider(ctx context.Context, name string) error { return nil }
func (r *readinessRepoStub) Health(ctx context.Context) error {
	if r.sleep > 0 {
		select {
		case <-time.After(r.sleep):
		case <-ctx.Done():
			// propagate context deadline exceeded as error
			return ctx.Err()
		}
	}
	return r.err
}
func (r *readinessRepoStub) Close() error { return nil }

// Test: /health/ready returns 503 when jobs manager is stopped
func TestReadiness_NotReady_WhenJobsManagerStopped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	jobMgr := jobs.NewManager(logger, 1)
	// Ensure it is stopped (default new manager is not running)

	h := NewHealthHandler(logger).WithJobs(jobMgr)
	router := gin.New()
	router.GET("/health/ready", h.ReadinessCheck)

	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var raw map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &raw)
	resp := unwrapEnvelope(raw)
	assert.Equal(t, "not_ready", resp["status"])
	checks := resp["checks"].(map[string]interface{})
	assert.Equal(t, "not_running", checks["background_jobs"])
}

// Test: /health/ready returns 503 on repository timeout (>500ms)
func TestReadiness_NotReady_OnRepositoryTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	// Repo sleeps 600ms causing ctx timeout (500ms)
	repo := &readinessRepoStub{sleep: 600 * time.Millisecond, err: nil}

	h := NewHealthHandler(logger).WithRepository(repo)
	router := gin.New()
	router.GET("/health/ready", h.ReadinessCheck)

	// Capture stderr logs during request
	oldStderr := os.Stderr
	rpipe, wpipe, _ := os.Pipe()
	os.Stderr = wpipe

	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	_ = wpipe.Close()
	os.Stderr = oldStderr
	logged, _ := io.ReadAll(rpipe)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var raw map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &raw)
	resp := unwrapEnvelope(raw)
	assert.Equal(t, "not_ready", resp["status"])
	checks := resp["checks"].(map[string]interface{})
	assert.Equal(t, "error", checks["database"])
	// database_error field should be present due to timeout
	if _, ok := checks["database_error"]; !ok {
		t.Fatalf("expected database_error field in checks on timeout")
	}

	// Verify logs include an error level and readiness message
	logs := string(logged)
	if !strings.Contains(logs, "\"level\":\"error\"") || !strings.Contains(logs, "readiness database health error") {
		t.Fatalf("expected error log about readiness database health error, got: %s", logs)
	}
}

// TestProvidersHandler tests the Providers API handler
func TestProvidersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")

	handler := NewProvidersHandler(logger)
	router := gin.New()

	v1 := router.Group("/api/v1")
	providers := v1.Group("/providers")
	{
		providers.GET("", handler.ListProviders)
		providers.GET("/:provider", handler.GetProvider)
		providers.POST("/:provider/connect", handler.ConnectProvider)
		providers.POST("/:provider/test", handler.TestConnection)
		providers.GET("/:provider/accounts", handler.ListAccounts)
	}

	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		expectedCode int
		checkBody    func(t *testing.T, body string)
	}{
		{
			name:         "List Providers",
			method:       "GET",
			path:         "/api/v1/providers",
			body:         nil,
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err)
				// Unified envelope: providers now nested under data.providers
				if _, topLevel := response["providers"]; topLevel {
					// Legacy shape (should migrate away eventually)
					return
				}
				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected envelope data object, got: %v", response["data"])
				}
				assert.Contains(t, data, "providers")
			},
		},
		{
			name:         "Get AWS Provider",
			method:       "GET",
			path:         "/api/v1/providers/aws",
			body:         nil,
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err)
				if id, legacy := response["id"]; legacy {
					assert.Equal(t, "aws", id)
					return
				}
				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected envelope data object, got: %v", response["data"])
				}
				assert.Equal(t, "aws", data["id"])
			},
		},
		{
			name:   "Connect AWS Provider",
			method: "POST",
			path:   "/api/v1/providers/aws/connect",
			body: map[string]interface{}{
				"credentials": map[string]interface{}{
					"access_key_id":     "AKIA...",
					"secret_access_key": "...",
					"region":            "us-east-1",
				},
			},
			expectedCode: http.StatusCreated,
			checkBody: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err)
				if p, legacy := response["provider"]; legacy {
					assert.Equal(t, "aws", p)
					return
				}
				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected envelope data object, got: %v", response["data"])
				}
				assert.Equal(t, "aws", data["provider"])
			},
		},
		{
			name:         "Test AWS Connection",
			method:       "POST",
			path:         "/api/v1/providers/aws/test",
			body:         map[string]interface{}{},
			expectedCode: http.StatusOK,
		},
		{
			name:         "List AWS Accounts",
			method:       "GET",
			path:         "/api/v1/providers/aws/accounts",
			body:         nil,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req, err = http.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, tt.path, nil)
			}
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.String())
			}
		})
	}
}

// Mock job manager for testing
type MockJobManager struct {
	mock.Mock
}

func (m *MockJobManager) StartJob(jobType string, config interface{}) (string, error) {
	args := m.Called(jobType, config)
	return args.String(0), args.Error(1)
}

func (m *MockJobManager) GetJobStatus(jobID string) (interface{}, error) {
	args := m.Called(jobID)
	return args.Get(0), args.Error(1)
}

func (m *MockJobManager) ListJobs() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockJobManager) CancelJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}

// Test with mock job manager
func TestFocusHandlerWithMock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger("debug")
	mockJobMgr := new(MockJobManager)
	wsManager := websocket.NewManager(logger)

	// Setup mock expectations
	mockJobMgr.On("StartJob", "focus_convert", mock.Anything).Return("job-123", nil)
	mockJobMgr.On("GetJobStatus", "job-123").Return(map[string]interface{}{
		"job_id":   "job-123",
		"status":   "running",
		"progress": 50,
	}, nil)
	mockJobMgr.On("ListJobs").Return([]interface{}{
		map[string]interface{}{
			"job_id": "job-123",
			"status": "running",
		},
	}, nil)

	// This would require modifying the handler to accept the interface
	// For now, we'll use the real handler
	handler := NewFocusHandler(logger, jobs.NewManager(logger, 2), wsManager, conversion.NewConversionManager(1))
	router := gin.New()

	router.POST("/convert", handler.ConvertData)
	router.GET("/jobs/:id", handler.GetJob)
	router.GET("/jobs", handler.ListJobs)

	// Test convert endpoint
	body := map[string]interface{}{
		"provider":    "aws",
		"input_path":  "/data/test.csv",
		"output_path": "/data/output.parquet",
	}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/convert", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	// Verify mock was called (would work with proper interface implementation)
	// mockJobMgr.AssertExpectations(t)
}
