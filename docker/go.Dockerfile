# Multi-stage build for Go services
FROM golang:1.24-alpine AS builder

ARG SERVICE

WORKDIR /app

# Install git and timezone data
RUN apk add --no-cache git tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./cmd/${SERVICE}

# Final minimal image
FROM gcr.io/distroless/static-debian12

# Include timezone data for Asia/Ho_Chi_Minh support
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/service /service

EXPOSE 8080

# Health checks are handled by docker-compose/k8s probes (distroless has no shell)
ENTRYPOINT ["/service"]
