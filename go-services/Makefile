.PHONY: all build clean test test-indicators coverage fmt vet run-api-gateway run-technical run-forecast run-orchestrator

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directory
BUILD_DIR=./bin

# Services
SERVICES=api-gateway technical-agent forecast-agent master-orchestrator

all: build

build: $(SERVICES)

$(SERVICES):
	$(GOBUILD) -o $(BUILD_DIR)/$@ ./cmd/$@

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	$(GOTEST) -v -race -timeout=5m ./...

test-indicators:
	$(GOTEST) -v -race ./internal/indicators/...

coverage:
	$(GOTEST) -race -timeout=5m -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run individual services
run-api-gateway:
	$(GOCMD) run ./cmd/api-gateway

run-technical:
	$(GOCMD) run ./cmd/technical-agent

run-forecast:
	$(GOCMD) run ./cmd/forecast-agent

run-orchestrator:
	$(GOCMD) run ./cmd/master-orchestrator

# Docker builds
docker-build:
	docker build -t vnstock-api-gateway --build-arg SERVICE=api-gateway -f ../docker/go.Dockerfile .
	docker build -t vnstock-technical --build-arg SERVICE=technical-agent -f ../docker/go.Dockerfile .
	docker build -t vnstock-forecast --build-arg SERVICE=forecast-agent -f ../docker/go.Dockerfile .
	docker build -t vnstock-orchestrator --build-arg SERVICE=master-orchestrator -f ../docker/go.Dockerfile .

# Lint
lint:
	golangci-lint run ./...

# Generate mocks
mocks:
	mockgen -source=internal/services/technical_service.go -destination=internal/mocks/technical_service_mock.go

# Database migrations (requires golang-migrate CLI)
# Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
DB_URL ?= postgres://vnstock:$${POSTGRES_PASSWORD}@localhost:5432/vnstock?sslmode=disable

migrate-up:
	migrate -database "$(DB_URL)" -path migrations up

migrate-down:
	migrate -database "$(DB_URL)" -path migrations down

migrate-down-1:
	migrate -database "$(DB_URL)" -path migrations down 1

migrate-version:
	migrate -database "$(DB_URL)" -path migrations version

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

