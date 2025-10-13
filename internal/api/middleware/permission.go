package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/core/security"

	"github.com/gin-gonic/gin"
)

const auditModeEnv = "COSTSCOPE_RBAC_AUDIT_MODE"

var auditModeEnabled atomic.Value // bool

func init() {
	// Initialize audit mode flag once; tests can override via SetAuditModeForTests
	auditModeEnabled.Store(isAuditEnvEnabled())
}

func isAuditEnvEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(auditModeEnv)))
	switch v {
	case "1", "true", "t", "yes", "y", "on", "audit", "enabled":
		return true
	}
	return false
}

// SetAuditModeForTests is an intentional test hook allowing deterministic toggling of audit soft-deny
// semantics without process restart (env variable is read only once in init). Not used by production code.
//
//nolint:deadcode // accessed from multiple *_test.go files
func SetAuditModeForTests(enabled bool) { auditModeEnabled.Store(enabled) }

// NOTE: we intentionally avoid defining an interface until a second implementation (e.g. Casbin) is live; premature
// abstraction can hide compile-time breakages. Keep direct dependency on *security.RBACService for now.

// RequirePermission enforces (resource, action) against a user_role present in context (set by auth middleware) using RBACService.CheckPermission.
// It allows a comma separated list of fallback roles in header X-User-Roles if user_role not set.
// Tracing: RBACService.CheckPermission creates the rbac.has_permission span.
func RequirePermission(rbac *security.RBACService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := ""
		if v, ok := c.Get("user_role"); ok {
			if rs, ok2 := v.(string); ok2 {
				role = rs
			}
		}
		// Optional multi-role header fallback
		if role == "" {
			rolesHeader := c.GetHeader("X-User-Roles")
			if rolesHeader != "" {
				roles := strings.Split(rolesHeader, ",")
				for _, r := range roles {
					r = strings.TrimSpace(r)
					if r == "" {
						continue
					}
					if rbac.CheckPermission(c.Request.Context(), r, resource, action) {
						c.Set("user_role", r)
						c.Next()
						return
					}
				}
			}
		} else {
			if rbac.CheckPermission(c.Request.Context(), role, resource, action) {
				c.Next()
				return
			}
		}
		// Denied path
		if auditModeEnabled.Load().(bool) {
			// Audit mode: proceed but annotate response for downstream awareness
			c.Writer.Header().Set("X-RBAC-Audit", "deny")
			telemetry.RBACAuditSoftDenies.WithLabelValues(resource, action).Inc()
			c.Next()
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied", "resource": resource, "action": action})
		c.Abort()
	}
}
