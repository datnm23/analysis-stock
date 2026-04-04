package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	cacheOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total cache operations (hit/miss)",
		},
		[]string{"type", "result"},
	)
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(cacheOpsTotal)
	prometheus.MustRegister(circuitBreakerState)
}

// PrometheusMetrics returns a Gin middleware that records request metrics.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // normalized route pattern, e.g. "/api/v1/technical/:symbol"
		if path == "" {
			path = "unknown"
		}

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler returns the Prometheus metrics HTTP handler for mounting on a route.
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordCacheHit increments the cache hit counter.
func RecordCacheHit(cacheType string) {
	cacheOpsTotal.WithLabelValues(cacheType, "hit").Inc()
}

// RecordCacheMiss increments the cache miss counter.
func RecordCacheMiss(cacheType string) {
	cacheOpsTotal.WithLabelValues(cacheType, "miss").Inc()
}

// RecordCircuitBreakerState records the current circuit breaker state for a service.
func RecordCircuitBreakerState(service string, state int) {
	circuitBreakerState.WithLabelValues(service).Set(float64(state))
}
