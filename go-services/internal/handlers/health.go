package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/services"
)

// HealthCheck returns a basic health check handler
func HealthCheck(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "healthy"

		// Check database
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				status = "unhealthy"
			}
		}

		// Check Redis
		if rdb != nil {
			if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
				status = "degraded"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  status,
			"service": "vnstock-api-gateway",
		})
	}
}

// ReadinessCheck checks if the service is ready to accept requests
func ReadinessCheck(db *gorm.DB, rdb *redis.Client, sentimentClient *services.SentimentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ready := true
		checks := make(map[string]string)

		// Check database
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				ready = false
				checks["database"] = "unhealthy"
			} else {
				checks["database"] = "healthy"
			}
		}

		// Check Redis
		if rdb != nil {
			if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
				ready = false
				checks["redis"] = "unhealthy"
			} else {
				checks["redis"] = "healthy"
			}
		}

		// Check sentiment service
		if sentimentClient != nil {
			if err := sentimentClient.Health(c.Request.Context()); err != nil {
				checks["sentiment_service"] = "unhealthy"
				// Don't fail readiness for sentiment service
			} else {
				checks["sentiment_service"] = "healthy"
			}
		}

		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"ready":  ready,
			"checks": checks,
		})
	}
}
