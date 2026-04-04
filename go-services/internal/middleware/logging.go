package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a custom structured logging middleware using slog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		slog.Info("HTTP request",
			"status", status,
			"latency", latency.String(),
			"method", c.Request.Method,
			"path", path,
			"client_ip", c.ClientIP(),
		)
	}
}
