package services

import (
	"os"
	"testing"
	"time"
)

func TestMarketCalendar(t *testing.T) {
	// Create temp config
	configJSON := `{
		"holidays": {
			"2026": [
				{"date": "2026-01-01", "name": "Tết Dương Lịch"},
				{"date": "2026-04-30", "name": "Ngày Giải Phóng Miền Nam"}
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

	t.Run("holiday detection", func(t *testing.T) {
		hol := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		isHol, name := cal.IsHoliday(hol)
		if !isHol {
			t.Error("expected 2026-01-01 to be a holiday")
		}
		if name != "Tết Dương Lịch" {
			t.Errorf("expected Tết Dương Lịch, got %s", name)
		}
	})

	t.Run("weekend detection", func(t *testing.T) {
		sat := time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC) // Saturday
		if !cal.IsWeekend(sat) {
			t.Error("expected Saturday to be weekend")
		}
		mon := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC) // Monday
		if cal.IsWeekend(mon) {
			t.Error("expected Monday not to be weekend")
		}
	})

	t.Run("trading day", func(t *testing.T) {
		// Holiday → not trading
		if cal.IsTradingDay(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)) {
			t.Error("holiday should not be trading day")
		}
		// Weekend → not trading
		if cal.IsTradingDay(time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC)) {
			t.Error("Saturday should not be trading day")
		}
		// Normal weekday → trading
		if !cal.IsTradingDay(time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)) {
			t.Error("Monday should be trading day")
		}
	})
}
