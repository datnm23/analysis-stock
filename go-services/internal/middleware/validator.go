package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidationConfig holds validation middleware configuration
type ValidationConfig struct {
	// Stock symbol validation
	SymbolPattern *regexp.Regexp
	SymbolMinLen  int
	SymbolMaxLen  int

	// Text validation
	MinTextLength int
	MaxTextLength int

	// Batch size limits
	MaxBatchSize int
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() ValidationConfig {
	// Vietnamese stock symbols: 3 uppercase letters (HSX, HNX) or more (UPCOM)
	return ValidationConfig{
		SymbolPattern:   regexp.MustCompile("^[A-Z0-9]{1,10}$"),
		SymbolMinLen:    1,
		SymbolMaxLen:    10,
		MinTextLength:   1,
		MaxTextLength:   10000,
		MaxBatchSize:    100,
	}
}

// ValidateInput returns input validation middleware
func ValidateInput(cfg ValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate stock symbol in path
		symbol := c.Param("symbol")
		if symbol != "" && !cfg.SymbolPattern.MatchString(symbol) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid stock symbol format",
				"code":  "INVALID_SYMBOL",
				"hint":  "Symbol must be 1-10 uppercase letters or numbers",
			})
			return
		}

		// Validate request body for POST/PUT
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if err := validateRequestBody(c, cfg); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
					"code": "INVALID_REQUEST",
				})
				return
			}
		}

		c.Next()
	}
}

// validateRequestBody validates the request body based on content
func validateRequestBody(c *gin.Context, cfg ValidationConfig) error {
	// Skip for now - detailed validation would require parsing JSON
	// This is a basic check that can be enhanced

	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// Basic size check
		if c.Request.ContentLength > 0 {
			maxSize := int64(cfg.MaxTextLength * 10) // Approximate for JSON
			if c.Request.ContentLength > maxSize {
				return &ValidationError{
					Message: "request body too large",
					Code:    "PAYLOAD_TOO_LARGE",
				}
			}
		}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateSymbol validates a stock symbol
func ValidateSymbol(symbol string) error {
	if symbol == "" {
		return &ValidationError{
			Message: "symbol is required",
			Code:    "MISSING_SYMBOL",
		}
	}

	cfg := DefaultValidationConfig()
	if len(symbol) < cfg.SymbolMinLen || len(symbol) > cfg.SymbolMaxLen {
		return &ValidationError{
			Message: "symbol must be 1-10 characters",
			Code:    "INVALID_SYMBOL_LENGTH",
		}
	}

	if !cfg.SymbolPattern.MatchString(symbol) {
		return &ValidationError{
			Message: "symbol must contain only uppercase letters and numbers",
			Code:    "INVALID_SYMBOL_FORMAT",
		}
	}

	return nil
}

// ValidateBatchSize validates batch request size
func ValidateBatchSize(size int, maxSize int) error {
	if size <= 0 {
		return &ValidationError{
			Message: "batch size must be greater than 0",
			Code:    "INVALID_BATCH_SIZE",
		}
	}

	if size > maxSize {
		return &ValidationError{
			Message: "batch size exceeds maximum allowed (" + strconv.Itoa(maxSize) + ")",
			Code:    "BATCH_SIZE_EXCEEDED",
		}
	}

	return nil
}

// symbolSanitizePattern is pre-compiled for performance.
var symbolSanitizePattern = regexp.MustCompile("[^A-Z0-9]")

// SanitizeSymbol sanitizes a stock symbol
func SanitizeSymbol(symbol string) string {
	// Remove whitespace and convert to uppercase
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	// Remove any non-alphanumeric characters
	return symbolSanitizePattern.ReplaceAllString(symbol, "")
}
