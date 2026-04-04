package middleware

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrMissingAuthHeader  = errors.New("missing authorization header")
	ErrInvalidAuthScheme  = errors.New("invalid authorization scheme")
	ErrInvalidSigningKey  = errors.New("invalid signing key")
)

// Claims represents JWT claims structure
type Claims struct {
	UserID   string            `json:"user_id"`
	Email    string            `json:"email"`
	Role     string            `json:"role"`
	Scopes   []string          `json:"scopes"`
	Service  string            `json:"service,omitempty"` // For service-to-service auth
	jwt.RegisteredClaims
}

// JWTConfig holds JWT middleware configuration
type JWTConfig struct {
	SecretKey          string
	ExpirationHours   int
	SigningAlgorithm   string
	TokenLookup        string // Default: "header:Authorization"
	AuthScheme         string // Default: "Bearer"
	SkipPaths          []string
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:        getEnv("JWT_SECRET_KEY", ""),
		ExpirationHours:  24,
		SigningAlgorithm: "HS256",
		TokenLookup:      "header:Authorization",
		AuthScheme:       "Bearer",
		SkipPaths: []string{
			"/health",
			"/ready",
			"/docs",
			"/swagger/index.html",
			"/swagger/",
			"/openapi.json",
		},
	}
}

// JWTAuth returns JWT authentication middleware
func JWTAuth(cfg JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for paths in skip list
		if shouldSkipPath(c.Request.URL.Path, cfg.SkipPaths) {
			c.Next()
			return
		}

		// Check if auth is disabled (empty secret key = auth off)
		if cfg.SecretKey == "" {
			slog.Warn("JWT_SECRET_KEY not set — authentication is DISABLED. Do not use in production.")
			c.Next()
			return
		}

		token, err := extractToken(c, cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "UNAUTHORIZED",
			})
			return
		}

		claims, err := validateToken(token, cfg.SecretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "TOKEN_INVALID",
			})
			return
		}

		// Set claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("scopes", claims.Scopes)
		c.Set("service", claims.Service)
		c.Set("claims", claims)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(c *gin.Context, cfg JWTConfig) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], cfg.AuthScheme) {
		return "", ErrInvalidAuthScheme
	}

	return parts[1], nil
}

// validateToken validates JWT token and returns claims
func validateToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// shouldSkipPath checks if path should skip authentication
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// GenerateToken generates a new JWT token
func GenerateToken(claims *Claims, secretKey string, expirationHours int) (string, error) {
	now := time.Now()
	if claims.RegisteredClaims.NotBefore == nil {
		claims.RegisteredClaims.NotBefore = jwt.NewNumericDate(now)
	}

	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(expirationHours) * time.Hour))
	claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	claims.RegisteredClaims.Issuer = "vnstock-api"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// HasScope checks if claims contain a specific scope
func HasScope(c *gin.Context, scope string) bool {
	scopes, exists := c.Get("scopes")
	if !exists {
		return false
	}

	scopeList, ok := scopes.([]string)
	if !ok {
		return false
	}

	for _, s := range scopeList {
		if subtle.ConstantTimeCompare([]byte(s), []byte(scope)) == 1 {
			return true
		}
	}
	return false
}

// RequireScope returns a middleware that requires a specific scope
func RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasScope(c, scope) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
				"code":  "FORBIDDEN",
				"required_scope": scope,
			})
			return
		}
		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if s, ok := userID.(string); ok {
			return s
		}
	}
	return ""
}

// GetClaims retrieves full claims from context
func GetClaims(c *gin.Context) *Claims {
	if claims, exists := c.Get("claims"); exists {
		if v, ok := claims.(*Claims); ok {
			return v
		}
	}
	return nil
}

// JWTOrAPIKeyAuth returns a middleware that accepts EITHER a valid JWT token
// OR a valid API key. This is the standard pattern for APIs that support both
// programmatic (API key) and user-facing (JWT) access.
func JWTOrAPIKeyAuth(jwtCfg JWTConfig, apiKeyCfg APIKeyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for paths in skip list
		if shouldSkipPath(c.Request.URL.Path, jwtCfg.SkipPaths) ||
			shouldSkipPath(c.Request.URL.Path, apiKeyCfg.SkipPaths) {
			c.Next()
			return
		}

		// Check if auth is effectively disabled
		if jwtCfg.SecretKey == "" && len(apiKeyCfg.APIKeys) == 0 {
			slog.Warn("Both JWT_SECRET_KEY and API_KEYS are empty — authentication is DISABLED")
			c.Next()
			return
		}

		// Try JWT first (if Authorization header is present)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && jwtCfg.SecretKey != "" {
			token, err := extractToken(c, jwtCfg)
			if err == nil {
				claims, err := validateToken(token, jwtCfg.SecretKey)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("email", claims.Email)
					c.Set("role", claims.Role)
					c.Set("scopes", claims.Scopes)
					c.Set("service", claims.Service)
					c.Set("claims", claims)
					c.Next()
					return
				}
			}
		}

		// Try API key (if header is present)
		apiKey := c.GetHeader(apiKeyCfg.HeaderName)
		if apiKey != "" && len(apiKeyCfg.APIKeys) > 0 {
			serviceName, valid := apiKeyCfg.APIKeys[apiKey]
			if valid {
				c.Set("service", serviceName)
				c.Set("api_key", apiKey)
				c.Next()
				return
			}
		}

		// Neither auth method succeeded
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "valid JWT token or API key required",
			"code":  "UNAUTHORIZED",
		})
	}
}
