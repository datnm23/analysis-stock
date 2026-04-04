package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// AnomalyDetector monitors sentiment scores per symbol in a sliding window
// and flags statistically anomalous spikes that may indicate coordinated
// manipulation (e.g. Telegram pump groups).
type AnomalyDetector struct {
	redis        *redis.Client
	windowSize   time.Duration // sliding window to accumulate scores
	spikeStdDevs float64       // how many σ triggers an anomaly
}

// SentimentAlert represents a detected anomaly.
type SentimentAlert struct {
	Symbol    string    `json:"symbol"`
	AlertType string    `json:"alert_type"` // SENTIMENT_SPIKE, SOURCE_FLOOD
	Severity  string    `json:"severity"`   // LOW, MEDIUM, HIGH
	Mean      float64   `json:"mean"`
	StdDev    float64   `json:"std_dev"`
	Current   float64   `json:"current_score"`
	Deviation float64   `json:"deviation_sigma"`
	Timestamp time.Time `json:"detected_at"`
	Detail    string    `json:"detail"`
}

// NewAnomalyDetector creates a detector with sensible defaults.
func NewAnomalyDetector(rdb *redis.Client) *AnomalyDetector {
	return &AnomalyDetector{
		redis:        rdb,
		windowSize:   6 * time.Hour,
		spikeStdDevs: 2.0,
	}
}

// RecordAndCheck records a new sentiment score for a symbol and checks
// whether it constitutes an anomaly relative to the recent history.
//
// Returns a non-nil *SentimentAlert if an anomaly is detected, or nil
// if the score is within normal bounds.
func (d *AnomalyDetector) RecordAndCheck(ctx context.Context, symbol string, score float64) (*SentimentAlert, error) {
	if d.redis == nil {
		return nil, nil // no-op when Redis unavailable
	}

	key := fmt.Sprintf("sentiment:history:%s", symbol)
	now := time.Now()
	cutoff := now.Add(-d.windowSize)

	member := redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%.2f:%d", score, now.UnixMilli()),
	}

	pipe := d.redis.Pipeline()
	// Add the new score
	pipe.ZAdd(ctx, key, member)
	// Remove entries older than the window
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoff.UnixMilli()))
	// Keep a TTL on the key
	pipe.Expire(ctx, key, d.windowSize+time.Hour)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("anomaly detector pipeline: %w", err)
	}

	// Retrieve recent scores
	vals, err := d.redis.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("anomaly detector zrange: %w", err)
	}

	if len(vals) < 5 {
		// Not enough data to compute statistics
		return nil, nil
	}

	scores := make([]float64, 0, len(vals))
	for _, v := range vals {
		var s float64
		fmt.Sscanf(v, "%f:", &s)
		scores = append(scores, s)
	}

	mean, stddev := meanStdDev(scores)
	if stddev < 1.0 {
		stddev = 1.0 // prevent division by near-zero
	}

	deviation := math.Abs(score-mean) / stddev

	if deviation >= d.spikeStdDevs {
		severity := "MEDIUM"
		if deviation >= 3.0 {
			severity = "HIGH"
		}

		alert := &SentimentAlert{
			Symbol:    symbol,
			AlertType: "SENTIMENT_SPIKE",
			Severity:  severity,
			Mean:      math.Round(mean*100) / 100,
			StdDev:    math.Round(stddev*100) / 100,
			Current:   score,
			Deviation: math.Round(deviation*100) / 100,
			Timestamp: now,
			Detail: fmt.Sprintf(
				"Sentiment for %s deviated %.1fσ from 6h mean (%.1f vs avg %.1f±%.1f)",
				symbol, deviation, score, mean, stddev,
			),
		}
		slog.Warn("Sentiment anomaly detected",
			"symbol", symbol,
			"deviation_sigma", deviation,
			"score", score,
			"mean", mean,
		)
		return alert, nil
	}

	return nil, nil
}

// WeightPenalty returns a multiplier [0.05, 1.0] to apply to sentiment
// weight when an anomaly is detected.
func (a *SentimentAlert) WeightPenalty() float64 {
	switch a.Severity {
	case "HIGH":
		return 0.05 // almost ignore sentiment
	case "MEDIUM":
		return 0.25
	default:
		return 0.50
	}
}

func meanStdDev(values []float64) (float64, float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / n

	var variance float64
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	variance /= n

	return mean, math.Sqrt(variance)
}
