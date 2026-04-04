package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// incrWithExpire atomically increments a counter and sets TTL on first call.
// Uses a Lua script to prevent the INCR/EXPIRE race condition.
var incrWithExpire = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

// RateLimiter creates a rate limiting middleware using Redis
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	windowSecs := int(window.Seconds())

	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		// Atomically increment counter and set TTL (Lua script prevents race)
		result, err := incrWithExpire.Run(ctx, rdb, []string{key}, windowSecs).Int64()
		if err != nil {
			slog.Error("Rate limiter Redis error, allowing request", "error", err, "ip", clientIP)
			c.Next()
			return
		}

		count := result

		// Add rate limit headers on all responses
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		remaining := int64(limit) - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		// Check limit
		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
