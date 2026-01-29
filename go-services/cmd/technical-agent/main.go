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
	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/pkg/vnstock"
)

func main() {
	cfg := config.Load()

	// Database connection
	var db *gorm.DB
	if cfg.Database.Password != "" {
		var err error
		db, err = database.NewPostgresDB(cfg.Database)
		if err != nil {
			log.Printf("Warning: Database not available: %v", err)
		}
	}

	// Redis connection
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

	// Setup Gin
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Health endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "technical-agent",
		})
	})

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
