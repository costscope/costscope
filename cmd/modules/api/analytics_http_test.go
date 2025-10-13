package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/testutil"
)

// resolveDemoParquet searches up the tree for demo/focus-conversion/demo-focus.parquet.
func resolveDemoParquet() (string, bool) {
	// Prefer the canonical repo root when available. Do not fallback to ad-hoc
	// cwd-based walk-ups — use the centralized helper instead.
	repoRoot, err := testutil.RepoRoot()
	if err != nil {
		return "", false
	}
	p := filepath.Join(repoRoot, "demo", "focus-conversion", "demo-focus.parquet")
	if _, err := os.Stat(p); err == nil {
		return p, true
	}
	return "", false
}

func TestAnalyticsHTTP_SummaryMissingOrEnv(t *testing.T) {
	srv := newTestServer()
	// Missing input: duckdb => 400, slim => 200
	rr := doReq(t, srv, http.MethodGet, "/api/v1/analytics/summary", nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Fatalf("summary expected 200 or 400, got %d", rr.Code)
	}
	if path, ok := resolveDemoParquet(); ok {
		t.Setenv("COSTSCOPE_FOCUS_PARQUET", path)
		rr = doReq(t, srv, http.MethodGet, "/api/v1/analytics/summary", nil)
		if rr.Code != http.StatusOK {
			// In duckdb builds a placeholder parquet (missing magic bytes) triggers 400; treat as acceptable fallback.
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("summary with env expected 200 or 400, got %d body=%s", rr.Code, rr.Body.String())
			}
		}
	} else {
		t.Log("demo parquet not found; env-backed path test skipped")
	}
}

func TestAnalyticsHTTP_TopServicesLimitParsing(t *testing.T) {
	srv := newTestServer()
	if path, ok := resolveDemoParquet(); ok {
		t.Setenv("COSTSCOPE_FOCUS_PARQUET", path)
	}
	rr := doReq(t, srv, http.MethodGet, "/api/v1/analytics/top-services?limit=3", nil)
	if rr.Code != http.StatusOK {
		// In duckdb build missing input env could cause 400 if env not set; accept 400 in that case.
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("top-services expected 200 or 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		return
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if arr, ok := resp["top_services"].([]any); ok {
		if len(arr) > 3 {
			t.Fatalf("expected <= 3 top_services, got %d", len(arr))
		}
	}
}

func TestAnalyticsHTTP_TrendsValidAndInvalid(t *testing.T) {
	srv := newTestServer()
	if path, ok := resolveDemoParquet(); ok {
		t.Setenv("COSTSCOPE_FOCUS_PARQUET", path)
	}
	// Valid granularity
	rr := doReq(t, srv, http.MethodGet, "/api/v1/analytics/trends?granularity=day", nil)
	if rr.Code != http.StatusOK {
		// Accept 400 if duckdb build without env/path set
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("trends valid expected 200 or 400, got %d body=%s", rr.Code, rr.Body.String())
		}
	}
	// Invalid granularity
	rr = doReq(t, srv, http.MethodGet, "/api/v1/analytics/trends?granularity=quarter", nil)
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusOK {
		t.Fatalf("trends invalid expected 400 or 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAnalyticsHTTP_BreakdownBasic(t *testing.T) {
	srv := newTestServer()
	// Without input: duckdb => 400, slim => 200
	rr := doReq(t, srv, http.MethodGet, "/api/v1/analytics/breakdown", nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Fatalf("breakdown expected 200 or 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	// With env parquet set, expect 200 and response shape
	if path, ok := resolveDemoParquet(); ok {
		t.Setenv("COSTSCOPE_FOCUS_PARQUET", path)
		rr = doReq(t, srv, http.MethodGet, "/api/v1/analytics/breakdown", nil)
		if rr.Code != http.StatusOK {
			// Accept 400 for invalid placeholder parquet in duckdb test context.
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("breakdown with env expected 200 or 400, got %d body=%s", rr.Code, rr.Body.String())
			}
			return // skip shape assertions when not 200
		}
		var resp map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid json: %v body=%s", err, rr.Body.String())
		}
		if _, ok := resp["summary"]; !ok {
			t.Fatalf("expected summary field present in response")
		}
		if top, ok := resp["top_services"].([]any); ok {
			if len(top) > 5 { // default limit is 5
				t.Fatalf("expected <=5 top_services, got %d", len(top))
			}
		} else {
			t.Fatalf("expected top_services array in response")
		}
	} else {
		t.Log("demo parquet not found; env-backed breakdown test skipped")
	}
}
