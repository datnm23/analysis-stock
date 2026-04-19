package vnstock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MarketDataProvider is the interface for fetching market data.
// Services should depend on this interface for testability.
type MarketDataProvider interface {
	GetHistoricalData(ctx context.Context, symbol string, days int) ([]OHLCV, error)
	GetMockData(symbol string, days int) []OHLCV
	GetMarketIndex(ctx context.Context, index string) (*MarketIndex, error)
	GetForeignFlow(ctx context.Context, symbol string) (*ForeignFlow, error)
}

// MarketIndex represents a market index (e.g. VNINDEX, VN30) snapshot.
type MarketIndex struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Change    float64 `json:"change"`   // absolute change
	ChangePct float64 `json:"changePct"` // percentage change
	Volume    int64   `json:"volume"`
}

// ForeignFlow represents foreign investor net buy/sell data.
type ForeignFlow struct {
	Symbol     string  `json:"symbol"`
	NetBuyVol  int64   `json:"netBuyVol"`   // positive = net buy
	NetBuyVal  float64 `json:"netBuyVal"`   // value in VND
	BuyVol     int64   `json:"buyVol"`
	SellVol    int64   `json:"sellVol"`
}

// OHLCV represents a single candlestick data point
type OHLCV struct {
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// Client provides access to Vietnamese stock market data via KBS Securities API
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// defaultBaseURL is the KB Securities public API — no authentication required
const defaultBaseURL = "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks"

// NewClient creates a new vnstock client using the default KBS endpoint
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: defaultBaseURL,
	}
}

// NewClientWithConfig creates a client with a custom base URL and timeout
func NewClientWithConfig(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// kbsResponse is the top-level response from the KBS data_day API
type kbsResponse struct {
	Symbol  string   `json:"symbol"`
	DataDay []kbsBar `json:"data_day"`
}

// kbsBar holds one day of OHLCV data returned by KBS Securities
type kbsBar struct {
	Time   string  `json:"t"`
	Open   float64 `json:"o"`
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Close  float64 `json:"c"`
	Volume int64   `json:"v"`
}

// GetHistoricalData fetches historical OHLCV data from the KBS Securities API.
// It retries up to 3 times with exponential back-off before returning an error.
func (c *Client) GetHistoricalData(ctx context.Context, symbol string, days int) ([]OHLCV, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	url := fmt.Sprintf(
		"%s/%s/data_day?sdate=%s&edate=%s",
		c.baseURL, symbol,
		start.Format("02-01-2006"),
		end.Format("02-01-2006"),
	)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(1<<uint(attempt-1)) * time.Second):
			}
		}

		data, err := c.doFetch(ctx, url)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all attempts failed for %s: %w", symbol, lastErr)
}

// doFetch performs a single HTTP GET and parses the KBS response
func (c *Client) doFetch(ctx context.Context, url string) ([]OHLCV, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; VNStock-Hybrid/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var parsed kbsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(parsed.DataDay) == 0 {
		return nil, fmt.Errorf("empty data array in response")
	}

	ohlcv := make([]OHLCV, 0, len(parsed.DataDay))
	for _, bar := range parsed.DataDay {
		t, err := time.Parse("2006-01-02 15:04", bar.Time)
		if err != nil {
			t, err = time.Parse("2006-01-02", bar.Time[:10])
			if err != nil {
				continue
			}
		}
		ohlcv = append(ohlcv, OHLCV{
			Date:   t,
			Open:   bar.Open,
			High:   bar.High,
			Low:    bar.Low,
			Close:  bar.Close,
			Volume: bar.Volume,
		})
	}

	if len(ohlcv) == 0 {
		return nil, fmt.Errorf("no valid data points could be parsed")
	}

	return ohlcv, nil
}

// GetMockData returns synthetic historical data for testing / offline mode
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

// GetMarketIndex returns a stub market index — live index data not available from KBS API.
func (c *Client) GetMarketIndex(_ context.Context, index string) (*MarketIndex, error) {
	return &MarketIndex{Name: index}, nil
}

// GetForeignFlow returns a stub foreign flow — live data not available from KBS API.
func (c *Client) GetForeignFlow(_ context.Context, symbol string) (*ForeignFlow, error) {
	return &ForeignFlow{Symbol: symbol}, nil
}
