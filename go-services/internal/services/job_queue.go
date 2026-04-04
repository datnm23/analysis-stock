package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Stream and consumer-group names used for async analysis jobs
const (
	analysisStream    = "vnstock:analysis:jobs"
	analysisGroup     = "analysis-workers"
	analysisConsumer  = "worker-1"
	statusKeyPrefix   = "vnstock:analysis:status:"
	statusTTL         = 24 * time.Hour
)

// JobStatus tracks the lifecycle of an async analysis request
type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

// AnalysisJob is the payload pushed onto the Redis Stream
type AnalysisJob struct {
	RequestID string    `json:"request_id"`
	Symbols   []string  `json:"symbols"`
	QueuedAt  time.Time `json:"queued_at"`
}

// JobResult is stored in Redis after a job completes
type JobResult struct {
	RequestID string                       `json:"request_id"`
	Status    JobStatus                    `json:"status"`
	Results   map[string]*ForecastResult   `json:"results,omitempty"`
	Error     string                       `json:"error,omitempty"`
	StartedAt time.Time                    `json:"started_at,omitempty"`
	FinishedAt time.Time                   `json:"finished_at,omitempty"`
}

// JobQueue wraps Redis Streams for durable async analysis jobs.
// It is safe to call Enqueue from multiple goroutines concurrently.
type JobQueue struct {
	rdb         *redis.Client
	forecastSvc *ForecastService
}

// NewJobQueue creates a queue and ensures the consumer group exists.
func NewJobQueue(rdb *redis.Client, forecastSvc *ForecastService) (*JobQueue, error) {
	q := &JobQueue{rdb: rdb, forecastSvc: forecastSvc}
	if err := q.ensureGroup(context.Background()); err != nil {
		return nil, err
	}
	return q, nil
}

// ensureGroup creates the stream consumer group if it does not exist yet.
func (q *JobQueue) ensureGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(ctx, analysisStream, analysisGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// Enqueue adds an analysis job to the stream and records its initial status.
// Returns the request ID so the caller can poll for results.
func (q *JobQueue) Enqueue(ctx context.Context, requestID string, symbols []string) error {
	job := AnalysisJob{
		RequestID: requestID,
		Symbols:   symbols,
		QueuedAt:  time.Now(),
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Push to Redis Stream
	if err := q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: analysisStream,
		Values: map[string]interface{}{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Write initial status so callers can poll immediately
	return q.setStatus(ctx, &JobResult{
		RequestID: requestID,
		Status:    JobStatusQueued,
	})
}

// GetStatus retrieves the current result/status for a given request ID.
// Returns nil if no job with that ID exists.
func (q *JobQueue) GetStatus(ctx context.Context, requestID string) (*JobResult, error) {
	raw, err := q.rdb.Get(ctx, statusKeyPrefix+requestID).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var result JobResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}
	return &result, nil
}

// StartWorker launches a background goroutine that reads jobs from the stream
// and processes them via ForecastService. It blocks until ctx is cancelled.
func (q *JobQueue) StartWorker(ctx context.Context) {
	slog.Info("Analysis job queue worker started", "stream", analysisStream, "group", analysisGroup)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Job queue worker stopped")
			return
		default:
		}

		// Block up to 2 seconds waiting for a new message
		streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    analysisGroup,
			Consumer: analysisConsumer,
			Streams:  []string{analysisStream, ">"},
			Count:    1,
			Block:    2 * time.Second,
		}).Result()

		if err == redis.Nil || (err != nil && err.Error() == "redis: nil") {
			// Timeout — no messages, loop again
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled
			}
			slog.Error("XReadGroup error", "error", err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				q.processMessage(ctx, msg)
			}
		}
	}
}

// processMessage deserialises one stream message, runs the analysis, and
// acknowledges the message whether or not the analysis succeeds.
func (q *JobQueue) processMessage(ctx context.Context, msg redis.XMessage) {
	raw, ok := msg.Values["payload"].(string)
	if !ok {
		slog.Error("Malformed stream message — missing payload field", "id", msg.ID)
		q.rdb.XAck(ctx, analysisStream, analysisGroup, msg.ID)
		return
	}

	var job AnalysisJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		slog.Error("Failed to parse job payload", "error", err)
		q.rdb.XAck(ctx, analysisStream, analysisGroup, msg.ID)
		return
	}

	slog.Info("Processing analysis job", "request_id", job.RequestID, "symbols", job.Symbols)

	// Mark as processing
	startedAt := time.Now()
	_ = q.setStatus(ctx, &JobResult{
		RequestID: job.RequestID,
		Status:    JobStatusProcessing,
		StartedAt: startedAt,
	})

	// Run forecast for each symbol concurrently
	results := make(map[string]*ForecastResult, len(job.Symbols))
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error
	semaphore := make(chan struct{}, reportConcurrencyLimit)

	for _, sym := range job.Symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			res, err := q.forecastSvc.Forecast(ctx, symbol)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.Warn("Forecast failed for symbol", "symbol", symbol, "error", err)
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			results[symbol] = res
		}(sym)
	}
	wg.Wait()

	jobResult := &JobResult{
		RequestID:  job.RequestID,
		Results:    results,
		StartedAt:  startedAt,
		FinishedAt: time.Now(),
	}
	if firstErr != nil && len(results) == 0 {
		jobResult.Status = JobStatusFailed
		jobResult.Error = firstErr.Error()
	} else {
		jobResult.Status = JobStatusDone
	}

	if err := q.setStatus(ctx, jobResult); err != nil {
		slog.Error("Failed to persist job result", "request_id", job.RequestID, "error", err)
	}

	// Acknowledge so the message is not redelivered
	q.rdb.XAck(ctx, analysisStream, analysisGroup, msg.ID)
	slog.Info("Job complete", "request_id", job.RequestID, "status", jobResult.Status,
		"duration_ms", time.Since(startedAt).Milliseconds())
}

func (q *JobQueue) setStatus(ctx context.Context, result *JobResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return q.rdb.Set(ctx, statusKeyPrefix+result.RequestID, data, statusTTL).Err()
}
