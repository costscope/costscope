package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testRemoteAddr = "127.0.0.1:12345"

const (
	testJWTSecret = "test-secret"
	testJWTIssuer = "test-issuer"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())

	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SecurityHeaders())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	// X-XSS-Protection is deprecated in modern browsers and intentionally omitted
	assert.Empty(t, w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set very low rate limit for testing
	router.Use(RateLimit(2, time.Minute))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = testRemoteAddr
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should succeed
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = testRemoteAddr
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request should be rate limited
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = testRemoteAddr
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuth(testJWTSecret, testJWTIssuer))

	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "No Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Authorization header required")
			},
		},
		{
			name:           "Invalid Authorization Format",
			authHeader:     "Invalid token",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid authorization header format")
			},
		},
		{
			name:           "Empty Token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid")
			},
		},
		{
			name:           "Valid Token",
			authHeader:     "Bearer " + mustSignedToken(t, "user-123"),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "user-123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validKeys := []string{"key1", "key2", "test-api-key"}
	router := gin.New()
	router.Use(APIKeyAuth(validKeys))

	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "No API Key Header",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "API key required")
			},
		},
		{
			name:           "Invalid API Key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid API key")
			},
		},
		{
			name:           "Valid API Key",
			apiKey:         "test-api-key",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "service-test-api")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestCombinedAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtAuth := JWTAuth(testJWTSecret, testJWTIssuer)
	apiKeyAuth := APIKeyAuth([]string{"test-api-key"})
	router.Use(CombinedAuth(jwtAuth, apiKeyAuth))

	router.GET("/protected", func(c *gin.Context) {
		authMethod, exists := c.Get("auth_method")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth_method not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"auth_method": authMethod})
	})

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "No Authentication",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Authentication required")
			},
		},
		{
			name: "JWT Authentication",
			headers: map[string]string{
				"Authorization": "Bearer " + mustSignedToken(t, "user-abc"),
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "jwt")
			},
		},
		{
			name: "API Key Authentication",
			headers: map[string]string{
				"X-API-Key": "test-api-key",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "api_key")
			},
		},
		{
			name: "JWT Takes Precedence",
			headers: map[string]string{
				"Authorization": "Bearer " + mustSignedToken(t, "user-abc"),
				"X-API-Key":     "test-api-key",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "jwt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			for header, value := range tt.headers {
				req.Header.Set(header, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up auth middleware that sets user role
	router.Use(func(c *gin.Context) {
		role := c.GetHeader("X-User-Role")
		if role != "" {
			c.Set("user_role", role)
		}
		c.Next()
	})

	// Apply RBAC middleware
	router.Use(RBAC("admin", "manager"))

	router.GET("/admin-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	tests := []struct {
		name           string
		userRole       string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "No Role Set",
			userRole:       "",
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "User role not found")
			},
		},
		{
			name:           "Insufficient Permissions",
			userRole:       "user",
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Insufficient permissions")
			},
		},
		{
			name:           "Admin Role Allowed",
			userRole:       "admin",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "admin access granted")
			},
		},
		{
			name:           "Manager Role Allowed",
			userRole:       "manager",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "admin access granted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/admin-only", nil)
			if tt.userRole != "" {
				req.Header.Set("X-User-Role", tt.userRole)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Chain multiple middleware
	router.Use(
		RequestID(),
		SecurityHeaders(),
		JWTAuth(testJWTSecret, testJWTIssuer),
		RBAC("admin"),
	)

	router.GET("/secure", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "secure endpoint accessed"})
	})

	// Test successful access with all middleware
	req, _ := http.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+mustSignedToken(t, "user-xyz"))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Contains(t, w.Body.String(), "secure endpoint accessed")
}

// mustSignedToken creates a signed HS256 JWT for tests and fails t on error.
func mustSignedToken(t *testing.T, sub string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss":  testJWTIssuer,
		"sub":  sub,
		"role": "admin",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestRateLimitWithDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimit(2, time.Hour)) // Longer window, higher limit

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Test first IP
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Test second IP
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.2")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
