package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SentimentClient handles communication with Python sentiment service
type SentimentClient struct {
	baseURL    string
	httpClient *http.Client
}

// TextItem represents a text to analyze
type TextItem struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Source      string    `json:"source,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
}

// SentimentRequest is the request payload for sentiment analysis
type SentimentRequest struct {
	Texts []TextItem `json:"texts"`
}

// SentimentResponse is the response from sentiment service
type SentimentResponse struct {
	Results          []SentimentResult `json:"results"`
	ProcessingTimeMs float64           `json:"processing_time_ms"`
	ModelVersion     string            `json:"model_version"`
}

// SentimentResult represents a single sentiment analysis result
type SentimentResult struct {
	ID         string   `json:"id"`
	Sentiment  string   `json:"sentiment"`
	Confidence float64  `json:"confidence"`
	Symbols    []string `json:"symbols"`
	Keywords   []string `json:"keywords"`
}

// NewSentimentClient creates a new sentiment service client
func NewSentimentClient(baseURL string) *SentimentClient {
	return &SentimentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for ML inference
		},
	}
}

// Analyze sends texts to sentiment service for analysis
func (c *SentimentClient) Analyze(ctx context.Context, texts []TextItem) (*SentimentResponse, error) {
	reqBody := SentimentRequest{Texts: texts}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/analyze", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sentiment service returned status %d", resp.StatusCode)
	}

	var result SentimentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Health checks if the sentiment service is healthy
func (c *SentimentClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sentiment service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
