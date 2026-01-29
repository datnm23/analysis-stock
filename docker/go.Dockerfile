# Multi-stage build for Go services
FROM golang:1.22-alpine AS builder

ARG SERVICE

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./cmd/${SERVICE}

# Final minimal image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/service /service

EXPOSE 8080

ENTRYPOINT ["/service"]
