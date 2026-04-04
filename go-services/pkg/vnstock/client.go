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

// Client provides access to Vietnamese stock market data via TCBS API
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// defaultBaseURL is the TCBS public API — no authentication required
const defaultBaseURL = "https://apipubaws.tcbs.com.vn/stock-insight/v1"

// NewClient creates a new vnstock client using the default TCBS endpoint
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

// tcbsResponse is the top-level response from the TCBS bars API
type tcbsResponse struct {
	Data []tcbsBar `json:"data"`
}

// tcbsBar holds one day of OHLCV data returned by TCBS
type tcbsBar struct {
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	Volume      int64   `json:"volume"`
	TradingDate string  `json:"tradingDate"`
}

// tradingDateFormats lists the date layouts TCBS may return
var tradingDateFormats = []string{
	"2006-01-02",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05Z",
	time.RFC3339,
}

// parseTradingDate tries each known layout until one succeeds
func parseTradingDate(s string) (time.Time, error) {
	for _, layout := range tradingDateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse trading date: %q", s)
}

// GetHistoricalData fetches historical OHLCV data from the TCBS public API.
// It retries up to 2 times with exponential back-off before returning an error.
func (c *Client) GetHistoricalData(ctx context.Context, symbol string, days int) ([]OHLCV, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	url := fmt.Sprintf(
		"%s/stock/bars-long-term?ticker=%s&type=stock&resolution=D&from=%d&to=%d",
		c.baseURL, symbol, start.Unix(), end.Unix(),
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

// doFetch performs a single HTTP GET and parses the TCBS response
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

	var parsed tcbsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("empty data array in response")
	}

	ohlcv := make([]OHLCV, 0, len(parsed.Data))
	for _, bar := range parsed.Data {
		date, err := parseTradingDate(bar.TradingDate)
		if err != nil {
			continue // skip malformed rows
		}
		ohlcv = append(ohlcv, OHLCV{
			Date:   date,
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

// GetMarketIndex fetches the latest snapshot of a market index (e.g. "VNINDEX", "VN30").
func (c *Client) GetMarketIndex(ctx context.Context, index string) (*MarketIndex, error) {
	url := fmt.Sprintf("%s/stock/overview?symbol=%s", c.baseURL, index)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create index request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch index %s: %w", index, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index %s returned status %d", index, resp.StatusCode)
	}

	var raw struct {
		Symbol       string  `json:"ticker"`
		Price        float64 `json:"marketPrice"`
		PriceChange  float64 `json:"priceChange"`
		ChangePct    float64 `json:"percentChange"`
		TotalVolume  int64   `json:"totalVolume"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode index %s: %w", index, err)
	}

	return &MarketIndex{
		Name:      index,
		Value:     raw.Price,
		Change:    raw.PriceChange,
		ChangePct: raw.ChangePct,
		Volume:    raw.TotalVolume,
	}, nil
}

// GetForeignFlow fetches foreign investor net buy/sell data for a symbol.
func (c *Client) GetForeignFlow(ctx context.Context, symbol string) (*ForeignFlow, error) {
	url := fmt.Sprintf("%s/stock/overview?symbol=%s", c.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create foreign flow request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch foreign flow %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("foreign flow %s returned status %d", symbol, resp.StatusCode)
	}

	var raw struct {
		ForeignBuyVol  int64   `json:"foreignBuyVolume"`
		ForeignSellVol int64   `json:"foreignSellVolume"`
		ForeignBuyVal  float64 `json:"foreignBuyValue"`
		ForeignSellVal float64 `json:"foreignSellValue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode foreign flow %s: %w", symbol, err)
	}

	return &ForeignFlow{
		Symbol:    symbol,
		NetBuyVol: raw.ForeignBuyVol - raw.ForeignSellVol,
		NetBuyVal: raw.ForeignBuyVal - raw.ForeignSellVal,
		BuyVol:    raw.ForeignBuyVol,
		SellVol:   raw.ForeignSellVol,
	}, nil
}
