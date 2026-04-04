package database

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"vnstock-hybrid/internal/config"
	"vnstock-hybrid/internal/models"
)

// NewPostgresDB creates a new PostgreSQL database connection with connection pooling.
func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	logLevel := logger.Warn // default: only warnings and errors
	switch strings.ToLower(os.Getenv("DB_LOG_LEVEL")) {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "info":
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	slog.Info("Database pool configured",
		"max_open", cfg.MaxOpenConns,
		"max_idle", cfg.MaxIdleConns,
		"max_lifetime", cfg.ConnMaxLifetime,
	)

	// Auto migrate models (dev/staging only; use golang-migrate in production)
	if cfg.AutoMigrate {
		if err := models.AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return db, nil
}
