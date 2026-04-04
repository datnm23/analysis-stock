package middleware

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RequestIDHeader is the HTTP header used to propagate trace correlation IDs.
// Downstream services must forward this header to preserve the trace chain.
const RequestIDHeader = "X-Request-ID"

// contextKey is an unexported type to avoid key collisions in context values.
type contextKey string

const requestIDKey contextKey = "request_id"

// CorrelationID is a middleware that assigns a unique X-Request-ID to every
// incoming request. If the client already provides one, it is reused so that
// external callers (n8n, curl, etc.) can correlate their own logs.
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		// Make it available both in the Gin context and the stdlib context
		c.Set("request_id", id)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), requestIDKey, id),
		)
		// Echo it back so clients can correlate responses with their requests
		c.Header(RequestIDHeader, id)

		// Annotate the active OTel span (created by otelgin before this middleware).
		// This links the human-readable X-Request-ID to the distributed trace.
		if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
			span.SetAttributes(attribute.String("http.request_id", id))
		}

		c.Next()
	}
}

// RequestIDFromContext extracts the request ID from a stdlib context.
// Returns an empty string if not set (e.g. background jobs).
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// newRequestID generates a short random hex string — 16 hex chars (8 bytes).
// Using crypto/rand keeps IDs unpredictable and collision-resistant.
func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to a fixed placeholder so the app
		// keeps running rather than crashing.
		return "fallback-id"
	}
	return fmt.Sprintf("%x-%x", b[:4], b[4:])
}
