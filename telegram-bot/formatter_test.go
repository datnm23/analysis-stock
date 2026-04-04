package main

import (
	"strings"
	"testing"
)

func TestFormatTechnicalAnalysis(t *testing.T) {
	data := map[string]interface{}{
		"price": map[string]interface{}{
			"close":      50000.0,
			"change_pct": 2.5,
			"volume":     1500000.0,
		},
		"signal": "BUY",
		"score":  7.5,
		"rsi":    65.0,
		"macd": map[string]interface{}{
			"line":   0.5,
			"signal": 0.3,
		},
		"bollinger": map[string]interface{}{
			"lower":  48000.0,
			"middle": 50000.0,
			"upper":  52000.0,
		},
		"sma_20": 49500.0,
		"reasons": []interface{}{
			"RSI ở vùng trung tính",
			"MACD cắt lên signal",
		},
	}

	result := FormatTechnicalAnalysis("VNM", data)

	checks := []string{
		"VNM",
		"50000",
		"2.50%",
		"1.5M",
		"BUY",
		"7.5",
		"RSI",
		"MACD",
		"BB:",
		"SMA20",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("output should contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatForecast(t *testing.T) {
	data := map[string]interface{}{
		"recommendation":  "STRONG_BUY",
		"confidence":      85.0,
		"technical_score":  75.0,
		"sentiment_score":  60.0,
		"market_score":     70.0,
		"combined_score":   68.5,
		"support_price":    48000.0,
		"resistance_price": 52000.0,
		"reasoning": []interface{}{
			"Kỹ thuật tích cực",
			"Sentiment trung lập",
		},
	}

	result := FormatForecast("FPT", data)

	checks := []string{"FPT", "STRONG_BUY", "85.0%", "75.0", "60.0", "70.0", "48000", "52000"}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("output should contain %q, got:\n%s", check, result)
		}
	}
}

func TestFormatDailyReport(t *testing.T) {
	data := map[string]interface{}{
		"date":                    "2024-01-15",
		"total_symbols_analyzed":  10.0,
		"buy_signals":             5.0,
		"sell_signals":            2.0,
		"hold_signals":            3.0,
		"market_summary":          "Thị trường tích cực hôm nay",
		"top_picks": []interface{}{
			map[string]interface{}{
				"symbol":         "VNM",
				"recommendation": "BUY",
				"confidence":     80.0,
			},
		},
	}

	result := FormatDailyReport(data)

	checks := []string{"2024-01-15", "10 mã", "Mua: 5", "Bán: 2", "VNM", "BUY"}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("output should contain %q, got:\n%s", check, result)
		}
	}
}

func TestSignalEmoji(t *testing.T) {
	tests := []struct {
		signal string
		want   string
	}{
		{"STRONG_BUY", "🟢🟢"},
		{"BUY", "🟢"},
		{"HOLD", "⚪"},
		{"SELL", "🔴"},
		{"STRONG_SELL", "🔴🔴"},
		{"UNKNOWN", "❓"},
	}

	for _, tt := range tests {
		got := signalEmoji(tt.signal)
		if got != tt.want {
			t.Errorf("signalEmoji(%q) = %q, want %q", tt.signal, got, tt.want)
		}
	}
}

func TestRsiLevel(t *testing.T) {
	tests := []struct {
		rsi     float64
		want    string
	}{
		{75, "⚠️ Quá mua"},
		{25, "💡 Quá bán"},
		{50, ""},
	}

	for _, tt := range tests {
		got := rsiLevel(tt.rsi)
		if got != tt.want {
			t.Errorf("rsiLevel(%v) = %q, want %q", tt.rsi, got, tt.want)
		}
	}
}

func TestFormatVolume(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{1500000, "1.5M"},
		{500000, "500K"},
		{500, "500"},
	}

	for _, tt := range tests {
		got := formatVolume(tt.v)
		if got != tt.want {
			t.Errorf("formatVolume(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestFormatTechnicalAnalysis_EmptyData(t *testing.T) {
	result := FormatTechnicalAnalysis("VNM", map[string]interface{}{})
	if !strings.Contains(result, "VNM") {
		t.Error("should still contain symbol name")
	}
}

func TestFormatForecast_EmptyData(t *testing.T) {
	result := FormatForecast("FPT", map[string]interface{}{})
	if !strings.Contains(result, "FPT") {
		t.Error("should still contain symbol name")
	}
}
