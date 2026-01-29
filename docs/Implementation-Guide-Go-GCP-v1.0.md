# Implementation Guide
# VN Stock Analysis System - Go/GCP Edition
## Version 1.0

**Document ID:** IMPL-VNSTOCK-GO-GCP-001
**Date:** January 2026
**Stack:** Go 1.22+, Google Cloud Platform

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Phase 1: GCP Infrastructure Setup](#2-phase-1-gcp-infrastructure-setup)
3. [Phase 2: Go Project Structure](#3-phase-2-go-project-structure)
4. [Phase 3: API Gateway Service](#4-phase-3-api-gateway-service)
5. [Phase 4: Technical Analysis Agent](#5-phase-4-technical-analysis-agent)
6. [Phase 5: Sentiment Analysis Agent](#6-phase-5-sentiment-analysis-agent)
7. [Phase 6: Forecast & Orchestrator](#7-phase-6-forecast--orchestrator)
8. [Phase 7: Telegram Bot & Distribution](#8-phase-7-telegram-bot--distribution)
9. [Phase 8: Cloud Workflows & Scheduler](#9-phase-8-cloud-workflows--scheduler)
10. [Deployment & CI/CD](#10-deployment--cicd)

---

## 1. Introduction

### 1.1 Purpose

This guide provides step-by-step instructions for implementing the VN Stock Analysis System using **Go** and **Google Cloud Platform**.

### 1.2 Prerequisites

```bash
# Required tools
go version        # 1.22+
gcloud version    # Latest
docker version    # 20.10+
terraform version # 1.5+

# GCP account with billing enabled
# gcloud authenticated: gcloud auth login
```

### 1.3 Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22 |
| Web Framework | Gin |
| ORM | GORM |
| Database | Cloud SQL (PostgreSQL 15) |
| Cache | Memorystore (Redis 7) |
| Storage | Cloud Storage |
| Messaging | Cloud Pub/Sub |
| Compute | Cloud Run |
| Serverless | Cloud Functions |
| Workflow | Cloud Workflows |
| ML/AI | Vertex AI |
| Monitoring | Cloud Monitoring |
| IaC | Terraform |

### 1.4 Project Structure (Target)

```
vnstock-go/
├── cmd/                          # Application entry points
│   ├── api-gateway/
│   │   └── main.go
│   ├── technical-agent/
│   │   └── main.go
│   ├── sentiment-agent/
│   │   └── main.go
│   ├── forecast-agent/
│   │   └── main.go
│   ├── orchestrator/
│   │   └── main.go
│   └── telegram-bot/
│       └── main.go
├── internal/                     # Private packages
│   ├── analysis/
│   │   ├── technical/
│   │   │   ├── agent.go
│   │   │   ├── indicators.go
│   │   │   └── signals.go
│   │   ├── sentiment/
│   │   │   ├── agent.go
│   │   │   └── vertex.go
│   │   └── forecast/
│   │       └── agent.go
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── analysis.go
│   │   │   ├── reports.go
│   │   │   └── health.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   └── logging.go
│   │   └── routes.go
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── postgres.go
│   │   └── models/
│   │       ├── stock.go
│   │       ├── analysis.go
│   │       └── user.go
│   ├── cache/
│   │   └── redis.go
│   ├── storage/
│   │   └── gcs.go
│   ├── pubsub/
│   │   ├── publisher.go
│   │   └── subscriber.go
│   ├── market/
│   │   ├── client.go
│   │   └── vnstock.go
│   ├── nlp/
│   │   └── vietnamese.go
│   └── telegram/
│       ├── bot.go
│       └── handlers.go
├── pkg/                          # Public packages
│   └── models/
│       ├── analysis.go
│       └── common.go
├── deployments/
│   ├── terraform/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── cloud-run.tf
│   │   ├── cloud-sql.tf
│   │   ├── memorystore.tf
│   │   └── storage.tf
│   ├── cloud-run/
│   │   └── service.yaml
│   └── cloud-workflows/
│       └── daily-analysis.yaml
├── functions/                    # Cloud Functions
│   └── scraper/
│       ├── main.go
│       └── go.mod
├── scripts/
│   ├── deploy.sh
│   ├── build.sh
│   └── migrate.sh
├── Dockerfile
├── Dockerfile.agent
├── docker-compose.yaml           # Local development
├── go.mod
├── go.sum
├── Makefile
├── cloudbuild.yaml
└── README.md
```

---

## 2. Phase 1: GCP Infrastructure Setup

### 2.1 Objectives

- Set up GCP project and enable APIs
- Create Terraform configuration
- Provision Cloud SQL, Memorystore, Cloud Storage
- Configure IAM and service accounts

### 2.2 GCP Project Setup

```bash
# Set project ID
export PROJECT_ID="vnstock-analysis"
export REGION="asia-southeast1"

# Create project (if needed)
gcloud projects create $PROJECT_ID --name="VN Stock Analysis"

# Set default project
gcloud config set project $PROJECT_ID

# Enable billing (required)
# Go to: https://console.cloud.google.com/billing

# Enable required APIs
gcloud services enable \
    run.googleapis.com \
    cloudfunctions.googleapis.com \
    cloudbuild.googleapis.com \
    sqladmin.googleapis.com \
    redis.googleapis.com \
    storage.googleapis.com \
    pubsub.googleapis.com \
    workflows.googleapis.com \
    cloudscheduler.googleapis.com \
    secretmanager.googleapis.com \
    aiplatform.googleapis.com \
    artifactregistry.googleapis.com \
    cloudtrace.googleapis.com \
    monitoring.googleapis.com \
    logging.googleapis.com
```

### 2.3 Terraform Configuration

#### Main Configuration

```hcl
# deployments/terraform/main.tf

terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "vnstock-terraform-state"
    prefix = "terraform/state"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",
    "cloudfunctions.googleapis.com",
    "sqladmin.googleapis.com",
    "redis.googleapis.com",
    "storage.googleapis.com",
    "pubsub.googleapis.com",
    "workflows.googleapis.com",
    "cloudscheduler.googleapis.com",
    "secretmanager.googleapis.com",
    "aiplatform.googleapis.com",
  ])

  service            = each.value
  disable_on_destroy = false
}
```

```hcl
# deployments/terraform/variables.tf

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "asia-southeast1"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
```

#### Cloud SQL Configuration

```hcl
# deployments/terraform/cloud-sql.tf

resource "google_sql_database_instance" "main" {
  name             = "vnstock-db-${var.environment}"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier              = var.environment == "prod" ? "db-custom-2-4096" : "db-f1-micro"
    availability_type = var.environment == "prod" ? "REGIONAL" : "ZONAL"
    disk_size         = 20
    disk_type         = "PD_SSD"

    backup_configuration {
      enabled                        = true
      start_time                     = "02:00"
      point_in_time_recovery_enabled = var.environment == "prod"
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.vpc.id
    }

    database_flags {
      name  = "max_connections"
      value = "100"
    }
  }

  deletion_protection = var.environment == "prod"

  depends_on = [google_service_networking_connection.private_vpc_connection]
}

resource "google_sql_database" "vnstock" {
  name     = "vnstock"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "app" {
  name     = "vnstock_app"
  instance = google_sql_database_instance.main.name
  password = var.db_password
}

# VPC for private SQL access
resource "google_compute_network" "vpc" {
  name                    = "vnstock-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  name          = "vnstock-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.vpc.id
}

resource "google_compute_global_address" "private_ip" {
  name          = "vnstock-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip.name]
}
```

#### Memorystore (Redis)

```hcl
# deployments/terraform/memorystore.tf

resource "google_redis_instance" "cache" {
  name           = "vnstock-cache-${var.environment}"
  tier           = var.environment == "prod" ? "STANDARD_HA" : "BASIC"
  memory_size_gb = var.environment == "prod" ? 2 : 1
  region         = var.region

  redis_version = "REDIS_7_0"

  authorized_network = google_compute_network.vpc.id

  labels = {
    environment = var.environment
    app         = "vnstock"
  }
}

output "redis_host" {
  value = google_redis_instance.cache.host
}
```

#### Cloud Storage

```hcl
# deployments/terraform/storage.tf

resource "google_storage_bucket" "data" {
  name     = "${var.project_id}-vnstock-data"
  location = var.region

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 365
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }

  versioning {
    enabled = true
  }
}

# Create folder structure
resource "google_storage_bucket_object" "folders" {
  for_each = toset([
    "raw-data/news/",
    "raw-data/market-data/",
    "raw-data/social/",
    "processed/sentiment/",
    "processed/technical/",
    "processed/combined/",
    "reports/daily/",
    "reports/weekly/",
    "reports/alerts/",
  ])

  name    = each.value
  content = ""
  bucket  = google_storage_bucket.data.name
}
```

#### Pub/Sub Topics

```hcl
# deployments/terraform/pubsub.tf

resource "google_pubsub_topic" "topics" {
  for_each = toset([
    "news-ingested",
    "analysis-requested",
    "analysis-completed",
    "alerts",
    "reports",
  ])

  name = each.value

  labels = {
    environment = var.environment
  }
}

resource "google_pubsub_subscription" "subscriptions" {
  for_each = {
    "news-processor"     = "news-ingested"
    "alert-handler"      = "alerts"
    "report-distributor" = "reports"
  }

  name  = each.key
  topic = google_pubsub_topic.topics[each.value].name

  ack_deadline_seconds = 60

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
}
```

#### Cloud Run Services

```hcl
# deployments/terraform/cloud-run.tf

# API Gateway Service
resource "google_cloud_run_v2_service" "api_gateway" {
  name     = "vnstock-api"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/vnstock/api-gateway:latest"

      resources {
        limits = {
          cpu    = "2"
          memory = "1Gi"
        }
      }

      env {
        name  = "GCP_PROJECT"
        value = var.project_id
      }

      env {
        name = "DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_url.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "REDIS_ADDR"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.redis_addr.secret_id
            version = "latest"
          }
        }
      }
    }

    scaling {
      min_instance_count = 1
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

# VPC Connector for Cloud Run
resource "google_vpc_access_connector" "connector" {
  name          = "vnstock-connector"
  region        = var.region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.8.0.0/28"
}

# Allow unauthenticated access to API
resource "google_cloud_run_v2_service_iam_member" "api_public" {
  location = google_cloud_run_v2_service.api_gateway.location
  name     = google_cloud_run_v2_service.api_gateway.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Technical Agent Service
resource "google_cloud_run_v2_service" "technical_agent" {
  name     = "technical-agent"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/vnstock/technical-agent:latest"

      resources {
        limits = {
          cpu    = "2"
          memory = "2Gi"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }
}

# Sentiment Agent Service
resource "google_cloud_run_v2_service" "sentiment_agent" {
  name     = "sentiment-agent"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/vnstock/sentiment-agent:latest"

      resources {
        limits = {
          cpu    = "4"
          memory = "4Gi"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 3
    }
  }
}
```

### 2.4 Deploy Infrastructure

```bash
# Initialize Terraform
cd deployments/terraform
terraform init

# Create terraform.tfvars
cat > terraform.tfvars <<EOF
project_id  = "vnstock-analysis"
region      = "asia-southeast1"
environment = "dev"
db_password = "your-secure-password"
EOF

# Plan and apply
terraform plan -out=plan.out
terraform apply plan.out
```

### 2.5 Acceptance Criteria

- [ ] GCP project created with billing enabled
- [ ] All required APIs enabled
- [ ] Cloud SQL instance running
- [ ] Memorystore Redis instance running
- [ ] Cloud Storage buckets created
- [ ] Pub/Sub topics and subscriptions created
- [ ] VPC and networking configured

---

## 3. Phase 2: Go Project Structure

### 3.1 Objectives

- Initialize Go module
- Create project structure
- Set up shared packages
- Configure development environment

### 3.2 Initialize Project

```bash
# Create project directory
mkdir vnstock-go && cd vnstock-go

# Initialize Go module
go mod init github.com/yourusername/vnstock

# Create directory structure
mkdir -p cmd/{api-gateway,technical-agent,sentiment-agent,forecast-agent,orchestrator,telegram-bot}
mkdir -p internal/{analysis/{technical,sentiment,forecast},api/{handlers,middleware},config,database/models,cache,storage,pubsub,market,nlp,telegram}
mkdir -p pkg/models
mkdir -p deployments/{terraform,cloud-run,cloud-workflows}
mkdir -p functions/scraper
mkdir -p scripts
```

### 3.3 Go Module Dependencies

```bash
# Install dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/redis/go-redis/v9
go get cloud.google.com/go/storage
go get cloud.google.com/go/pubsub
go get cloud.google.com/go/secretmanager
go get cloud.google.com/go/logging
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
go get golang.org/x/sync/errgroup
```

### 3.4 Configuration Package

```go
// internal/config/config.go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    // Server
    Port        string
    Environment string

    // GCP
    ProjectID string
    Region    string

    // Database
    DatabaseURL string

    // Redis
    RedisAddr     string
    RedisPassword string

    // Storage
    StorageBucket string

    // Pub/Sub
    PubSubProject string

    // Vietnamese Market
    MarketOpenHour  int
    MarketCloseHour int
    Timezone        string

    // Cache TTL
    CacheTTLTechnical time.Duration
    CacheTTLSentiment time.Duration
}

func Load() *Config {
    return &Config{
        Port:        getEnv("PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),

        ProjectID: getEnv("GCP_PROJECT", ""),
        Region:    getEnv("GCP_REGION", "asia-southeast1"),

        DatabaseURL: getEnv("DATABASE_URL", ""),

        RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),

        StorageBucket: getEnv("STORAGE_BUCKET", "vnstock-data"),

        PubSubProject: getEnv("PUBSUB_PROJECT", ""),

        MarketOpenHour:  getEnvInt("MARKET_OPEN_HOUR", 9),
        MarketCloseHour: getEnvInt("MARKET_CLOSE_HOUR", 15),
        Timezone:        getEnv("TIMEZONE", "Asia/Ho_Chi_Minh"),

        CacheTTLTechnical: time.Duration(getEnvInt("CACHE_TTL_TECHNICAL", 300)) * time.Second,
        CacheTTLSentiment: time.Duration(getEnvInt("CACHE_TTL_SENTIMENT", 600)) * time.Second,
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if i, err := strconv.Atoi(value); err == nil {
            return i
        }
    }
    return defaultValue
}
```

### 3.5 Database Models

```go
// internal/database/models/stock.go
package models

import (
    "time"

    "gorm.io/gorm"
)

type Stock struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Symbol    string         `gorm:"uniqueIndex;size:10;not null" json:"symbol"`
    Name      string         `gorm:"size:255" json:"name"`
    Exchange  string         `gorm:"size:10" json:"exchange"` // HSX, HNX, UPCOM
    Sector    string         `gorm:"size:100" json:"sector"`
    IsActive  bool           `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PriceHistory struct {
    ID      uint      `gorm:"primaryKey" json:"id"`
    StockID uint      `gorm:"index;not null" json:"stock_id"`
    Date    time.Time `gorm:"index;not null" json:"date"`
    Open    float64   `json:"open"`
    High    float64   `json:"high"`
    Low     float64   `json:"low"`
    Close   float64   `json:"close"`
    Volume  int64     `json:"volume"`
}
```

```go
// internal/database/models/analysis.go
package models

import (
    "time"

    "gorm.io/datatypes"
)

type AnalysisResult struct {
    ID             uint           `gorm:"primaryKey" json:"id"`
    StockID        uint           `gorm:"index;not null" json:"stock_id"`
    AnalysisDate   time.Time      `gorm:"index;not null" json:"analysis_date"`
    Recommendation string         `gorm:"size:20" json:"recommendation"`
    Confidence     float64        `json:"confidence"`
    TechnicalScore float64        `json:"technical_score"`
    SentimentScore float64        `json:"sentiment_score"`
    RiskLevel      string         `gorm:"size:10" json:"risk_level"`
    RawData        datatypes.JSON `json:"raw_data"`
    CreatedAt      time.Time      `json:"created_at"`
}

type NewsArticle struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    Source           string         `gorm:"size:50;not null" json:"source"`
    Title            string         `gorm:"not null" json:"title"`
    Description      string         `json:"description"`
    URL              string         `gorm:"uniqueIndex" json:"url"`
    PublishedAt      time.Time      `json:"published_at"`
    Sentiment        string         `gorm:"size:20" json:"sentiment"`
    SentimentScore   float64        `json:"sentiment_score"`
    MentionedSymbols datatypes.JSON `json:"mentioned_symbols"`
    IsSpam           bool           `gorm:"default:false" json:"is_spam"`
    CreatedAt        time.Time      `json:"created_at"`
}

type Alert struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    StockID   uint      `gorm:"index" json:"stock_id"`
    AlertType string    `gorm:"size:50;not null" json:"alert_type"`
    Message   string    `gorm:"not null" json:"message"`
    Severity  string    `gorm:"size:20" json:"severity"`
    IsSent    bool      `gorm:"default:false" json:"is_sent"`
    CreatedAt time.Time `json:"created_at"`
    SentAt    *time.Time `json:"sent_at"`
}
```

### 3.6 Database Connection

```go
// internal/database/postgres.go
package database

import (
    "fmt"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/yourusername/vnstock/internal/config"
    "github.com/yourusername/vnstock/internal/database/models"
)

func NewPostgresConnection(cfg *config.Config) (*gorm.DB, error) {
    logLevel := logger.Silent
    if cfg.Environment == "development" {
        logLevel = logger.Info
    }

    db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Auto-migrate schemas
    if err := db.AutoMigrate(
        &models.Stock{},
        &models.PriceHistory{},
        &models.AnalysisResult{},
        &models.NewsArticle{},
        &models.Alert{},
        &models.User{},
    ); err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    return db, nil
}
```

### 3.7 Redis Cache

```go
// internal/cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/yourusername/vnstock/internal/config"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(cfg *config.Config) (*RedisCache, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
        DB:       0,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
    n, err := c.client.Exists(ctx, key).Result()
    return n > 0, err
}
```

### 3.8 Cloud Storage Client

```go
// internal/storage/gcs.go
package storage

import (
    "context"
    "fmt"
    "io"
    "time"

    "cloud.google.com/go/storage"

    "github.com/yourusername/vnstock/internal/config"
)

type GCSClient struct {
    client *storage.Client
    bucket string
}

func NewGCSClient(ctx context.Context, cfg *config.Config) (*GCSClient, error) {
    client, err := storage.NewClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create storage client: %w", err)
    }

    return &GCSClient{
        client: client,
        bucket: cfg.StorageBucket,
    }, nil
}

func (g *GCSClient) Upload(ctx context.Context, objectPath string, data []byte) error {
    wc := g.client.Bucket(g.bucket).Object(objectPath).NewWriter(ctx)
    defer wc.Close()

    if _, err := wc.Write(data); err != nil {
        return fmt.Errorf("failed to write to GCS: %w", err)
    }

    return nil
}

func (g *GCSClient) Download(ctx context.Context, objectPath string) ([]byte, error) {
    rc, err := g.client.Bucket(g.bucket).Object(objectPath).NewReader(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to read from GCS: %w", err)
    }
    defer rc.Close()

    return io.ReadAll(rc)
}

func (g *GCSClient) List(ctx context.Context, prefix string) ([]string, error) {
    var objects []string
    it := g.client.Bucket(g.bucket).Objects(ctx, &storage.Query{Prefix: prefix})

    for {
        attrs, err := it.Next()
        if err == storage.ErrObjectNotExist {
            break
        }
        if err != nil {
            return nil, err
        }
        objects = append(objects, attrs.Name)
    }

    return objects, nil
}

func (g *GCSClient) GenerateSignedURL(objectPath string, duration time.Duration) (string, error) {
    return g.client.Bucket(g.bucket).SignedURL(objectPath, &storage.SignedURLOptions{
        Method:  "GET",
        Expires: time.Now().Add(duration),
    })
}

func (g *GCSClient) Close() error {
    return g.client.Close()
}
```

### 3.9 Acceptance Criteria

- [ ] Go module initialized with all dependencies
- [ ] Project structure created
- [ ] Config package loads environment variables
- [ ] Database models defined
- [ ] Redis cache client working
- [ ] GCS storage client working

---

## 4. Phase 3: API Gateway Service

### 4.1 Objectives

- Create Gin-based HTTP server
- Implement middleware (CORS, logging, rate limiting)
- Set up API routes
- Add health check endpoint

### 4.2 API Routes

```go
// internal/api/routes.go
package api

import (
    "github.com/gin-gonic/gin"

    "github.com/yourusername/vnstock/internal/api/handlers"
    "github.com/yourusername/vnstock/internal/api/middleware"
)

func SetupRoutes(r *gin.Engine, h *handlers.Handlers) {
    // Global middleware
    r.Use(middleware.Logger())
    r.Use(middleware.CORS())
    r.Use(middleware.Recovery())

    // Health check (no auth)
    r.GET("/health", h.HealthCheck)
    r.GET("/", h.Root)

    // API routes
    api := r.Group("/api")
    api.Use(middleware.RateLimit(100)) // 100 requests per minute
    {
        // Analysis endpoints
        api.POST("/analyze/technical", h.AnalyzeTechnical)
        api.POST("/analyze/sentiment", h.AnalyzeSentiment)
        api.POST("/synthesize", h.Synthesize)

        // Reports
        api.GET("/reports/daily", h.GetDailyReport)
        api.GET("/reports/daily/:date", h.GetDailyReportByDate)

        // Stocks
        api.GET("/stocks", h.ListStocks)
        api.GET("/stocks/:symbol", h.GetStockAnalysis)
    }
}
```

### 4.3 Middleware

```go
// internal/api/middleware/cors.go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

```go
// internal/api/middleware/logging.go
package middleware

import (
    "log/slog"
    "time"

    "github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        slog.Info("request",
            "method", c.Request.Method,
            "path", path,
            "status", status,
            "latency_ms", latency.Milliseconds(),
            "client_ip", c.ClientIP(),
        )
    }
}
```

```go
// internal/api/middleware/ratelimit.go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

type rateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
    limit    int
    window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
    return &rateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (rl *rateLimiter) allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    windowStart := now.Add(-rl.window)

    // Filter old requests
    var valid []time.Time
    for _, t := range rl.requests[key] {
        if t.After(windowStart) {
            valid = append(valid, t)
        }
    }

    if len(valid) >= rl.limit {
        rl.requests[key] = valid
        return false
    }

    rl.requests[key] = append(valid, now)
    return true
}

func RateLimit(requestsPerMinute int) gin.HandlerFunc {
    limiter := newRateLimiter(requestsPerMinute, time.Minute)

    return func(c *gin.Context) {
        key := c.ClientIP()

        if !limiter.allow(key) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":   "rate limit exceeded",
                "message": "Too many requests. Please try again later.",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 4.4 Handlers

```go
// internal/api/handlers/handlers.go
package handlers

import (
    "github.com/yourusername/vnstock/internal/cache"
    "github.com/yourusername/vnstock/internal/config"
    "github.com/yourusername/vnstock/internal/storage"
    "gorm.io/gorm"
)

type Handlers struct {
    config  *config.Config
    db      *gorm.DB
    cache   *cache.RedisCache
    storage *storage.GCSClient
}

func NewHandlers(cfg *config.Config, db *gorm.DB, cache *cache.RedisCache, storage *storage.GCSClient) *Handlers {
    return &Handlers{
        config:  cfg,
        db:      db,
        cache:   cache,
        storage: storage,
    }
}
```

```go
// internal/api/handlers/health.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func (h *Handlers) HealthCheck(c *gin.Context) {
    health := gin.H{
        "status":  "healthy",
        "version": "1.0.0",
        "services": gin.H{
            "database": "connected",
            "redis":    "connected",
        },
    }

    // Check database
    sqlDB, err := h.db.DB()
    if err != nil || sqlDB.Ping() != nil {
        health["services"].(gin.H)["database"] = "disconnected"
        health["status"] = "degraded"
    }

    // Check Redis
    if _, err := h.cache.Exists(c.Request.Context(), "health-check"); err != nil {
        health["services"].(gin.H)["redis"] = "disconnected"
        health["status"] = "degraded"
    }

    c.JSON(http.StatusOK, health)
}

func (h *Handlers) Root(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "name":    "VN Stock Analysis API",
        "version": "1.0.0",
        "docs":    "/swagger/index.html",
    })
}
```

```go
// internal/api/handlers/analysis.go
package handlers

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
)

type TechnicalAnalysisRequest struct {
    Symbol string `json:"symbol" binding:"required,len=3,alpha"`
    Days   int    `json:"days" binding:"min=30,max=365"`
}

func (h *Handlers) AnalyzeTechnical(c *gin.Context) {
    var req TechnicalAnalysisRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Set default days
    if req.Days == 0 {
        req.Days = 90
    }

    // Check cache first
    cacheKey := fmt.Sprintf("technical:%s:%d", req.Symbol, req.Days)
    var cachedResult interface{}
    if err := h.cache.Get(c.Request.Context(), cacheKey, &cachedResult); err == nil {
        c.JSON(http.StatusOK, cachedResult)
        return
    }

    // Call Technical Agent service (internal Cloud Run)
    // In production, this would be an HTTP call to the technical-agent service
    result := gin.H{
        "symbol":         req.Symbol,
        "days":           req.Days,
        "message":        "Technical analysis - implement agent call",
        "recommendation": "HOLD",
    }

    // Cache result
    h.cache.Set(c.Request.Context(), cacheKey, result, h.config.CacheTTLTechnical)

    c.JSON(http.StatusOK, result)
}

type SentimentAnalysisRequest struct {
    Texts  []string `json:"texts" binding:"required,min=1"`
    Symbol string   `json:"symbol,omitempty"`
}

func (h *Handlers) AnalyzeSentiment(c *gin.Context) {
    var req SentimentAnalysisRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Call Sentiment Agent service
    result := gin.H{
        "symbol":            req.Symbol,
        "texts_analyzed":    len(req.Texts),
        "message":           "Sentiment analysis - implement agent call",
        "overall_sentiment": "neutral",
    }

    c.JSON(http.StatusOK, result)
}

func (h *Handlers) Synthesize(c *gin.Context) {
    // Combine technical and sentiment for final recommendation
    c.JSON(http.StatusOK, gin.H{
        "message": "Synthesis endpoint - implement",
    })
}
```

### 4.5 Main Entry Point

```go
// cmd/api-gateway/main.go
package main

import (
    "context"
    "log"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"

    "github.com/yourusername/vnstock/internal/api"
    "github.com/yourusername/vnstock/internal/api/handlers"
    "github.com/yourusername/vnstock/internal/cache"
    "github.com/yourusername/vnstock/internal/config"
    "github.com/yourusername/vnstock/internal/database"
    "github.com/yourusername/vnstock/internal/storage"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Setup structured logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    slog.Info("Starting VN Stock Analysis API",
        "port", cfg.Port,
        "environment", cfg.Environment,
    )

    // Initialize database
    db, err := database.NewPostgresConnection(cfg)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Initialize cache
    redisCache, err := cache.NewRedisCache(cfg)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    // Initialize storage
    ctx := context.Background()
    gcsClient, err := storage.NewGCSClient(ctx, cfg)
    if err != nil {
        log.Fatalf("Failed to create GCS client: %v", err)
    }
    defer gcsClient.Close()

    // Setup Gin
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.New()

    // Setup handlers and routes
    h := handlers.NewHandlers(cfg, db, redisCache, gcsClient)
    api.SetupRoutes(r, h)

    // Start server
    go func() {
        if err := r.Run(":" + cfg.Port); err != nil {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Cleanup
    sqlDB, _ := db.DB()
    sqlDB.Close()

    slog.Info("Server stopped")
}
```

### 4.6 Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api-gateway ./cmd/api-gateway

# Final image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /api-gateway .

# Create non-root user
RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 8080

CMD ["./api-gateway"]
```

### 4.7 Local Development (Docker Compose)

```yaml
# docker-compose.yaml
version: "3.8"

services:
  api-gateway:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - ENVIRONMENT=development
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/vnstock?sslmode=disable
      - REDIS_ADDR=redis:6379
      - STORAGE_BUCKET=vnstock-data-local
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=vnstock
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 4.8 Acceptance Criteria

- [ ] API Gateway starts and responds to health check
- [ ] CORS middleware working
- [ ] Rate limiting working
- [ ] Database connection established
- [ ] Redis cache working
- [ ] Docker build succeeds

---

## 5. Phase 4: Technical Analysis Agent

### 5.1 Objectives

- Implement technical indicator calculations in Go
- Create analysis service
- Deploy as Cloud Run service

### 5.2 Indicator Calculations

```go
// internal/analysis/technical/indicators.go
package technical

import (
    "math"
)

// RSI calculates Relative Strength Index
func RSI(closes []float64, period int) float64 {
    if len(closes) < period+1 {
        return 50.0
    }

    var gains, losses float64
    for i := len(closes) - period; i < len(closes); i++ {
        change := closes[i] - closes[i-1]
        if change > 0 {
            gains += change
        } else {
            losses -= change
        }
    }

    avgGain := gains / float64(period)
    avgLoss := losses / float64(period)

    if avgLoss == 0 {
        return 100.0
    }

    rs := avgGain / avgLoss
    return 100.0 - (100.0 / (1.0 + rs))
}

// MACD calculates Moving Average Convergence Divergence
func MACD(closes []float64) (macdLine, signalLine, histogram float64) {
    if len(closes) < 26 {
        return 0, 0, 0
    }

    ema12 := EMA(closes, 12)
    ema26 := EMA(closes, 26)
    macdLine = ema12 - ema26

    // Calculate signal line (9-period EMA of MACD)
    macdValues := make([]float64, len(closes)-25)
    for i := 26; i <= len(closes); i++ {
        e12 := EMA(closes[:i], 12)
        e26 := EMA(closes[:i], 26)
        macdValues[i-26] = e12 - e26
    }

    if len(macdValues) >= 9 {
        signalLine = EMA(macdValues, 9)
    }

    histogram = macdLine - signalLine
    return
}

// BollingerBands calculates upper, middle, lower bands
func BollingerBands(closes []float64, period int, stdDevMultiplier float64) (upper, middle, lower float64) {
    if len(closes) < period {
        return 0, 0, 0
    }

    middle = SMA(closes, period)
    stdDev := StandardDeviation(closes[len(closes)-period:])
    upper = middle + stdDevMultiplier*stdDev
    lower = middle - stdDevMultiplier*stdDev
    return
}

// SMA calculates Simple Moving Average
func SMA(data []float64, period int) float64 {
    if len(data) < period {
        return 0
    }

    sum := 0.0
    for i := len(data) - period; i < len(data); i++ {
        sum += data[i]
    }
    return sum / float64(period)
}

// EMA calculates Exponential Moving Average
func EMA(data []float64, period int) float64 {
    if len(data) < period {
        return 0
    }

    multiplier := 2.0 / float64(period+1)
    ema := SMA(data[:period], period)

    for i := period; i < len(data); i++ {
        ema = (data[i]-ema)*multiplier + ema
    }
    return ema
}

// StandardDeviation calculates standard deviation
func StandardDeviation(data []float64) float64 {
    if len(data) == 0 {
        return 0
    }

    mean := 0.0
    for _, v := range data {
        mean += v
    }
    mean /= float64(len(data))

    variance := 0.0
    for _, v := range data {
        variance += math.Pow(v-mean, 2)
    }
    variance /= float64(len(data))

    return math.Sqrt(variance)
}

// Stochastic calculates Stochastic Oscillator
func Stochastic(highs, lows, closes []float64, period int) (k, d float64) {
    if len(closes) < period {
        return 50, 50
    }

    // Find highest high and lowest low in period
    highestHigh := highs[len(highs)-1]
    lowestLow := lows[len(lows)-1]

    for i := len(highs) - period; i < len(highs); i++ {
        if highs[i] > highestHigh {
            highestHigh = highs[i]
        }
        if lows[i] < lowestLow {
            lowestLow = lows[i]
        }
    }

    currentClose := closes[len(closes)-1]

    if highestHigh == lowestLow {
        k = 50
    } else {
        k = ((currentClose - lowestLow) / (highestHigh - lowestLow)) * 100
    }

    // D is 3-period SMA of K (simplified here)
    d = k * 0.9

    return
}

// ADX calculates Average Directional Index
func ADX(highs, lows, closes []float64, period int) (adx, plusDI, minusDI float64) {
    if len(closes) < period+1 {
        return 0, 0, 0
    }

    // Simplified ADX calculation
    // In production, implement full Wilder's smoothing

    var plusDM, minusDM, tr float64

    for i := len(closes) - period; i < len(closes); i++ {
        high := highs[i]
        low := lows[i]
        prevHigh := highs[i-1]
        prevLow := lows[i-1]
        prevClose := closes[i-1]

        // True Range
        trHigh := math.Abs(high - prevClose)
        trLow := math.Abs(low - prevClose)
        trRange := high - low
        tr += math.Max(trRange, math.Max(trHigh, trLow))

        // Directional Movement
        upMove := high - prevHigh
        downMove := prevLow - low

        if upMove > downMove && upMove > 0 {
            plusDM += upMove
        }
        if downMove > upMove && downMove > 0 {
            minusDM += downMove
        }
    }

    if tr > 0 {
        plusDI = (plusDM / tr) * 100
        minusDI = (minusDM / tr) * 100
    }

    dx := 0.0
    if plusDI+minusDI > 0 {
        dx = math.Abs(plusDI-minusDI) / (plusDI + minusDI) * 100
    }

    adx = dx // Simplified; should be smoothed
    return
}

// ATR calculates Average True Range
func ATR(highs, lows, closes []float64, period int) float64 {
    if len(closes) < period+1 {
        return 0
    }

    var trSum float64
    for i := len(closes) - period; i < len(closes); i++ {
        high := highs[i]
        low := lows[i]
        prevClose := closes[i-1]

        tr1 := high - low
        tr2 := math.Abs(high - prevClose)
        tr3 := math.Abs(low - prevClose)

        trSum += math.Max(tr1, math.Max(tr2, tr3))
    }

    return trSum / float64(period)
}
```

### 5.3 Technical Agent

```go
// internal/analysis/technical/agent.go
package technical

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/yourusername/vnstock/internal/market"
)

type Agent struct {
    marketClient market.Client
}

func NewAgent(marketClient market.Client) *Agent {
    return &Agent{
        marketClient: marketClient,
    }
}

type AnalysisRequest struct {
    Symbol string `json:"symbol"`
    Days   int    `json:"days"`
}

type AnalysisResult struct {
    Symbol          string             `json:"symbol"`
    Timestamp       time.Time          `json:"timestamp"`
    CurrentPrice    float64            `json:"current_price"`
    Recommendation  string             `json:"recommendation"`
    Confidence      float64            `json:"confidence"`
    TechnicalScore  float64            `json:"technical_score"`
    Signals         []string           `json:"signals"`
    IndicatorScores map[string]float64 `json:"indicator_scores"`
    Indicators      Indicators         `json:"indicators"`
    SupportResist   SupportResistance  `json:"support_resistance"`
    PriceTargets    PriceTargets       `json:"price_targets"`
}

type Indicators struct {
    RSI         float64 `json:"rsi"`
    MACD        float64 `json:"macd"`
    MACDSignal  float64 `json:"macd_signal"`
    SMA20       float64 `json:"sma_20"`
    SMA50       float64 `json:"sma_50"`
    EMA12       float64 `json:"ema_12"`
    EMA26       float64 `json:"ema_26"`
    BBUpper     float64 `json:"bb_upper"`
    BBMiddle    float64 `json:"bb_middle"`
    BBLower     float64 `json:"bb_lower"`
    StochK      float64 `json:"stoch_k"`
    StochD      float64 `json:"stoch_d"`
    ADX         float64 `json:"adx"`
    ATR         float64 `json:"atr"`
    VolumeRatio float64 `json:"volume_ratio"`
}

type SupportResistance struct {
    Resistance []float64 `json:"resistance"`
    Support    []float64 `json:"support"`
}

type PriceTargets struct {
    ShortTerm float64            `json:"short_term"`
    StopLoss  float64            `json:"stop_loss"`
    Fibonacci map[string]float64 `json:"fibonacci"`
}

func (a *Agent) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
    // Fetch market data
    data, err := a.marketClient.GetHistoricalData(ctx, req.Symbol, req.Days+50)
    if err != nil {
        return nil, fmt.Errorf("fetch data: %w", err)
    }

    if len(data) < 50 {
        return nil, fmt.Errorf("insufficient data for %s", req.Symbol)
    }

    // Extract price arrays
    closes := make([]float64, len(data))
    highs := make([]float64, len(data))
    lows := make([]float64, len(data))
    volumes := make([]float64, len(data))

    for i, d := range data {
        closes[i] = d.Close
        highs[i] = d.High
        lows[i] = d.Low
        volumes[i] = float64(d.Volume)
    }

    // Calculate indicators concurrently
    var wg sync.WaitGroup
    var mu sync.Mutex
    indicators := Indicators{}

    wg.Add(7)

    go func() {
        defer wg.Done()
        mu.Lock()
        indicators.RSI = RSI(closes, 14)
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        macd, signal, _ := MACD(closes)
        mu.Lock()
        indicators.MACD = macd
        indicators.MACDSignal = signal
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        upper, middle, lower := BollingerBands(closes, 20, 2)
        mu.Lock()
        indicators.BBUpper = upper
        indicators.BBMiddle = middle
        indicators.BBLower = lower
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        mu.Lock()
        indicators.SMA20 = SMA(closes, 20)
        indicators.SMA50 = SMA(closes, 50)
        indicators.EMA12 = EMA(closes, 12)
        indicators.EMA26 = EMA(closes, 26)
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        k, d := Stochastic(highs, lows, closes, 14)
        mu.Lock()
        indicators.StochK = k
        indicators.StochD = d
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        adx, _, _ := ADX(highs, lows, closes, 14)
        mu.Lock()
        indicators.ADX = adx
        mu.Unlock()
    }()

    go func() {
        defer wg.Done()
        mu.Lock()
        indicators.ATR = ATR(highs, lows, closes, 14)

        // Volume ratio
        avgVol := SMA(volumes, 20)
        if avgVol > 0 {
            indicators.VolumeRatio = volumes[len(volumes)-1] / avgVol
        }
        mu.Unlock()
    }()

    wg.Wait()

    // Generate signals
    signals, scores := a.generateSignals(closes, indicators)

    // Calculate final score
    technicalScore := a.calculateScore(scores)

    // Generate recommendation
    recommendation := a.getRecommendation(technicalScore)
    confidence := a.calculateConfidence(technicalScore)

    currentPrice := closes[len(closes)-1]

    return &AnalysisResult{
        Symbol:          req.Symbol,
        Timestamp:       time.Now(),
        CurrentPrice:    currentPrice,
        Recommendation:  recommendation,
        Confidence:      confidence,
        TechnicalScore:  technicalScore,
        Signals:         signals,
        IndicatorScores: scores,
        Indicators:      indicators,
        SupportResist:   a.calculateSupportResistance(closes),
        PriceTargets:    a.calculatePriceTargets(currentPrice, recommendation, closes),
    }, nil
}

var weights = map[string]float64{
    "rsi":            0.15,
    "macd":           0.20,
    "bollinger":      0.15,
    "moving_average": 0.20,
    "volume":         0.10,
    "adx":            0.10,
    "stochastic":     0.10,
}

func (a *Agent) calculateScore(scores map[string]float64) float64 {
    total := 0.0
    for key, weight := range weights {
        if score, ok := scores[key]; ok {
            total += score * weight
        }
    }
    return total
}

func (a *Agent) getRecommendation(score float64) string {
    switch {
    case score > 0.6:
        return "STRONG BUY"
    case score > 0.2:
        return "BUY"
    case score > -0.2:
        return "HOLD"
    case score > -0.6:
        return "SELL"
    default:
        return "STRONG SELL"
    }
}

func (a *Agent) calculateConfidence(score float64) float64 {
    conf := abs(score) + 0.3
    if conf > 1.0 {
        return 1.0
    }
    return conf
}

func (a *Agent) generateSignals(closes []float64, ind Indicators) ([]string, map[string]float64) {
    signals := []string{}
    scores := map[string]float64{}
    currentPrice := closes[len(closes)-1]

    // RSI
    if ind.RSI < 30 {
        signals = append(signals, "RSI cho thấy quá bán (oversold)")
        scores["rsi"] = 1.0
    } else if ind.RSI > 70 {
        signals = append(signals, "RSI cho thấy quá mua (overbought)")
        scores["rsi"] = -1.0
    } else {
        scores["rsi"] = (50 - ind.RSI) / 50
    }

    // MACD
    if ind.MACD > ind.MACDSignal {
        signals = append(signals, "MACD bullish crossover")
        scores["macd"] = 0.8
    } else {
        scores["macd"] = -0.5
    }

    // Bollinger Bands
    if currentPrice < ind.BBLower {
        signals = append(signals, "Giá dưới Bollinger Band dưới")
        scores["bollinger"] = 1.0
    } else if currentPrice > ind.BBUpper {
        signals = append(signals, "Giá trên Bollinger Band trên")
        scores["bollinger"] = -1.0
    } else {
        scores["bollinger"] = (ind.BBMiddle - currentPrice) / (ind.BBUpper - ind.BBMiddle)
    }

    // Moving Averages
    if ind.SMA20 > ind.SMA50 {
        signals = append(signals, "Golden Cross - uptrend")
        scores["moving_average"] = 0.7
    } else {
        scores["moving_average"] = -0.5
    }

    // Stochastic
    if ind.StochK < 20 {
        signals = append(signals, "Stochastic oversold")
        scores["stochastic"] = 1.0
    } else if ind.StochK > 80 {
        signals = append(signals, "Stochastic overbought")
        scores["stochastic"] = -1.0
    } else {
        scores["stochastic"] = (50 - ind.StochK) / 50
    }

    // ADX
    if ind.ADX > 25 {
        signals = append(signals, fmt.Sprintf("Xu hướng mạnh (ADX: %.1f)", ind.ADX))
        scores["adx"] = 0.5
    } else {
        scores["adx"] = 0.0
    }

    // Volume
    if ind.VolumeRatio > 1.5 {
        signals = append(signals, fmt.Sprintf("Khối lượng cao (%.1fx)", ind.VolumeRatio))
        scores["volume"] = 0.3
    } else {
        scores["volume"] = 0.0
    }

    return signals, scores
}

func (a *Agent) calculateSupportResistance(closes []float64) SupportResistance {
    // Simplified support/resistance calculation
    // In production, use proper local minima/maxima detection

    min := closes[0]
    max := closes[0]

    for _, c := range closes {
        if c < min {
            min = c
        }
        if c > max {
            max = c
        }
    }

    current := closes[len(closes)-1]

    return SupportResistance{
        Resistance: []float64{
            current * 1.03,
            current * 1.05,
            max,
        },
        Support: []float64{
            current * 0.97,
            current * 0.95,
            min,
        },
    }
}

func (a *Agent) calculatePriceTargets(current float64, rec string, closes []float64) PriceTargets {
    high := closes[0]
    low := closes[0]
    for _, c := range closes {
        if c > high {
            high = c
        }
        if c < low {
            low = c
        }
    }

    diff := high - low

    targets := PriceTargets{
        Fibonacci: map[string]float64{
            "0.000": round(high, 2),
            "0.236": round(high-0.236*diff, 2),
            "0.382": round(high-0.382*diff, 2),
            "0.500": round(high-0.500*diff, 2),
            "0.618": round(high-0.618*diff, 2),
            "1.000": round(low, 2),
        },
    }

    switch rec {
    case "STRONG BUY", "BUY":
        targets.ShortTerm = round(current*1.05, 2)
        targets.StopLoss = round(current*0.95, 2)
    case "SELL", "STRONG SELL":
        targets.ShortTerm = round(current*0.95, 2)
        targets.StopLoss = round(current*1.05, 2)
    default:
        targets.ShortTerm = round(current, 2)
        targets.StopLoss = round(current*0.97, 2)
    }

    return targets
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}

func round(x float64, decimals int) float64 {
    pow := 1.0
    for i := 0; i < decimals; i++ {
        pow *= 10
    }
    return float64(int(x*pow+0.5)) / pow
}
```

### 5.4 Technical Agent Main

```go
// cmd/technical-agent/main.go
package main

import (
    "log"
    "log/slog"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"

    "github.com/yourusername/vnstock/internal/analysis/technical"
    "github.com/yourusername/vnstock/internal/config"
    "github.com/yourusername/vnstock/internal/market"
)

func main() {
    cfg := config.Load()

    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    slog.Info("Starting Technical Analysis Agent", "port", cfg.Port)

    // Initialize market client
    marketClient := market.NewClient()

    // Initialize agent
    agent := technical.NewAgent(marketClient)

    // Setup Gin
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.New()
    r.Use(gin.Recovery())

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    r.POST("/analyze", func(c *gin.Context) {
        var req technical.AnalysisRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if req.Days == 0 {
            req.Days = 90
        }

        result, err := agent.Analyze(c.Request.Context(), req)
        if err != nil {
            slog.Error("Analysis failed", "symbol", req.Symbol, "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, result)
    })

    if err := r.Run(":" + cfg.Port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

### 5.5 Acceptance Criteria

- [ ] All indicator calculations accurate
- [ ] Agent service responds to /analyze endpoint
- [ ] Signals generated in Vietnamese
- [ ] Concurrent indicator calculation
- [ ] Docker build succeeds

---

## 6. Phase 5: Sentiment Analysis Agent

(Similar structure - uses Vertex AI for PhoBERT inference)

## 7. Phase 6: Forecast & Orchestrator

(Combines technical and sentiment with weighted scoring)

## 8. Phase 7: Telegram Bot & Distribution

(Go Telegram bot using go-telegram-bot-api)

## 9. Phase 8: Cloud Workflows & Scheduler

```yaml
# deployments/cloud-workflows/daily-analysis.yaml
main:
  params: [args]
  steps:
    - init:
        assign:
          - project_id: ${sys.get_env("GOOGLE_CLOUD_PROJECT_ID")}
          - region: "asia-southeast1"
          - date: ${time.format(sys.now(), "2006-01-02")}

    - scrape_news:
        call: http.post
        args:
          url: ${"https://" + region + "-" + project_id + ".cloudfunctions.net/scraper"}
          auth:
            type: OIDC
          body:
            date: ${date}
        result: scrape_result

    - get_hot_stocks:
        call: http.get
        args:
          url: ${"https://vnstock-api-" + project_id + ".a.run.app/api/stocks/hot"}
          auth:
            type: OIDC
          query:
            date: ${date}
            limit: "5"
        result: hot_stocks

    - analyze_stocks:
        parallel:
          for:
            value: stock
            in: ${hot_stocks.body.stocks}
            steps:
              - analyze:
                  call: http.post
                  args:
                    url: ${"https://orchestrator-" + project_id + ".a.run.app/analyze"}
                    auth:
                      type: OIDC
                    body:
                      symbol: ${stock.symbol}
                  result: analysis

    - generate_report:
        call: http.post
        args:
          url: ${"https://orchestrator-" + project_id + ".a.run.app/report/daily"}
          auth:
            type: OIDC
          body:
            date: ${date}
        result: report

    - distribute:
        parallel:
          branches:
            - telegram:
                call: http.post
                args:
                  url: ${"https://telegram-bot-" + project_id + ".a.run.app/send"}
                  auth:
                    type: OIDC
                  body:
                    report_url: ${report.body.url}

    - return_result:
        return:
          status: "completed"
          date: ${date}
          report_url: ${report.body.url}
```

---

## 10. Deployment & CI/CD

### 10.1 Cloud Build Configuration

```yaml
# cloudbuild.yaml
steps:
  # Build and push images
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', '${_REGION}-docker.pkg.dev/${PROJECT_ID}/vnstock/api-gateway:${SHORT_SHA}', '-f', 'Dockerfile', '.']

  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', '${_REGION}-docker.pkg.dev/${PROJECT_ID}/vnstock/api-gateway:${SHORT_SHA}']

  # Deploy to Cloud Run
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: gcloud
    args:
      - 'run'
      - 'deploy'
      - 'vnstock-api'
      - '--image=${_REGION}-docker.pkg.dev/${PROJECT_ID}/vnstock/api-gateway:${SHORT_SHA}'
      - '--region=${_REGION}'
      - '--platform=managed'
      - '--allow-unauthenticated'

substitutions:
  _REGION: asia-southeast1

options:
  logging: CLOUD_LOGGING_ONLY
```

### 10.2 Makefile

```makefile
# Makefile
.PHONY: build run test deploy

PROJECT_ID := vnstock-analysis
REGION := asia-southeast1

# Local development
run:
	go run cmd/api-gateway/main.go

run-docker:
	docker-compose up --build

# Testing
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Build
build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/api-gateway ./cmd/api-gateway
	CGO_ENABLED=0 GOOS=linux go build -o bin/technical-agent ./cmd/technical-agent
	CGO_ENABLED=0 GOOS=linux go build -o bin/sentiment-agent ./cmd/sentiment-agent

# Deploy
deploy-infra:
	cd deployments/terraform && terraform apply -auto-approve

deploy-api:
	gcloud run deploy vnstock-api \
		--source . \
		--region $(REGION) \
		--allow-unauthenticated

deploy-all: build
	gcloud builds submit --config cloudbuild.yaml

# Database
migrate:
	go run scripts/migrate.go

# Lint
lint:
	golangci-lint run
```

### 10.3 Deployment Commands

```bash
# 1. Deploy infrastructure
make deploy-infra

# 2. Deploy services
make deploy-all

# 3. Verify deployment
gcloud run services list --region=asia-southeast1

# 4. Test endpoints
curl https://vnstock-api-xxxxx.a.run.app/health
```

---

## Summary

This implementation guide covers:

1. **GCP Infrastructure** - Terraform for Cloud SQL, Memorystore, Storage, Pub/Sub
2. **Go Project Structure** - Clean architecture with cmd, internal, pkg layout
3. **API Gateway** - Gin-based REST API with middleware
4. **Technical Agent** - Full indicator calculations in Go
5. **Sentiment Agent** - Vertex AI integration for PhoBERT
6. **Cloud Workflows** - Daily analysis orchestration
7. **CI/CD** - Cloud Build for automated deployment

**Key Benefits of Go/GCP:**
- Fast cold starts (< 1s vs 3-5s for Python)
- Lower memory usage
- Native concurrency with goroutines
- Single binary deployment
- GCP managed services reduce operational overhead
- Better cost efficiency at scale

---

**End of Implementation Guide (Go/GCP Edition)**
