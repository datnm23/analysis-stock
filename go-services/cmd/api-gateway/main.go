package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"vnstock-hybrid/internal/app"
	"vnstock-hybrid/internal/handlers"
	"vnstock-hybrid/internal/middleware"
	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/pkg/vnstock"
)

func main() {
	infra := app.Setup(context.Background(), "api-gateway")
	defer infra.Close(context.Background())

	cfg := infra.Cfg

	// Setup JWT configuration
	jwtCfg := middleware.DefaultJWTConfig()
	jwtCfg.SecretKey = cfg.Auth.JWTSecretKey
	jwtCfg.ExpirationHours = cfg.Auth.JWTExpirationHours
	if cfg.Auth.EnableAuth && cfg.Auth.JWTSecretKey != "" {
		jwtCfg.SkipPaths = []string{"/health", "/ready", "/metrics"}
	}

	// Setup API key configuration
	apiKeyCfg := middleware.DefaultAPIKeyConfig()
	apiKeyCfg.HeaderName = cfg.Auth.APIKeyHeader
	if cfg.Auth.APIKeys != "" {
		pairs := strings.Split(cfg.Auth.APIKeys, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				apiKeyCfg.APIKeys[parts[0]] = parts[1]
			}
		}
	}

	// Initialize services
	marketClient := vnstock.NewClientWithConfig(cfg.Services.VnStockBaseURL, 30*time.Second)
	technicalSvc := services.NewTechnicalService(infra.DB, infra.Redis, marketClient)
	sentimentClient := services.NewSentimentClient(cfg.Services.SentimentURL)
	forecastSvc := services.NewForecastService(infra.DB, infra.Redis, technicalSvc, sentimentClient)
	orchestratorSvc := services.NewOrchestratorService(infra.DB, infra.Redis, forecastSvc)
	articleSvc := services.NewArticleService(infra.DB)
	internalKey := os.Getenv("INTERNAL_API_KEY")

	// Async job queue (Redis Streams)
	var jobQueue *services.JobQueue
	if infra.Redis != nil {
		var err error
		jobQueue, err = services.NewJobQueue(infra.Redis, forecastSvc)
		if err != nil {
			log.Printf("Warning: job queue unavailable: %v", err)
		}
	}

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

	if infra.Redis != nil {
		r.Use(middleware.RateLimiter(infra.Redis, 100, time.Minute))
	}

	if cfg.Auth.EnableAuth {
		r.Use(middleware.JWTOrAPIKeyAuth(jwtCfg, apiKeyCfg))
	}

	// Health & metrics endpoints (no auth)
	r.GET("/health", handlers.HealthCheck(infra.DB, infra.Redis))
	r.GET("/ready", handlers.ReadinessCheck(infra.DB, infra.Redis, sentimentClient))
	r.GET("/metrics", middleware.MetricsHandler())

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.Group("/api/v1")
	{
		v1.GET("/technical/:symbol", handlers.TechnicalAnalysis(technicalSvc))
		v1.POST("/technical/batch", handlers.TechnicalBatch(technicalSvc))
		v1.POST("/sentiment", handlers.SentimentProxy(sentimentClient))
		v1.POST("/analyze", handlers.FullAnalysis(technicalSvc, sentimentClient))
		v1.GET("/forecast/:symbol", handlers.ForecastAnalysis(forecastSvc))
		v1.GET("/reports/daily", handlers.DailyReport(orchestratorSvc))
		v1.POST("/reports/generate", handlers.GenerateDailyReport(orchestratorSvc))

		if jobQueue != nil {
			v1.POST("/analysis/queue", handlers.EnqueueAnalysis(jobQueue))
			v1.GET("/analysis/status/:id", handlers.AnalysisStatus(jobQueue))
		}

		// Chart OHLCV data
		v1.GET("/chart/:symbol", handlers.ChartData())

		// Articles (blog)
		v1.GET("/articles", handlers.ListArticles(articleSvc))
		v1.GET("/articles/:slug", handlers.GetArticleBySlug(articleSvc))
		v1.POST("/articles", handlers.CreateArticle(articleSvc, internalKey))
		v1.PATCH("/articles/:id/status", handlers.UpdateArticleStatus(articleSvc, internalKey))
	}

	// Legacy routes (n8n compatibility)
	legacy := r.Group("/api")
	{
		legacy.POST("/analyze/technical", handlers.LegacyTechnicalAnalysis(technicalSvc))
		legacy.POST("/analyze/sentiment", handlers.SentimentProxy(sentimentClient))
		legacy.POST("/synthesize", handlers.SynthesizeHandler(forecastSvc))
		legacy.POST("/market/data", handlers.MarketData(technicalSvc))
		legacy.GET("/reports/daily", handlers.DailyReport(orchestratorSvc))
	}

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	if jobQueue != nil {
		go jobQueue.StartWorker(workerCtx)
	}

	go func() {
		log.Printf("API Gateway starting on port %s", cfg.Server.Port)
		var err error
		if cfg.Server.EnableTLS {
			log.Printf("TLS enabled - using cert: %s, key: %s", cfg.Server.CertFile, cfg.Server.KeyFile)
			err = srv.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
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
