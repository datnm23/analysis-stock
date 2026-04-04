//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"vnstock-hybrid/internal/services"
)

// findConfigPath walks up from the test directory to find config/market_holidays.json
func findConfigPath() string {
	// Try common relative paths from go-services/tests/integration/
	candidates := []string{
		"../../../config/market_holidays.json",
		"../../config/market_holidays.json",
		"config/market_holidays.json",
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
}

func TestMarketCalendar_LoadFromConfig(t *testing.T) {
	configPath := findConfigPath()
	if configPath == "" {
		t.Skip("config/market_holidays.json not found, skipping")
	}

	cal, err := services.LoadMarketCalendar(configPath)
	if err != nil {
		t.Fatalf("LoadMarketCalendar failed: %v", err)
	}

	if cal == nil {
		t.Fatal("expected non-nil calendar")
	}

	t.Logf("Calendar loaded from: %s", configPath)
}

func TestMarketCalendar_IsHoliday_Known2026(t *testing.T) {
	configPath := findConfigPath()
	if configPath == "" {
		t.Skip("config/market_holidays.json not found, skipping")
	}

	cal, err := services.LoadMarketCalendar(configPath)
	if err != nil {
		t.Fatalf("LoadMarketCalendar failed: %v", err)
	}

	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	tests := []struct {
		date      string
		isHoliday bool
		name      string
	}{
		{"2026-01-01", true, "Tết Dương Lịch"},
		{"2026-01-29", true, "Tết Nguyên Đán (Mùng 1)"},
		{"2026-04-30", true, "Ngày Giải Phóng Miền Nam"},
		{"2026-05-01", true, "Quốc Tế Lao Động"},
		{"2026-09-02", true, "Quốc Khánh"},
		{"2026-03-15", false, ""}, // Regular weekday
		{"2026-06-10", false, ""}, // Regular weekday
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			d, _ := time.ParseInLocation("2006-01-02", tt.date, loc)
			isHol, name := cal.IsHoliday(d)
			if isHol != tt.isHoliday {
				t.Errorf("IsHoliday(%s) = %v, want %v", tt.date, isHol, tt.isHoliday)
			}
			if tt.isHoliday && name != tt.name {
				t.Errorf("Holiday name = %q, want %q", name, tt.name)
			}
		})
	}
}

func TestMarketCalendar_IsTradingDay(t *testing.T) {
	configPath := findConfigPath()
	if configPath == "" {
		t.Skip("config/market_holidays.json not found, skipping")
	}

	cal, err := services.LoadMarketCalendar(configPath)
	if err != nil {
		t.Fatalf("LoadMarketCalendar failed: %v", err)
	}

	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	tests := []struct {
		date      string
		isTrading bool
		reason    string
	}{
		{"2026-03-30", true, "Monday — normal trading day"},
		{"2026-03-28", false, "Saturday — weekend"},
		{"2026-03-29", false, "Sunday — weekend"},
		{"2026-01-01", false, "Holiday — Tết Dương Lịch"},
		{"2026-04-30", false, "Holiday — Giải Phóng Miền Nam"},
		{"2026-04-01", true, "Wednesday — normal trading day"},
	}

	for _, tt := range tests {
		t.Run(tt.date+"_"+tt.reason, func(t *testing.T) {
			d, _ := time.ParseInLocation("2006-01-02", tt.date, loc)
			result := cal.IsTradingDay(d)
			if result != tt.isTrading {
				t.Errorf("IsTradingDay(%s) = %v, want %v (%s)", tt.date, result, tt.isTrading, tt.reason)
			}
		})
	}
}

func TestMarketCalendar_IsMarketOpen(t *testing.T) {
	configPath := findConfigPath()
	if configPath == "" {
		t.Skip("config/market_holidays.json not found, skipping")
	}

	cal, err := services.LoadMarketCalendar(configPath)
	if err != nil {
		t.Fatalf("LoadMarketCalendar failed: %v", err)
	}

	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	tests := []struct {
		datetime string
		isOpen   bool
		reason   string
	}{
		// Trading day (Monday 2026-03-30)
		{"2026-03-30T08:30:00", false, "before market open"},
		{"2026-03-30T09:15:00", true, "morning session"},
		{"2026-03-30T10:30:00", true, "morning session mid"},
		{"2026-03-30T11:45:00", false, "lunch break"},
		{"2026-03-30T12:30:00", false, "lunch break mid"},
		{"2026-03-30T13:15:00", true, "afternoon session"},
		{"2026-03-30T14:30:00", true, "afternoon session mid"},
		{"2026-03-30T15:15:00", false, "after market close"},
		// Weekend
		{"2026-03-28T10:00:00", false, "Saturday — market closed"},
		// Holiday
		{"2026-01-01T10:00:00", false, "Holiday — market closed"},
	}

	for _, tt := range tests {
		t.Run(tt.datetime+"_"+tt.reason, func(t *testing.T) {
			dt, _ := time.ParseInLocation("2006-01-02T15:04:05", tt.datetime, loc)
			result := cal.IsMarketOpen(dt)
			if result != tt.isOpen {
				t.Errorf("IsMarketOpen(%s) = %v, want %v (%s)", tt.datetime, result, tt.isOpen, tt.reason)
			}
		})
	}
}
