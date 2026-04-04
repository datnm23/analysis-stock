//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"vnstock-hybrid/internal/services"
	"vnstock-hybrid/internal/testutil"
)

func TestJobQueue_EndToEnd(t *testing.T) {
	ctx := context.Background()

	rdb, cleanupRedis := testutil.SetupRedis(ctx)
	defer cleanupRedis()

	db, cleanupDB := testutil.SetupPostgres(ctx)
	defer cleanupDB()

	// Build services
	provider := &mockMarketDataProvider{data: generateTestOHLCV(100)}
	technicalSvc := services.NewTechnicalService(db, rdb, provider)
	sentimentClient := services.NewSentimentClient("http://localhost:0") // will fail, that's OK
	forecastSvc := services.NewForecastService(db, rdb, technicalSvc, sentimentClient)

	// Create job queue
	jq, err := services.NewJobQueue(rdb, forecastSvc)
	if err != nil {
		t.Fatalf("NewJobQueue failed: %v", err)
	}

	// Start worker in background
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	go jq.StartWorker(workerCtx)

	// Enqueue a job
	symbols := []string{"VNM", "FPT"}
	requestID := uuid.New().String()
	err = jq.Enqueue(ctx, requestID, symbols)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	t.Logf("Enqueued job: %s", requestID)

	// Verify initial status is queued
	status, err := jq.GetStatus(ctx, requestID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status == nil {
		t.Fatal("expected status, got nil")
	}
	if status.Status != services.JobStatusQueued {
		t.Errorf("expected queued status, got %s", status.Status)
	}

	// Poll for completion (max 30 seconds)
	deadline := time.Now().Add(30 * time.Second)
	var finalStatus services.JobStatus
	for time.Now().Before(deadline) {
		result, err := jq.GetStatus(ctx, requestID)
		if err != nil {
			t.Fatalf("GetStatus failed: %v", err)
		}
		if result == nil {
			t.Fatal("status disappeared")
		}
		finalStatus = result.Status
		if finalStatus == services.JobStatusDone || finalStatus == services.JobStatusFailed {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Job should finish (sentiment fails but forecast degrades gracefully → done or failed)
	t.Logf("Final job status: %s", finalStatus)
	if finalStatus != services.JobStatusDone && finalStatus != services.JobStatusFailed {
		t.Errorf("job did not finish in time, status: %s", finalStatus)
	}
}
