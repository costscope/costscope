package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Envelope is the unified top-level API response wrapper (additive; optional use).
// Fields are intentionally minimal to avoid breaking existing clients that rely
// on current ad-hoc shapes. Handlers can embed arbitrary domain payloads inside Data.
type Envelope[T any] struct {
	Data    T      `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Success bool   `json:"success"`
}

// Meta carries standard metadata emitted with responses.
type Meta struct {
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
	Source    string         `json:"source,omitempty"` // e.g. handler name
	Extra     map[string]any `json:"extra,omitempty"`
}

// Error represents a structured error payload.
type Error struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// OK writes a success envelope with optional meta extras.
func OK[T any](c *gin.Context, status int, payload T, opts ...Option) {
	env := &Envelope[T]{Data: payload, Success: true, Meta: &Meta{Timestamp: time.Now()}}
	applyOptions(env, opts)
	c.JSON(status, env)
}

// Fail writes an error envelope.
func Fail(c *gin.Context, status int, msg string, code string, opts ...Option) {
	env := &Envelope[struct{}]{Error: &Error{Message: msg, Code: code}, Success: false, Meta: &Meta{Timestamp: time.Now()}}
	applyOptions(env, opts)
	c.JSON(status, env)
}

// Option mutates an Envelope before write.
type Option func(meta *Meta)

// WithRequestID attaches a request id.
func WithRequestID(id string) Option {
	return func(m *Meta) {
		if m != nil {
			m.RequestID = id
		}
	}
}

// applyOptions applies option functions to the envelope's meta.
func applyOptions[T any](env *Envelope[T], opts []Option) {
	if env.Meta == nil {
		env.Meta = &Meta{Timestamp: time.Now()}
	}
	for _, o := range opts {
		o(env.Meta)
	}
}

// Convenience shims for common statuses.
// Helper shortcuts retained for prospective future API envelope consolidation.
// Excluded from default build to prevent deadcode reports.
// Convenience helpers (some may appear unused in direct code paths but kept for clarity in handlers).
// nolint:unused // wrappers around OK/Fail retained for explicit status semantics.
func OK200[T any](c *gin.Context, payload T) { OK(c, http.StatusOK, payload) }

// ==== Auto helpers (inject request id if present) ===================================

// AutoOK writes an OK envelope and attaches X-Request-ID if the middleware populated it.
func AutoOK[T any](c *gin.Context, status int, payload T) {
	rid := c.GetHeader("X-Request-ID")
	if rid != "" {
		OK(c, status, payload, WithRequestID(rid))
		return
	}
	OK(c, status, payload)
}

// AutoFail writes an error envelope and attaches X-Request-ID if present.
func AutoFail(c *gin.Context, status int, msg, code string) {
	rid := c.GetHeader("X-Request-ID")
	if rid != "" {
		Fail(c, status, msg, code, WithRequestID(rid))
		return
	}
	Fail(c, status, msg, code)
}

// AutoOK200 shortcut for AutoOK with 200 status.
func AutoOK200[T any](c *gin.Context, payload T) { AutoOK(c, http.StatusOK, payload) }

// AutoCreated201 writes a 201 Created success envelope with request id if present.
func AutoCreated201[T any](c *gin.Context, payload T) { AutoOK(c, http.StatusCreated, payload) }

// AutoNoContent204 writes a 204 with no body (true semantic No Content). Clients should not expect JSON.
func AutoNoContent204(c *gin.Context) { c.Status(http.StatusNoContent) }

// AutoNotFound404 writes a standardized 404 error envelope.
func AutoNotFound404(c *gin.Context, msg string) { AutoFail(c, http.StatusNotFound, msg, "not_found") }

// AutoBadRequest writes a standardized 400 error envelope (code=bad_request) with request id if present.
func AutoBadRequest(c *gin.Context, msg string) {
	AutoFail(c, http.StatusBadRequest, msg, ErrCodeBadRequest)
}

// AutoBadRequestCode writes a 400 error envelope with a custom error code while still attaching request id.
// Use when callers benefit from distinguishing specific validation failures beyond generic bad_request.
func AutoBadRequestCode(c *gin.Context, msg, code string) {
	if code == "" {
		code = ErrCodeBadRequest
	}
	AutoFail(c, http.StatusBadRequest, msg, code)
}
