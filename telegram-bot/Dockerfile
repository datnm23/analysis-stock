# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /telegram-bot .

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

# Set Vietnam timezone
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app
COPY --from=builder /telegram-bot .

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -q --spider http://localhost:${PORT:-8084}/health || exit 1

EXPOSE 8084

CMD ["./telegram-bot"]
