package middleware

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/core/config"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/multitenant"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// =====================================================================================
// Enterprise Security Middleware - JWT, API Keys, Rate Limiting, Security Headers
// =====================================================================================

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		// Propagate to request context for downstream libraries (logging, tracing)
		ctx := logging.ContextWithIDs(c.Request.Context(), requestID, "", "")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		// X-XSS-Protection is deprecated in modern browsers; omit to follow current best practices
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimit implements simple rate limiting
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	// Simple in-memory rate limiting (for production, use Redis or similar)
	requestCounts := make(map[string]int)
	lastReset := time.Now()
	var mu sync.Mutex

	return func(c *gin.Context) {
		now := time.Now()
		mu.Lock()
		if now.Sub(lastReset) > window {
			requestCounts = make(map[string]int)
			lastReset = now
		}
		clientIP := c.ClientIP()
		requestCounts[clientIP]++
		remaining := requests - requestCounts[clientIP]
		resetAt := lastReset.Add(window).Unix()
		overLimit := requestCounts[clientIP] > requests
		mu.Unlock()

		// Set standard headers (string values)
		c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if overLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests from this IP address",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// JWTAuth validates JWT tokens
func JWTAuth(secret, issuer string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if strings.TrimSpace(secret) == "" {
			// Misconfiguration: do not accept tokens without a validation key
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT auth not configured"})
			c.Abort()
			return
		}

		// Parse and validate JWT (HS* by default using provided secret)
		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			// Only allow HMAC unless extended to support RSA/ECDSA with proper keys
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		},
			jwt.WithLeeway(30*time.Second),
			jwt.WithIssuer(issuer), // no-op if issuer == ""
			jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
		)
		if err != nil || parsed == nil || !parsed.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract common fields
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			c.Set("user_id", sub)
		}
		if role, ok := claims["role"].(string); ok && role != "" {
			c.Set("user_role", role)
		}
		if roles, ok := claims["roles"].([]any); ok {
			out := make([]string, 0, len(roles))
			for _, r := range roles {
				if s, ok := r.(string); ok {
					out = append(out, s)
				}
			}
			if len(out) > 0 {
				c.Set("roles", out)
			}
		}
		// Attach raw claims for downstream tenant extraction
		c.Set("jwt_claims", claims)
		c.Set("auth_method", "jwt")
		c.Next()
	}
}

// APIKeyAuth validates API keys
func APIKeyAuth(validKeys []string) gin.HandlerFunc {
	keyMap := make(map[string]bool)
	for _, key := range validKeys {
		keyMap[key] = true
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			c.Abort()
			return
		}

		if !keyMap[apiKey] {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Mock service information
		c.Set("user_id", "service-"+apiKey[:8])
		c.Set("user_role", "service")
		c.Set("auth_method", "api_key")
		c.Next()
	}
}

// CombinedAuth allows either JWT or API key authentication
func CombinedAuth(jwtAuth, apiKeyAuth gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Authorization header exists (JWT)
		if c.GetHeader("Authorization") != "" {
			jwtAuth(c)
			return
		}

		// Check if API key header exists
		if c.GetHeader("X-API-Key") != "" {
			apiKeyAuth(c)
			return
		}

		// No authentication provided
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Provide either JWT token (Authorization: Bearer <token>) or API key (X-API-Key: <key>)",
		})
		c.Abort()
	}
}

// RBAC implements role-based access control
func RBAC(allowedRoles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool)
	for _, role := range allowedRoles {
		roleMap[role] = true
	}

	return func(c *gin.Context) {
		// Prefer explicit single role (non-enterprise path)
		if userRoleVal, exists := c.Get("user_role"); exists {
			if roleStr, ok := userRoleVal.(string); ok {
				if roleMap[roleStr] {
					c.Next()
					return
				}
				c.JSON(http.StatusForbidden, gin.H{
					"error":          "Insufficient permissions",
					"required_roles": allowedRoles,
					"user_role":      roleStr,
				})
				c.Abort()
				return
			}
		}

		// Fallback: roles slice (enterprise structured auth)
		if rolesVal, exists := c.Get("roles"); exists {
			if roles, ok := rolesVal.([]string); ok {
				for _, r := range roles {
					if roleMap[r] {
						// Optionally set user_role for downstream consumers
						c.Set("user_role", r)
						c.Next()
						return
					}
				}
				c.JSON(http.StatusForbidden, gin.H{
					"error":          "Insufficient permissions",
					"required_roles": allowedRoles,
					"user_roles":     roles,
				})
				c.Abort()
				return
			}
		}

		// No role information present
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User role not found",
		})
		c.Abort()
	}
}

// TenantContext resolves tenant from headers and attaches to context when multi-tenancy is enabled.
// Header: X-Tenant-ID. When enabled and header is present/non-empty, sets:
//   - Gin context key "tenant_id"
//   - Request context value multitenant.ContextKeyTenantID
//
// If feature flag is disabled, this middleware is a no-op.
// TenantContext: Intentional test helper kept separate from production buildTenantMiddleware.
// Used only in unit/e2e tests to validate tenant extraction chain (header -> jwt_claims -> context propagation).
// Production server enforces tenant requirements via buildTenantMiddleware in enterprise.go.
//
//nolint:deadcode // referenced from tests only
func TenantContext(cfg *config.ConsolidatedConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !multitenant.IsEnabled(cfg) {
			c.Next()
			return
		}
		tenant := c.GetHeader("X-Tenant-ID")
		// If header is empty, try to derive from JWT claims previously stored by enterprise auth
		if strings.TrimSpace(tenant) == "" {
			if claimsVal, exists := c.Get("jwt_claims"); exists && claimsVal != nil {
				// jwt_claims is of type *JWTClaims in enterprise build; use reflection to avoid import cycle
				// and keep this middleware decoupled from the auth implementation.
				// We look for an exported field named "TenantID" or a map key "tenant_id".
				switch cv := claimsVal.(type) {
				case interface{ GetTenantID() string }:
					// In case a getter exists in future, prefer it.
					if t := strings.TrimSpace(cv.GetTenantID()); t != "" {
						tenant = t
					}
				default:
					// Use reflect to read struct field if available
					// (safe since we only read exported field; ignore panics)
					// Fallback to map[string]any with "tenant_id" if present.
					// Note: keep lightweight; avoid heavy reflection helpers.
					// Struct path
					//nolint:all // reflective read on best-effort basis
					func() {
						defer func() { _ = recover() }()
						rv := reflect.ValueOf(claimsVal)
						if rv.Kind() == reflect.Ptr {
							rv = rv.Elem()
						}
						if rv.IsValid() && rv.Kind() == reflect.Struct {
							f := rv.FieldByName("TenantID")
							if f.IsValid() && f.Kind() == reflect.String {
								if t := strings.TrimSpace(f.String()); t != "" {
									tenant = t
									return
								}
							}
						}
						// Map path
						if m, ok := claimsVal.(map[string]any); ok {
							if v, ok2 := m["tenant_id"]; ok2 {
								if s, ok3 := v.(string); ok3 && strings.TrimSpace(s) != "" {
									tenant = s
								}
							}
						}
					}()
				}
			}
		}
		if strings.TrimSpace(tenant) != "" {
			// Set for handlers using Gin context directly
			c.Set("tenant_id", tenant)
			// And propagate into request context for libraries using context
			ctx := multitenant.WithTenantToContext(c.Request.Context(), tenant)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
