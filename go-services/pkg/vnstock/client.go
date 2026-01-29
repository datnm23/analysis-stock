package vnstock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OHLCV represents a single candlestick data point
type OHLCV struct {
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// Client provides access to Vietnamese stock market data
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new vnstock client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.vietstock.vn/finance",
	}
}

// NewClientWithConfig creates a client with custom configuration
func NewClientWithConfig(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// GetHistoricalData fetches historical OHLCV data for a symbol
func (c *Client) GetHistoricalData(ctx context.Context, symbol string, days int) ([]OHLCV, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Format dates for API
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")

	url := fmt.Sprintf("%s/histdata/%s?from=%s&to=%s", c.baseURL, symbol, start, end)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "VNStock-Hybrid/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data []OHLCV
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return data, nil
}

// GetMockData returns mock historical data for testing
func (c *Client) GetMockData(symbol string, days int) []OHLCV {
	data := make([]OHLCV, days)
	basePrice := 50000.0 // Base price in VND

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -days+i+1)

		// Generate somewhat realistic price movement
		change := (float64(i%10) - 5) * 100
		open := basePrice + change
		close := open + (float64(i%5)-2)*50
		high := max(open, close) + float64(i%3)*30
		low := min(open, close) - float64(i%3)*30
		volume := int64(1000000 + (i%10)*100000)

		data[i] = OHLCV{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		}

		basePrice = close // Use close as next day's base
	}

	return data
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
