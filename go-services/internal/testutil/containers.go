//go:build integration

package testutil

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"vnstock-hybrid/internal/models"
)

// SetupPostgres starts a PostgreSQL test container and returns a connected *gorm.DB.
// Call the returned cleanup function to stop the container.
func SetupPostgres(ctx context.Context) (*gorm.DB, func()) {
	container, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("vnstock_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to get postgres connection string: %v", err)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test postgres: %v", err)
	}

	// Run migrations
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Warning: failed to terminate postgres container: %v", err)
		}
	}

	return db, cleanup
}

// SetupRedis starts a Redis test container and returns a connected *redis.Client.
// Call the returned cleanup function to stop the container.
func SetupRedis(ctx context.Context) (*redis.Client, func()) {
	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		log.Fatalf("Failed to start redis container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatalf("Failed to get redis port: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Verify connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping test redis: %v", err)
	}

	cleanup := func() {
		rdb.Close()
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Warning: failed to terminate redis container: %v", err)
		}
	}

	return rdb, cleanup
}
