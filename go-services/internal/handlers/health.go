package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/services"
)

// HealthCheck returns a basic health check handler
// @Summary		Health check
// @Description	Returns the health status of the API gateway
// @Tags		health
// @Produce		json
// @Success		200	{object}	map[string]string
// @Router		/health [get]
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
				if status == "healthy" {
					status = "degraded"
				}
			}
		}

		httpStatus := http.StatusOK
		if status == "unhealthy" {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":  status,
			"service": "vnstock-api-gateway",
		})
	}
}

// ReadinessCheck checks if the service is ready to accept requests
// @Summary		Readiness check
// @Description	Returns the readiness status and component health
// @Tags		health
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Router		/ready [get]
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
