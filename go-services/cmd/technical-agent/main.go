package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"vnstock-hybrid/internal/app"
	"vnstock-hybrid/internal/handlers"
	"vnstock-hybrid/internal/middleware"
	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/pkg/vnstock"
)

func main() {
	infra := app.Setup(context.Background(), "technical-agent")
	defer infra.Close(context.Background())

	cfg := infra.Cfg

	// Initialize services
	marketClient := vnstock.NewClient()
	technicalSvc := services.NewTechnicalService(infra.DB, infra.Redis, marketClient)

	// Setup Gin
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware(cfg.Telemetry.ServiceName))
	r.Use(middleware.CorrelationID())
	r.Use(middleware.PrometheusMetrics())
	r.Use(middleware.Logger())

	// Health & metrics endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "technical-agent",
		})
	})
	r.GET("/metrics", middleware.MetricsHandler())

	// Technical analysis endpoints
	r.GET("/analyze/:symbol", handlers.TechnicalAnalysis(technicalSvc))
	r.POST("/analyze/batch", handlers.TechnicalBatch(technicalSvc))

	// Internal API for other services
	r.GET("/internal/indicators/:symbol", func(c *gin.Context) {
		symbol := c.Param("symbol")
		result, err := technicalSvc.Analyze(c.Request.Context(), symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Technical Agent starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Technical Agent...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Technical Agent exited")
}
