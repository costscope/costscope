//go:build casbinpoc
// +build casbinpoc

package security

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/golang-jwt/jwt/v5"
)

// NewEnforcerFromFiles loads a Casbin enforcer from model and policy file paths.
func NewEnforcerFromFiles(modelPath, policyPath string) (*casbin.Enforcer, error) {
	if modelPath == "" || policyPath == "" {
		return nil, errors.New("casbin modelPath and policyPath are required")
	}
	a := fileadapter.NewAdapter(policyPath)
	return casbin.NewEnforcer(modelPath, a)
}

// SubjectExtractor extracts a subject (e.g., role or user) from request.
type SubjectExtractor func(r *http.Request) (subjects []string)

// JWTSubjectExtractor returns roles from a Bearer JWT using provided secret and issuer.
func JWTSubjectExtractor(secret, issuer string) SubjectExtractor {
	return func(r *http.Request) []string {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return nil
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		type claims struct {
			Roles []string `json:"roles"`
			jwt.RegisteredClaims
		}
		token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return nil
		}
		c, ok := token.Claims.(*claims)
		if !ok {
			return nil
		}
		if issuer != "" && c.Issuer != issuer {
			return nil
		}
		// Normalize roles to include role: prefix for policies like role:admin
		out := make([]string, 0, len(c.Roles))
		for _, r := range c.Roles {
			if r == "" {
				continue
			}
			if strings.HasPrefix(r, "role:") {
				out = append(out, r)
			} else {
				out = append(out, "role:"+r)
			}
		}
		return out
	}
}

// HeaderSubjectExtractor extracts a single subject from X-Subject header (PoC/dev only).
func HeaderSubjectExtractor() SubjectExtractor {
	return func(r *http.Request) []string {
		v := r.Header.Get("X-Subject")
		if v == "" {
			return nil
		}
		return []string{v}
	}
}

// CasbinHTTPMiddleware wraps an http.Handler with Casbin authorization.
// It authorizes if ANY extracted subject is allowed for (obj=path, act=method).
func CasbinHTTPMiddleware(next http.Handler, e *casbin.Enforcer, extract SubjectExtractor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health/metrics unauthenticated by default
		path := r.URL.Path
		if path == "/health" || strings.HasPrefix(path, "/metrics") || strings.HasPrefix(path, "/docs") {
			next.ServeHTTP(w, r)
			return
		}

		subjects := extract(r)
		// Optional anonymous subject support via env var
		if len(subjects) == 0 {
			if anon := os.Getenv("CASBIN_ANON_SUBJECT"); anon != "" {
				subjects = []string{anon}
			}
		}
		if len(subjects) == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Determine optional domain (header takes precedence, then env)
		domain := r.Header.Get("X-Casbin-Domain")
		if domain == "" {
			domain = os.Getenv("CASBIN_DOMAIN")
			if domain == "" {
				domain = os.Getenv("CASBIN_DEFAULT_DOMAIN")
			}
		}

		for _, sub := range subjects {
			var ok bool
			var err error
			if domain != "" {
				ok, err = e.Enforce(sub, domain, path, r.Method)
				if err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			ok, err = e.Enforce(sub, path, r.Method)
			if err == nil && ok {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}
