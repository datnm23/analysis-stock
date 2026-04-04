package handlers

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"vnstock-hybrid/internal/services"
)

// @title			VN Stock API
// @version		1.0
// @description	Vietnamese Stock Market Analysis API
// @termsOfService	http://swagger.io/terms/

// @contact.name	API Support
// @contact.url		https://github.com/datnm
// @contact.email	support@vnstock.local

// @license.name	Apache 2.0
// @license.url		http://www.apache.org/licenses/LICENSE-2.0.html

// @host		localhost:8080
// @BasePath	/

var symbolPattern = regexp.MustCompile(`^[A-Z0-9]{1,10}$`)

// TechnicalAnalysis handles single symbol technical analysis
// @Summary		Get technical analysis for a symbol
// @Description	Returns technical indicators and signals for a single stock symbol
// @Tags		technical
// @Accept		json
// @Produce		json
// @Param		symbol	path		string	true	"Stock symbol (e.g., VNM)"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/api/v1/technical/{symbol} [get]
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
			slog.Error("Technical analysis failed", "symbol", symbol, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "technical analysis failed",
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
// @Summary		Batch technical analysis
// @Description	Returns technical indicators for multiple stock symbols
// @Tags		technical
// @Accept		json
// @Produce		json
// @Param		request	body		TechnicalBatchRequest	true	"Batch request"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/api/v1/technical/batch [post]
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
			slog.Error("Batch analysis failed", "symbols", req.Symbols, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "batch analysis failed",
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
// @Summary		Sentiment analysis
// @Description	Analyze sentiment of Vietnamese stock news
// @Tags		sentiment
// @Accept		json
// @Produce		json
// @Param		request	body		map[string]interface{}	true	"Sentiment request"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Failure		503		{object}	map[string]string
// @Router		/api/v1/sentiment [post]
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

// LegacyTechnicalAnalysis handles the n8n-compatible POST /api/analyze/technical endpoint.
// It accepts {"symbol": "VNM"} in the request body instead of a URL path parameter.
func LegacyTechnicalAnalysis(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbol string `json:"symbol" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if !symbolPattern.MatchString(req.Symbol) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol format, expected 3 uppercase letters"})
			return
		}

		result, err := svc.Analyze(c.Request.Context(), req.Symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "analysis failed"})
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
// @Summary		Full stock analysis
// @Description	Perform combined technical and sentiment analysis on multiple symbols
// @Tags		analysis
// @Accept		json
// @Produce		json
// @Param		request	body		FullAnalysisRequest	true	"Analysis request"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/api/v1/analyze [post]
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
			slog.Error("Full analysis failed", "symbols", req.Symbols, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "analysis failed",
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
