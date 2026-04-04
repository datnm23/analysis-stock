package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyConfig holds API key middleware configuration
type APIKeyConfig struct {
	HeaderName   string
	APIKeys      map[string]string // api_key -> service_name
	SkipPaths    []string
}

// DefaultAPIKeyConfig returns default API key configuration
func DefaultAPIKeyConfig() APIKeyConfig {
	return APIKeyConfig{
		HeaderName: "X-API-Key",
		APIKeys:    make(map[string]string),
		SkipPaths: []string{
			"/health",
			"/ready",
			"/docs",
			"/swagger/",
			"/openapi.json",
		},
	}
}

// APIKeyAuth returns API key authentication middleware
func APIKeyAuth(cfg APIKeyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for paths in skip list
		if shouldSkipPath(c.Request.URL.Path, cfg.SkipPaths) {
			c.Next()
			return
		}

		// Check if auth is disabled (no API keys configured)
		if len(cfg.APIKeys) == 0 {
			c.Next()
			return
		}

		apiKey := c.GetHeader(cfg.HeaderName)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing API key",
				"code":  "MISSING_API_KEY",
				"header": cfg.HeaderName,
			})
			return
		}

		// Validate API key
		serviceName, valid := cfg.APIKeys[apiKey]
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
				"code":  "INVALID_API_KEY",
			})
			return
		}

		// Set service name in context
		c.Set("service", serviceName)
		c.Set("api_key", apiKey)

		c.Next()
	}
}

// APIKeyFromEnv creates APIKeyConfig from environment variables
// Format: API_KEYS=key1:service1,key2:service2
func APIKeyFromEnv() APIKeyConfig {
	cfg := DefaultAPIKeyConfig()
	cfg.HeaderName = getEnv("API_KEY_HEADER", "X-API-Key")

	apiKeysEnv := getEnv("API_KEYS", "")
	if apiKeysEnv != "" {
		pairs := strings.Split(apiKeysEnv, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				cfg.APIKeys[parts[0]] = parts[1]
			}
		}
	}

	return cfg
}

// GetServiceName retrieves service name from context
func GetServiceName(c *gin.Context) string {
	if service, exists := c.Get("service"); exists {
		return service.(string)
	}
	return ""
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
