package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// HolidayConfig represents the market_holidays.json structure.
type HolidayConfig struct {
	Holidays    map[string][]HolidayEntry `json:"holidays"`
	WeekendDays []string                  `json:"weekend_days"`
	MarketHours MarketHours               `json:"market_hours"`
}

// HolidayEntry represents a single holiday.
type HolidayEntry struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

// MarketHours represents trading hours.
type MarketHours struct {
	Open       string `json:"open"`
	LunchStart string `json:"lunch_start"`
	LunchEnd   string `json:"lunch_end"`
	Close      string `json:"close"`
	Timezone   string `json:"timezone"`
}

// MarketCalendar provides holiday and trading-hours checking.
type MarketCalendar struct {
	config   HolidayConfig
	holidays map[string]string // date→name lookup
	tz       *time.Location
}

// LoadMarketCalendar reads the holiday config from the given JSON path.
func LoadMarketCalendar(configPath string) (*MarketCalendar, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read holiday config: %w", err)
	}

	var cfg HolidayConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse holiday config: %w", err)
	}

	tz, err := time.LoadLocation(cfg.MarketHours.Timezone)
	if err != nil {
		tz = time.FixedZone("ICT", 7*3600)
	}

	// Build fast lookup map
	holidays := make(map[string]string)
	for _, entries := range cfg.Holidays {
		for _, e := range entries {
			holidays[e.Date] = e.Name
		}
	}

	slog.Info("Market calendar loaded", "holidays", len(holidays), "timezone", tz.String())

	return &MarketCalendar{
		config:   cfg,
		holidays: holidays,
		tz:       tz,
	}, nil
}

// IsHoliday checks if the given date is a market holiday.
func (mc *MarketCalendar) IsHoliday(t time.Time) (bool, string) {
	key := t.In(mc.tz).Format("2006-01-02")
	name, ok := mc.holidays[key]
	return ok, name
}

// IsWeekend checks if the given date is a weekend.
func (mc *MarketCalendar) IsWeekend(t time.Time) bool {
	day := t.In(mc.tz).Weekday().String()
	for _, wd := range mc.config.WeekendDays {
		if day == wd {
			return true
		}
	}
	return false
}

// IsTradingDay checks if the given date is a trading day.
func (mc *MarketCalendar) IsTradingDay(t time.Time) bool {
	if mc.IsWeekend(t) {
		return false
	}
	isHol, _ := mc.IsHoliday(t)
	return !isHol
}

// IsMarketOpen checks if the market is currently open.
func (mc *MarketCalendar) IsMarketOpen(t time.Time) bool {
	if !mc.IsTradingDay(t) {
		return false
	}

	local := t.In(mc.tz)
	now := local.Format("15:04")

	// Morning session: open -> lunch_start
	if now >= mc.config.MarketHours.Open && now < mc.config.MarketHours.LunchStart {
		return true
	}
	// Afternoon session: lunch_end -> close
	if now >= mc.config.MarketHours.LunchEnd && now < mc.config.MarketHours.Close {
		return true
	}
	return false
}
