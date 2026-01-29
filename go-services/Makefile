.PHONY: all build clean test run-api-gateway run-technical run-forecast run-orchestrator

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
	$(GOTEST) -v ./...

test-indicators:
	$(GOTEST) -v ./internal/indicators/...

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
