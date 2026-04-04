package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"vnstock-hybrid/internal/services"
)

// ForecastAnalysis handles single symbol forecast
// @Summary		Get forecast for a symbol
// @Description	Returns combined technical + sentiment forecast
// @Tags		forecast
// @Accept		json
// @Produce		json
// @Param		symbol	path		string	true	"Stock symbol (e.g., VNM)"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router		/api/v1/forecast/{symbol} [get]
func ForecastAnalysis(svc *services.ForecastService) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")

		if !symbolPattern.MatchString(symbol) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid symbol format, expected 3 uppercase letters",
			})
			return
		}

		result, err := svc.Forecast(c.Request.Context(), symbol)
		if err != nil {
			slog.Error("Forecast failed", "symbol", symbol, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "forecast analysis failed",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// MarketData handles market data endpoint for n8n compatibility
// @Summary		Get market data for a symbol
// @Description	Returns raw market data for a stock symbol
// @Tags		market
// @Accept		json
// @Produce		json
// @Param		request	body		map[string]interface{}	true	"Market data request"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Router		/api/market/data [post]
func MarketData(svc *services.TechnicalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbol string `json:"symbol" binding:"required"`
			Days   int    `json:"days"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !symbolPattern.MatchString(req.Symbol) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid symbol format",
			})
			return
		}

		result, err := svc.Analyze(c.Request.Context(), req.Symbol)
		if err != nil {
			slog.Error("Market data fetch failed", "symbol", req.Symbol, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "market data unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"symbol":      result.Symbol,
			"market_data": result.Price,
			"timestamp":   result.Timestamp,
		})
	}
}

// DailyReport handles daily report retrieval
// @Summary		Get daily report
// @Description	Returns the latest daily analysis report
// @Tags		reports
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		404	{object}	map[string]string
// @Router		/api/reports/daily [get]
func DailyReport(orchSvc *services.OrchestratorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		report, err := orchSvc.GetLatestReport(c.Request.Context())
		if err != nil {
			slog.Warn("Daily report not found", "error", err)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no daily report available",
			})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// GenerateDailyReport triggers daily report generation
// @Summary		Generate daily report
// @Description	Triggers analysis for all watchlist symbols
// @Tags		reports
// @Accept		json
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		500	{object}	map[string]string
// @Router		/api/reports/generate [post]
func GenerateDailyReport(orchSvc *services.OrchestratorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbols []string `json:"symbols"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		report, err := orchSvc.GenerateDailyReport(c.Request.Context(), req.Symbols)
		if err != nil {
			slog.Error("Daily report generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "report generation failed"})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// QueueAnalysisRequest is the request body for async analysis
type QueueAnalysisRequest struct {
	Symbols []string `json:"symbols" binding:"required,min=1,max=50"`
}

// EnqueueAnalysis accepts a batch analysis job and returns immediately.
// Clients poll GET /api/v1/analysis/status/:id for the result.
// @Summary		Queue async batch analysis
// @Description	Enqueues symbols for background analysis via Redis Streams
// @Tags		analysis
// @Accept		json
// @Produce		json
// @Param		request	body		QueueAnalysisRequest	true	"Symbols to analyze"
// @Success		202		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]string
// @Failure		503		{object}	map[string]string
// @Router		/api/v1/analysis/queue [post]
func EnqueueAnalysis(q *services.JobQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueueAnalysisRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		for _, sym := range req.Symbols {
			if !symbolPattern.MatchString(sym) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol", "symbol": sym})
				return
			}
		}

		requestID := c.GetString("request_id") // set by CorrelationID middleware
		if requestID == "" {
			requestID = fmt.Sprintf("job-%d", time.Now().UnixNano())
		}

		if err := q.Enqueue(c.Request.Context(), requestID, req.Symbols); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "queue unavailable: " + err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"request_id": requestID,
			"status":     "queued",
			"symbols":    req.Symbols,
			"poll_url":   "/api/v1/analysis/status/" + requestID,
		})
	}
}

// AnalysisStatus returns the current state of an async analysis job.
// @Summary		Poll async analysis status
// @Description	Returns queued/processing/done/failed status and results when ready
// @Tags		analysis
// @Produce		json
// @Param		id	path	string	true	"Request ID returned by the queue endpoint"
// @Success		200	{object}	map[string]interface{}
// @Failure		404	{object}	map[string]string
// @Router		/api/v1/analysis/status/{id} [get]
func AnalysisStatus(q *services.JobQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, err := q.GetStatus(c.Request.Context(), id)
		if err != nil {
			slog.Error("Job status lookup failed", "request_id", id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "status lookup failed"})
			return
		}
		if result == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found", "request_id": id})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// SynthesizeHandler handles the /api/synthesize endpoint for n8n compatibility
func SynthesizeHandler(forecastSvc *services.ForecastService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Symbol string `json:"symbol" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !symbolPattern.MatchString(req.Symbol) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol format"})
			return
		}

		result, err := forecastSvc.Forecast(c.Request.Context(), req.Symbol)
		if err != nil {
			slog.Error("Synthesize failed", "symbol", req.Symbol, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "synthesis failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
