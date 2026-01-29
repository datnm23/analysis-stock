package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a custom logging middleware
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

		// Log format: status | latency | method | path
		gin.DefaultWriter.Write([]byte(
			c.ClientIP() + " | " +
				time.Now().Format("2006/01/02 - 15:04:05") + " | " +
				string(rune(status)) + " | " +
				latency.String() + " | " +
				c.Request.Method + " | " +
				path + "\n",
		))
	}
}
