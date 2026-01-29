package handlers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"vnstock-hybrid/internal/services"
)

var symbolPattern = regexp.MustCompile(`^[A-Z]{3}$`)

// TechnicalAnalysis handles single symbol technical analysis
func TechnicalAnalysis(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")

		if !symbolPattern.MatchString(symbol) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid symbol format, expected 3 uppercase letters",
			})
			return
		}

		result, err := svc.Analyze(c.Request.Context(), symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// TechnicalBatchRequest represents batch analysis request
type TechnicalBatchRequest struct {
	Symbols []string `json:"symbols" binding:"required,min=1,max=50"`
}

// TechnicalBatch handles batch technical analysis
func TechnicalBatch(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TechnicalBatchRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Validate symbols
		for _, symbol := range req.Symbols {
			if !symbolPattern.MatchString(symbol) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "invalid symbol format",
					"symbol": symbol,
				})
				return
			}
		}

		results, err := svc.AnalyzeBatch(c.Request.Context(), req.Symbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"count":   len(results),
		})
	}
}

// SentimentProxy proxies requests to the Python sentiment service
func SentimentProxy(client *services.SentimentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Texts []services.TextItem `json:"texts" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		result, err := client.Analyze(c.Request.Context(), req.Texts)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "sentiment service unavailable: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// FullAnalysisRequest represents a full analysis request
type FullAnalysisRequest struct {
	Symbols          []string `json:"symbols" binding:"required,min=1,max=50"`
	IncludeSentiment bool     `json:"include_sentiment"`
	IncludeForecast  bool     `json:"include_forecast"`
}

// FullAnalysis performs combined technical and sentiment analysis
func FullAnalysis(techSvc *services.TechnicalService, sentClient *services.SentimentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req FullAnalysisRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx := c.Request.Context()

		// Get technical analysis
		techResults, err := techSvc.AnalyzeBatch(ctx, req.Symbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Build response
		results := make(map[string]interface{})
		for symbol, tech := range techResults {
			results[symbol] = gin.H{
				"symbol":    symbol,
				"technical": tech,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"request_id": c.GetString("request_id"),
			"results":    results,
			"count":      len(results),
		})
	}
}
