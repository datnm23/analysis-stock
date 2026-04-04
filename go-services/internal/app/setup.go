package app

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vnstock-hybrid/internal/config"
	"vnstock-hybrid/internal/database"
	"vnstock-hybrid/internal/telemetry"
)

// Infrastructure holds shared database and cache connections used by all services.
type Infrastructure struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Cfg    *config.Config
	otelFn func(context.Context) error // OTel shutdown function
}

// Setup initializes shared infrastructure (config, database, Redis, telemetry).
// serviceName is used to identify the binary in tracing (e.g. "api-gateway").
func Setup(ctx context.Context, serviceName string) *Infrastructure {
	cfg := config.Load()

	infra := &Infrastructure{Cfg: cfg}

	// Initialize OpenTelemetry tracing (non-fatal)
	cfg.Telemetry.ServiceName = serviceName
	if cfg.Telemetry.Enabled {
		shutdown, err := telemetry.InitTracer(ctx, serviceName, cfg.Telemetry.Endpoint, cfg.Telemetry.SampleRate)
		if err != nil {
			log.Printf("Warning: OTel init failed (tracing disabled): %v", err)
		} else {
			infra.otelFn = shutdown
			log.Printf("OTel tracing enabled → %s", cfg.Telemetry.Endpoint)
		}
	}

	// Database connection (optional)
	if cfg.Database.Password != "" {
		db, err := database.NewPostgresDB(cfg.Database)
		if err != nil {
			log.Printf("Warning: Database not available: %v", err)
		} else {
			infra.DB = db
		}
	}

	// Redis connection (optional)
	if cfg.Redis.Host != "" {
		rdb, err := database.NewRedisClient(cfg.Redis)
		if err != nil {
			log.Printf("Warning: Redis not available: %v", err)
		} else {
			infra.Redis = rdb
		}
	}

	return infra
}

// Close releases all infrastructure resources.
func (i *Infrastructure) Close(ctx context.Context) {
	if i.otelFn != nil {
		if err := i.otelFn(ctx); err != nil {
			log.Printf("Warning: OTel shutdown error: %v", err)
		}
	}
	if i.Redis != nil {
		i.Redis.Close()
	}
	if i.DB != nil {
		if sqlDB, err := i.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}
