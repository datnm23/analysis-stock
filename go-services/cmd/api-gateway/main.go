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
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/config"
	"vnstock-hybrid/internal/database"
	"vnstock-hybrid/internal/handlers"
	"vnstock-hybrid/internal/middleware"
	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/pkg/vnstock"
)

func main() {
	cfg := config.Load()

	// Database connection (optional - skip if not configured)
	var db *gorm.DB
	if cfg.Database.Password != "" {
		var err error
		db, err = database.NewPostgresDB(cfg.Database)
		if err != nil {
			log.Printf("Warning: Database not available: %v", err)
		}
	}

	// Redis connection (optional - skip if not configured)
	var rdb *redis.Client
	if cfg.Redis.Host != "" {
		var err error
		rdb, err = database.NewRedisClient(cfg.Redis)
		if err != nil {
			log.Printf("Warning: Redis not available: %v", err)
		}
	}

	// Initialize services
	marketClient := vnstock.NewClient()
	technicalSvc := services.NewTechnicalService(db, rdb, marketClient)
	sentimentClient := services.NewSentimentClient(cfg.Services.SentimentURL)

	// Setup Gin
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	if rdb != nil {
		r.Use(middleware.RateLimiter(rdb, 100, time.Minute))
	}

	// Health endpoints
	r.GET("/health", handlers.HealthCheck(db, rdb))
	r.GET("/ready", handlers.ReadinessCheck(db, rdb, sentimentClient))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Technical analysis
		v1.GET("/technical/:symbol", handlers.TechnicalAnalysis(technicalSvc))
		v1.POST("/technical/batch", handlers.TechnicalBatch(technicalSvc))

		// Sentiment (proxy to Python)
		v1.POST("/sentiment", handlers.SentimentProxy(sentimentClient))

		// Combined analysis
		v1.POST("/analyze", handlers.FullAnalysis(technicalSvc, sentimentClient))
	}

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("API Gateway starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
