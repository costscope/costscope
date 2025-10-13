//go:build enterprise
// +build enterprise

package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// =====================================================================================
// Authentication & Authorization Middleware - Enterprise Security
// =====================================================================================

// JWTClaims represents JWT token claims
type JWTClaims struct {
	// Optional tenant claim for multi-tenant scoping. When present, middleware may use
	// it to derive the effective tenant (alongside X-Tenant-ID header).
	TenantID  string   `json:"tenant_id,omitempty"`
	UserID    string   `json:"user_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	Scopes    []string `json:"scopes"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	logger    *logging.Logger
	jwtSecret string
	issuer    string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(logger *logging.Logger, jwtSecret, issuer string) *AuthMiddleware {
	return &AuthMiddleware{
		logger:    logger,
		jwtSecret: jwtSecret,
		issuer:    issuer,
	}
}

// RequireAuth middleware requires valid JWT authentication
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			telemetry.AuthFailures.WithLabelValues("missing_header").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			telemetry.AuthFailures.WithLabelValues("bad_format").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := a.validateToken(tokenString)
		if err != nil {
			// Classify common error substrings
			reason := "validation_error"
			if strings.Contains(err.Error(), "expired") {
				reason = "expired"
			} else if strings.Contains(err.Error(), "issuer") {
				reason = "issuer"
			}
			telemetry.AuthFailures.WithLabelValues(reason).Inc()
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("scopes", claims.Scopes)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RequireRole middleware requires specific role
func (a *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			telemetry.AuthFailures.WithLabelValues("forbidden_role").Inc()
			c.JSON(http.StatusForbidden, gin.H{"error": "No roles found in token"})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			telemetry.AuthFailures.WithLabelValues("forbidden_role").Inc()
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid roles format"})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range userRoles {
			if role == requiredRole || role == "admin" { // Admin can access everything
				hasRole = true
				break
			}
		}

		if !hasRole {
			telemetry.AuthFailures.WithLabelValues("forbidden_role").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Required role '%s' not found", requiredRole),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireScope middleware requires specific scope
func (a *AuthMiddleware) RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("scopes")
		if !exists {
			telemetry.AuthFailures.WithLabelValues("forbidden_scope").Inc()
			c.JSON(http.StatusForbidden, gin.H{"error": "No scopes found in token"})
			c.Abort()
			return
		}

		userScopes, ok := scopes.([]string)
		if !ok {
			telemetry.AuthFailures.WithLabelValues("forbidden_scope").Inc()
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid scopes format"})
			c.Abort()
			return
		}

		// Check if user has required scope
		hasScope := false
		for _, scope := range userScopes {
			if scope == requiredScope || scope == "admin:all" { // Admin scope
				hasScope = true
				break
			}
		}

		if !hasScope {
			telemetry.AuthFailures.WithLabelValues("forbidden_scope").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Required scope '%s' not found", requiredScope),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware adds user info if token is present but doesn't require it
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		// Try to validate token, but don't fail if invalid
		claims, err := a.validateToken(tokenString)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("roles", claims.Roles)
			c.Set("scopes", claims.Scopes)
			c.Set("jwt_claims", claims)
			c.Set("authenticated", true)
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// validateToken validates JWT token and returns claims
func (a *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	// Check issuer
	if claims.Issuer != a.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for testing/demo purposes
func (a *AuthMiddleware) GenerateToken(userID, username, email string, roles, scopes []string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Roles:     roles,
		Scopes:    scopes,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(duration).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtSecret))
}

// =====================================================================================
// CORS Middleware
// =====================================================================================

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "X-API-Key", "X-Request-ID", "X-Tenant-ID",
		},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowOrigin := false
		for _, allowedOrigin := range config.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowOrigin = true
				break
			}
		}

		if allowOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%.0f", config.MaxAge.Seconds()))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// =====================================================================================
// Rate Limiting Middleware
// =====================================================================================

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	logger   *logging.Logger
	requests map[string][]time.Time
	mutex    sync.RWMutex
	maxReqs  int
	window   time.Duration
	keyFunc  func(*gin.Context) string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(logger *logging.Logger, maxReqs int, window time.Duration, keyFunc func(*gin.Context) string) *RateLimiter {
	if keyFunc == nil {
		keyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return &RateLimiter{
		logger:   logger,
		requests: make(map[string][]time.Time),
		maxReqs:  maxReqs,
		window:   window,
		keyFunc:  keyFunc,
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.keyFunc(c)
		now := time.Now()

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		// Get or create request history for this key
		requests, exists := rl.requests[key]
		if !exists {
			requests = []time.Time{}
		}

		// Remove old requests outside the window
		cutoff := now.Add(-rl.window)
		validRequests := []time.Time{}
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if rate limit exceeded
		if len(validRequests) >= rl.maxReqs {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxReqs))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(rl.window).Unix()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		// Add current request
		validRequests = append(validRequests, now)
		rl.requests[key] = validRequests

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxReqs))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.maxReqs-len(validRequests)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(rl.window).Unix()))

		c.Next()
	}
}
