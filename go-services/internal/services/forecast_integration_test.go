package services

import (
	"os"
	"testing"
	"time"

	"vnstock-hybrid/internal/indicators"
)

func TestCalculateMarketScore_WithMarketContext(t *testing.T) {
	tests := []struct {
		name     string
		tech     TechnicalResult
		minScore float64
		maxScore float64
	}{
		{
			name: "neutral baseline",
			tech: TechnicalResult{
				Price: PriceData{Close: 100},
			},
			minScore: 45,
			maxScore: 55,
		},
		{
			name: "bullish with strong ADX and VN-Index rising",
			tech: TechnicalResult{
				Price:            PriceData{Close: 100},
				ADX:              &indicators.ADX{ADX: 35, PlusDI: 25, MinusDI: 10},
				SMA20:            90,
				ATR:              1.0,
				VNIndexChangePct: 2.5,
				ForeignNetBuyVol: 50000,
			},
			minScore: 80,
			maxScore: 100,
		},
		{
			name: "bearish with foreign selling",
			tech: TechnicalResult{
				Price:            PriceData{Close: 100},
				ADX:              &indicators.ADX{ADX: 35, PlusDI: 10, MinusDI: 25},
				SMA20:            110,
				ATR:              6.0,
				VNIndexChangePct: -2.0,
				ForeignNetBuyVol: -30000,
			},
			minScore: 0,
			maxScore: 35,
		},
		{
			name: "VN-Index data enriches neutral score",
			tech: TechnicalResult{
				Price:            PriceData{Close: 100},
				VNIndexChangePct: 1.5,
				ForeignNetBuyVol: 10000,
			},
			minScore: 55,
			maxScore: 65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateMarketScore(&tt.tech)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("score = %.1f, want [%.1f, %.1f]", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestMarketCalendar_Integration(t *testing.T) {
	configJSON := `{
		"holidays": {
			"2026": [
				{"date": "2026-01-01", "name": "Tết Dương Lịch"},
				{"date": "2026-04-30", "name": "Ngày Giải Phóng Miền Nam"},
				{"date": "2026-09-02", "name": "Quốc Khánh"}
			]
		},
		"weekend_days": ["Saturday", "Sunday"],
		"market_hours": {
			"open": "09:00", "lunch_start": "11:30",
			"lunch_end": "13:00", "close": "15:00",
			"timezone": "Asia/Ho_Chi_Minh"
		}
	}`

	tmpFile, err := os.CreateTemp("", "holidays*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cal, err := LoadMarketCalendar(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadMarketCalendar: %v", err)
	}

	ict, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	tests := []struct {
		name      string
		time      time.Time
		isTrading bool
		isOpen    bool
	}{
		{
			name:      "Monday morning 10:00 → trading + open",
			time:      time.Date(2026, 3, 30, 10, 0, 0, 0, ict),
			isTrading: true,
			isOpen:    true,
		},
		{
			name:      "Saturday → not trading",
			time:      time.Date(2026, 3, 28, 10, 0, 0, 0, ict),
			isTrading: false,
			isOpen:    false,
		},
		{
			name:      "Holiday April 30 → not trading",
			time:      time.Date(2026, 4, 30, 10, 0, 0, 0, ict),
			isTrading: false,
			isOpen:    false,
		},
		{
			name:      "Lunch break 12:00 → trading but closed",
			time:      time.Date(2026, 3, 30, 12, 0, 0, 0, ict),
			isTrading: true,
			isOpen:    false,
		},
		{
			name:      "Afternoon 14:00 → trading + open",
			time:      time.Date(2026, 3, 30, 14, 0, 0, 0, ict),
			isTrading: true,
			isOpen:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cal.IsTradingDay(tt.time); got != tt.isTrading {
				t.Errorf("IsTradingDay = %v, want %v", got, tt.isTrading)
			}
			if got := cal.IsMarketOpen(tt.time); got != tt.isOpen {
				t.Errorf("IsMarketOpen = %v, want %v", got, tt.isOpen)
			}
		})
	}
}
