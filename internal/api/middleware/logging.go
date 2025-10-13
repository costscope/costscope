package middleware

import (
	"time"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// RequestLogging logs structured request/response with correlation IDs.
func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get base logger with context IDs
		logger := logging.FromContext(c.Request.Context())
		fields := map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"ip":     c.ClientIP(),
		}
		if fields["path"] == nil || fields["path"] == "" {
			fields["path"] = c.Request.URL.Path
		}
		logger.InfoWithFields("http_request_start", fields)

		c.Next()

		dur := time.Since(start)
		logger.InfoWithFields("http_request_end", map[string]interface{}{
			"status":   c.Writer.Status(),
			"duration": dur.String(),
			"bytes":    c.Writer.Size(),
		})
	}
}
