package main

import (
	"fmt"
	"strings"
)

// FormatTechnicalAnalysis formats technical analysis as Telegram Markdown
func FormatTechnicalAnalysis(symbol string, data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📊 *Phân Tích Kỹ Thuật — %s*\n", symbol))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Price info
	if price, ok := data["price"].(map[string]interface{}); ok {
		close := getFloat(price, "close")
		change := getFloat(price, "change_pct")
		emoji := "🟢"
		if change < 0 {
			emoji = "🔴"
		}
		sb.WriteString(fmt.Sprintf("%s *Giá:* %s VNĐ (%.2f%%)\n", emoji, formatNumber(close), change))
		sb.WriteString(fmt.Sprintf("📈 *Volume:* %s\n\n", formatVolume(getFloat(price, "volume"))))
	}

	// Signal
	if signal, ok := data["signal"].(string); ok {
		emoji := signalEmoji(signal)
		sb.WriteString(fmt.Sprintf("%s *Tín hiệu:* `%s`\n", emoji, signal))
	}
	if score, ok := data["score"].(float64); ok {
		sb.WriteString(fmt.Sprintf("🎯 *Điểm:* %.1f/10\n\n", score))
	}

	// Indicators
	sb.WriteString("📉 *Chỉ báo:*\n")
	if rsi := getFloatPtr(data, "rsi"); rsi != nil {
		sb.WriteString(fmt.Sprintf("  • RSI(14): `%.1f` %s\n", *rsi, rsiLevel(*rsi)))
	}
	if macd, ok := data["macd"].(map[string]interface{}); ok {
		line := getFloat(macd, "line")
		signal := getFloat(macd, "signal")
		cross := "—"
		if line > signal {
			cross = "🟢 Bull"
		} else {
			cross = "🔴 Bear"
		}
		sb.WriteString(fmt.Sprintf("  • MACD: `%.2f` / Signal: `%.2f` %s\n", line, signal, cross))
	}
	if bb, ok := data["bollinger"].(map[string]interface{}); ok {
		sb.WriteString(fmt.Sprintf("  • BB: `%.0f` — `%.0f` — `%.0f`\n",
			getFloat(bb, "lower"), getFloat(bb, "middle"), getFloat(bb, "upper")))
	}
	if sma20 := getFloatPtr(data, "sma_20"); sma20 != nil {
		sb.WriteString(fmt.Sprintf("  • SMA20: `%.0f`\n", *sma20))
	}

	// Reasons
	if reasons, ok := data["reasons"].([]interface{}); ok && len(reasons) > 0 {
		sb.WriteString("\n💬 *Nhận xét:*\n")
		for _, r := range reasons {
			if s, ok := r.(string); ok {
				sb.WriteString(fmt.Sprintf("  › %s\n", s))
			}
		}
	}

	sb.WriteString("\n⏰ _Dữ liệu cập nhật tự động_")
	return sb.String()
}

// FormatForecast formats forecast as Telegram Markdown
func FormatForecast(symbol string, data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔮 *Dự Báo Tổng Hợp — %s*\n", symbol))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Recommendation
	if rec, ok := data["recommendation"].(string); ok {
		emoji := signalEmoji(rec)
		sb.WriteString(fmt.Sprintf("%s *Khuyến nghị:* `%s`\n", emoji, rec))
	}
	if conf, ok := data["confidence"].(float64); ok {
		sb.WriteString(fmt.Sprintf("🎯 *Độ tin cậy:* `%.1f%%`\n\n", conf))
	}

	// Component scores
	sb.WriteString("📊 *Điểm thành phần:*\n")
	if ts, ok := data["technical_score"].(float64); ok {
		sb.WriteString(fmt.Sprintf("  • Kỹ thuật (40%%): `%.1f`\n", ts))
	}
	if ss, ok := data["sentiment_score"].(float64); ok {
		sb.WriteString(fmt.Sprintf("  • Sentiment (30%%): `%.1f`\n", ss))
	}
	if ms, ok := data["market_score"].(float64); ok {
		sb.WriteString(fmt.Sprintf("  • Thị trường (30%%): `%.1f`\n", ms))
	}
	if cs, ok := data["combined_score"].(float64); ok {
		sb.WriteString(fmt.Sprintf("  • *Tổng hợp:* `%.1f`\n", cs))
	}

	// Price targets
	if sp, ok := data["support_price"].(float64); ok {
		sb.WriteString(fmt.Sprintf("\n🛡 *Hỗ trợ:* %s VNĐ\n", formatNumber(sp)))
	}
	if rp, ok := data["resistance_price"].(float64); ok {
		sb.WriteString(fmt.Sprintf("🚀 *Kháng cự:* %s VNĐ\n", formatNumber(rp)))
	}

	// Reasoning
	if reasons, ok := data["reasoning"].([]interface{}); ok && len(reasons) > 0 {
		sb.WriteString("\n💬 *Phân tích:*\n")
		max := 5
		if len(reasons) < max {
			max = len(reasons)
		}
		for i := 0; i < max; i++ {
			if s, ok := reasons[i].(string); ok {
				sb.WriteString(fmt.Sprintf("  › %s\n", s))
			}
		}
	}

	return sb.String()
}

// FormatDailyReport formats daily report as Telegram Markdown
func FormatDailyReport(data map[string]interface{}) string {
	var sb strings.Builder

	date := getString(data, "date")
	sb.WriteString(fmt.Sprintf("📋 *Báo Cáo Thị Trường — %s*\n", date))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")

	total := getFloat(data, "total_symbols_analyzed")
	buy := getFloat(data, "buy_signals")
	sell := getFloat(data, "sell_signals")
	hold := getFloat(data, "hold_signals")

	sb.WriteString(fmt.Sprintf("📊 *Tổng phân tích:* %d mã\n", int(total)))
	sb.WriteString(fmt.Sprintf("🟢 Mua: %d | 🔴 Bán: %d | ⚪ Giữ: %d\n\n", int(buy), int(sell), int(hold)))

	// Market summary
	if summary, ok := data["market_summary"].(string); ok {
		sb.WriteString(fmt.Sprintf("📝 %s\n\n", summary))
	}

	// Top picks
	if picks, ok := data["top_picks"].([]interface{}); ok && len(picks) > 0 {
		sb.WriteString("⭐ *Top Picks:*\n")
		for i, p := range picks {
			if pick, ok := p.(map[string]interface{}); ok {
				symbol := getString(pick, "symbol")
				rec := getString(pick, "recommendation")
				conf := getFloat(pick, "confidence")
				sb.WriteString(fmt.Sprintf("  %d. *%s* — `%s` (%.0f%%)\n", i+1, symbol, rec, conf))
			}
		}
	}

	return sb.String()
}

// --- Helper functions ---

func signalEmoji(signal string) string {
	switch signal {
	case "STRONG_BUY":
		return "🟢🟢"
	case "BUY":
		return "🟢"
	case "HOLD":
		return "⚪"
	case "SELL":
		return "🔴"
	case "STRONG_SELL":
		return "🔴🔴"
	default:
		return "❓"
	}
}

func rsiLevel(rsi float64) string {
	switch {
	case rsi > 70:
		return "⚠️ Quá mua"
	case rsi < 30:
		return "💡 Quá bán"
	default:
		return ""
	}
}

func formatNumber(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func formatVolume(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.0fK", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

func getFloat(data map[string]interface{}, key string) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return 0
}

func getFloatPtr(data map[string]interface{}, key string) *float64 {
	if v, ok := data[key].(float64); ok {
		return &v
	}
	return nil
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}
