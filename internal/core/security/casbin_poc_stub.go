//go:build !casbinpoc
// +build !casbinpoc

package security

import (
	"errors"
	"net/http"

	"github.com/casbin/casbin/v2"
)

// SubjectExtractor extracts a subject (e.g., role or user) from request.
// Default stub returns nil.
type SubjectExtractor func(r *http.Request) (subjects []string)

// NewEnforcerFromFiles is a stub when Casbin PoC is disabled.
func NewEnforcerFromFiles(modelPath, policyPath string) (*casbin.Enforcer, error) {
	return nil, errors.New("casbin PoC disabled: build with -tags casbinpoc to enable")
}

// JWTSubjectExtractor stub returns an extractor that yields no subjects.
func JWTSubjectExtractor(secret, issuer string) SubjectExtractor {
	return func(r *http.Request) []string { return nil }
}

// CasbinHTTPMiddleware stub passes through without authorization checks.
func CasbinHTTPMiddleware(next http.Handler, _ *casbin.Enforcer, _ SubjectExtractor) http.Handler {
	return next
}
