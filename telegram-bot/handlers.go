package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var symbolRegex = regexp.MustCompile(`^[A-Z0-9]{1,10}$`)

// BotHandler processes incoming Telegram messages
type BotHandler struct {
	bot       *tgbotapi.BotAPI
	apiClient *APIClient
	cfg       Config
}

func NewBotHandler(bot *tgbotapi.BotAPI, cfg Config) *BotHandler {
	return &BotHandler{
		bot:       bot,
		apiClient: NewAPIClient(cfg.APIGatewayURL),
		cfg:       cfg,
	}
}

// ProcessUpdates handles the update channel
func (h *BotHandler) ProcessUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		var reply string
		var parseMode string

		switch update.Message.Command() {
		case "start":
			reply = h.handleStart(update.Message)
		case "help":
			reply = h.handleHelp()
		case "analyze", "a":
			reply, parseMode = h.handleAnalyze(update.Message)
		case "forecast", "f":
			reply, parseMode = h.handleForecast(update.Message)
		case "report", "r":
			reply, parseMode = h.handleReport()
		default:
			reply = "❓ Lệnh không hợp lệ. Gõ /help để xem danh sách lệnh."
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		if parseMode != "" {
			msg.ParseMode = parseMode
		}
		msg.DisableWebPagePreview = true

		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func (h *BotHandler) handleStart(msg *tgbotapi.Message) string {
	return fmt.Sprintf(
		"👋 Xin chào *%s*!\n\n"+
			"Tôi là VNStock Bot — trợ lý phân tích chứng khoán Việt Nam.\n\n"+
			"📊 Gõ /help để xem các lệnh có sẵn.",
		msg.From.FirstName,
	)
}

func (h *BotHandler) handleHelp() string {
	return `📖 *Danh sách lệnh:*

🔍 /analyze <MÃ> — Phân tích kỹ thuật
📈 /forecast <MÃ> — Dự báo tổng hợp
📋 /report — Báo cáo thị trường hôm nay

⚡ *Lệnh tắt:*
/a VNM — Phân tích nhanh VNM
/f HPG — Dự báo nhanh HPG
/r — Báo cáo nhanh

💡 *Ví dụ:*
` + "`/analyze VNM`" + ` — Phân tích kỹ thuật VNM
` + "`/forecast FPT`" + ` — Dự báo tổng hợp FPT`
}

func (h *BotHandler) handleAnalyze(msg *tgbotapi.Message) (string, string) {
	args := msg.CommandArguments()
	symbol := strings.ToUpper(strings.TrimSpace(args))

	if symbol == "" || !symbolRegex.MatchString(symbol) {
		return "⚠️ Vui lòng nhập mã cổ phiếu hợp lệ.\nVí dụ: `/analyze VNM`", "Markdown"
	}

	data, err := h.apiClient.GetTechnicalAnalysis(context.Background(), symbol)
	if err != nil {
		log.Printf("Technical analysis error for %s: %v", symbol, err)
		return fmt.Sprintf("❌ Không thể phân tích *%s*: %v", symbol, err), "Markdown"
	}

	return FormatTechnicalAnalysis(symbol, data), "Markdown"
}

func (h *BotHandler) handleForecast(msg *tgbotapi.Message) (string, string) {
	args := msg.CommandArguments()
	symbol := strings.ToUpper(strings.TrimSpace(args))

	if symbol == "" || !symbolRegex.MatchString(symbol) {
		return "⚠️ Vui lòng nhập mã cổ phiếu hợp lệ.\nVí dụ: `/forecast FPT`", "Markdown"
	}

	data, err := h.apiClient.GetForecast(context.Background(), symbol)
	if err != nil {
		log.Printf("Forecast error for %s: %v", symbol, err)
		return fmt.Sprintf("❌ Không thể dự báo *%s*: %v", symbol, err), "Markdown"
	}

	return FormatForecast(symbol, data), "Markdown"
}

func (h *BotHandler) handleReport() (string, string) {
	data, err := h.apiClient.GetDailyReport(context.Background())
	if err != nil {
		log.Printf("Daily report error: %v", err)
		return "❌ Không thể tải báo cáo: " + err.Error(), ""
	}

	return FormatDailyReport(data), "Markdown"
}

// SendAlert sends a message to the configured channel
func (h *BotHandler) SendAlert(message string, urgent bool) error {
	channelID := h.cfg.ChannelID
	if urgent {
		channelID = h.cfg.UrgentChannelID
	}

	msg := tgbotapi.NewMessageToChannel(channelID, message)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	_, err := h.bot.Send(msg)
	return err
}
