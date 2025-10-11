package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOKAndFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Success path
	{
		r := gin.New()
		r.GET("/ok", func(c *gin.Context) {
			OK200(c, gin.H{"value": 42})
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/ok", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", w.Code)
		}
		var env struct {
			Data struct {
				Value int `json:"value"`
			} `json:"data"`
			Success bool `json:"success"`
			Meta    struct {
				Timestamp string `json:"timestamp"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !env.Success || env.Data.Value != 42 {
			t.Fatalf("unexpected payload: %+v", env)
		}
		if env.Meta.Timestamp == "" {
			t.Fatalf("timestamp missing")
		}
	}
	// Error path
	{
		r := gin.New()
		r.GET("/err", func(c *gin.Context) { Fail(c, http.StatusBadRequest, "boom", "bad") })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/err", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", w.Code)
		}
		var env struct {
			Error   struct{ Message, Code string } `json:"error"`
			Success bool                           `json:"success"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.Success {
			t.Fatalf("expected success=false")
		}
		if env.Error.Message != "boom" || env.Error.Code != "bad" {
			t.Fatalf("unexpected error payload: %+v", env)
		}
	}
}

func TestWithRequestIDOption(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/rid", func(c *gin.Context) { OK(c, http.StatusOK, gin.H{"ok": true}, WithRequestID("req-123")) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/rid", nil))
	var env struct {
		Meta struct {
			RequestID string `json:"request_id"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Meta.RequestID != "req-123" {
		t.Fatalf("expected request_id=req-123 got %s", env.Meta.RequestID)
	}
}

func TestAutoNoContentAndNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 204 path (no body expected)
	{
		r := gin.New()
		r.GET("/empty", func(c *gin.Context) { AutoNoContent204(c) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/empty", nil))
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204 got %d", w.Code)
		}
		if len(w.Body.Bytes()) != 0 {
			t.Fatalf("expected empty body for 204, got %q", w.Body.String())
		}
	}
	// 404 path with request id propagation
	{
		r := gin.New()
		r.GET("/nf", func(c *gin.Context) { AutoNotFound404(c, "missing") })
		req := httptest.NewRequest("GET", "/nf", nil)
		req.Header.Set("X-Request-ID", "rid-xyz")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", w.Code)
		}
		var env struct {
			Error struct{ Message, Code string }
			Meta  struct {
				RequestID string `json:"request_id"`
			}
			Success bool
		}
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.Success {
			t.Fatalf("expected success=false")
		}
		if env.Error.Message != "missing" || env.Error.Code != "not_found" {
			t.Fatalf("unexpected error payload: %+v", env)
		}
		if env.Meta.RequestID != "rid-xyz" {
			t.Fatalf("expected request id propagation, got %s", env.Meta.RequestID)
		}
	}
}

func TestAutoCreatedAndBadRequestCode(t *testing.T) {
	r := gin.Default()
	r.POST("/c", func(c *gin.Context) { AutoCreated201(c, gin.H{"id": 123}) })
	r.POST("/e", func(c *gin.Context) { AutoBadRequestCode(c, "missing input", ErrCodeMissingInput) })

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/c", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w1.Code)
	}
	var envCreated Envelope[map[string]any]
	if err := json.Unmarshal(w1.Body.Bytes(), &envCreated); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if !envCreated.Success || envCreated.Data["id"].(float64) != 123 {
		t.Fatalf("unexpected created envelope: %+v", envCreated)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/e", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w2.Code)
	}
	var envErr Envelope[struct{}]
	if err := json.Unmarshal(w2.Body.Bytes(), &envErr); err != nil {
		t.Fatalf("unmarshal error env: %v", err)
	}
	if envErr.Success {
		t.Fatalf("expected success=false: %+v", envErr)
	}
	if envErr.Error == nil || envErr.Error.Code != ErrCodeMissingInput {
		t.Fatalf("expected code %s, got %+v", ErrCodeMissingInput, envErr.Error)
	}
}
