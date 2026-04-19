package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	PubSub    PubSubConfig
	Services  ServicesConfig
	Auth      AuthConfig
	Telemetry TelemetryConfig
}

// TelemetryConfig holds OpenTelemetry tracing configuration.
type TelemetryConfig struct {
	Enabled     bool
	Endpoint    string  // host:port of OTLP collector, e.g. "jaeger:4318"
	ServiceName string  // set per-binary in main.go
	SampleRate  float64 // 0.0–1.0; 1.0 = trace every request
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecretKey       string
	JWTExpirationHours int
	APIKeyHeader      string
	APIKeys           string // Comma-separated key:value pairs
	EnableAuth        bool
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	// TLS configuration
	EnableTLS    bool
	CertFile     string
	KeyFile      string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	AutoMigrate     bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type PubSubConfig struct {
	ProjectID string
}

type ServicesConfig struct {
	TechnicalURL   string
	SentimentURL   string
	ForecastURL    string
	VnStockBaseURL string
	MarketURL      string // crawl-agent price board service
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			EnableTLS:    getEnv("ENABLE_TLS", "false") == "true",
			CertFile:     getEnv("TLS_CERT_FILE", "certs/server.crt"),
			KeyFile:      getEnv("TLS_KEY_FILE", "certs/server.key"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "vnstock"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "vnstock"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			AutoMigrate:     getEnv("DB_AUTO_MIGRATE", "true") == "true",
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		PubSub: PubSubConfig{
			ProjectID: getEnv("GCP_PROJECT_ID", ""),
		},
		Services: ServicesConfig{
			TechnicalURL:   getEnv("TECHNICAL_SERVICE_URL", "http://localhost:8081"),
			SentimentURL:   getEnv("SENTIMENT_SERVICE_URL", "http://localhost:8000"),
			ForecastURL:    getEnv("FORECAST_SERVICE_URL", "http://localhost:8082"),
			VnStockBaseURL: getEnv("VNSTOCK_BASE_URL", "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks"),
			MarketURL:      getEnv("MARKET_SERVICE_URL", "http://localhost:8085"),
		},
		Auth: AuthConfig{
			JWTSecretKey:       getEnv("JWT_SECRET_KEY", ""),
			JWTExpirationHours: getIntEnv("JWT_EXPIRATION_HOURS", 24),
			APIKeyHeader:      getEnv("API_KEY_HEADER", "X-API-Key"),
			APIKeys:           getEnv("API_KEYS", ""),
			EnableAuth:        getEnv("ENABLE_AUTH", "false") == "true",
		},
		Telemetry: TelemetryConfig{
			Enabled:    getEnv("OTEL_ENABLED", "false") == "true",
			Endpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "jaeger:4318"),
			SampleRate: getFloatEnv("OTEL_SAMPLE_RATE", 1.0),
			// ServiceName is intentionally left empty here; each binary sets it.
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		slog.Warn("Invalid integer env var, using default", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		slog.Warn("Invalid duration env var, using default", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
		slog.Warn("Invalid float env var, using default", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}
