//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"vnstock-hybrid/internal/services"
)

func TestSentimentClient_Analyze_Success(t *testing.T) {
	// Mock sentiment service returning valid response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/analyze" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":         "item-1",
					"sentiment":  "positive",
					"confidence": 87.5,
					"symbols":    []string{"VNM"},
					"keywords":   []string{"tăng", "mạnh"},
				},
			},
			"processing_time_ms": 150.0,
			"model_version":      "phobert-base-v1",
		})
	}))
	defer server.Close()

	client := services.NewSentimentClient(server.URL)

	ctx := context.Background()
	texts := []services.TextItem{
		{ID: "item-1", Content: "Cổ phiếu VNM tăng mạnh hôm nay", Source: "test"},
	}

	result, err := client.Analyze(ctx, texts)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Sentiment != "positive" {
		t.Errorf("expected positive sentiment, got %s", result.Results[0].Sentiment)
	}
	if result.Results[0].Confidence != 87.5 {
		t.Errorf("expected 87.5 confidence, got %f", result.Results[0].Confidence)
	}
	if result.ModelVersion != "phobert-base-v1" {
		t.Errorf("expected model version phobert-base-v1, got %s", result.ModelVersion)
	}

	t.Logf("Sentiment: %s (%.1f%%), model: %s, time: %.0fms",
		result.Results[0].Sentiment, result.Results[0].Confidence,
		result.ModelVersion, result.ProcessingTimeMs)
}

func TestSentimentClient_Analyze_ServerError(t *testing.T) {
	// Mock server that returns 500
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := services.NewSentimentClient(server.URL)
	ctx := context.Background()

	texts := []services.TextItem{
		{ID: "item-1", Content: "test text"},
	}

	_, err := client.Analyze(ctx, texts)
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}

	t.Logf("Got expected error: %v (calls: %d)", err, atomic.LoadInt32(&callCount))
}

func TestSentimentClient_Analyze_Timeout(t *testing.T) {
	// Mock slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Longer than context timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := services.NewSentimentClient(server.URL)

	// Use short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	texts := []services.TextItem{
		{ID: "item-1", Content: "test text"},
	}

	_, err := client.Analyze(ctx, texts)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	t.Logf("Got expected timeout error: %v", err)
}

func TestSentimentClient_Health_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"model":  "phobert-base",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := services.NewSentimentClient(server.URL)
	ctx := context.Background()

	err := client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestSentimentClient_Health_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := services.NewSentimentClient(server.URL)
	ctx := context.Background()

	err := client.Health(ctx)
	if err == nil {
		t.Error("expected error for unhealthy service, got nil")
	}

	t.Logf("Got expected unhealthy error: %v", err)
}
