package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Config holds the bot configuration
type Config struct {
	BotToken         string
	ChannelID        string
	UrgentChannelID  string
	APIGatewayURL    string
	WebhookURL       string
	UseWebhook       bool
	Port             string
}

func loadConfig() Config {
	return Config{
		BotToken:        getEnv("TELEGRAM_BOT_TOKEN", ""),
		ChannelID:       getEnv("TELEGRAM_CHANNEL_ID", "@vnstock_alerts"),
		UrgentChannelID: getEnv("TELEGRAM_URGENT_CHANNEL_ID", "@vnstock_urgent_alerts"),
		APIGatewayURL:   getEnv("API_GATEWAY_URL", "http://api-gateway:8080"),
		WebhookURL:      getEnv("TELEGRAM_WEBHOOK_URL", ""),
		UseWebhook:      getEnv("TELEGRAM_USE_WEBHOOK", "false") == "true",
		Port:            getEnv("PORT", "8084"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	cfg := loadConfig()

	if cfg.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized as @%s", bot.Self.UserName)

	handler := NewBotHandler(bot, cfg)

	if cfg.UseWebhook {
		// Webhook mode (production)
		wh, _ := tgbotapi.NewWebhook(cfg.WebhookURL + "/" + bot.Token)
		if _, err := bot.Request(wh); err != nil {
			log.Fatalf("Failed to set webhook: %v", err)
		}

		updates := bot.ListenForWebhook("/" + bot.Token)

		// Health endpoint — register before starting the server
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "healthy",
				"service": "telegram-bot",
			})
		})

		go func() {
			log.Printf("Webhook server starting on port %s", cfg.Port)
			if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
				log.Fatalf("HTTP server error: %v", err)
			}
		}()

		go handler.ProcessUpdates(updates)
	} else {
		// Long-polling mode (development)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := bot.GetUpdatesChan(u)
		go handler.ProcessUpdates(updates)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Telegram Bot...")
}

// APIClient wraps HTTP calls to the API Gateway
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) GetTechnicalAnalysis(ctx context.Context, symbol string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/technical/%s", c.baseURL, symbol), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

func (c *APIClient) GetForecast(ctx context.Context, symbol string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/forecast/%s", c.baseURL, symbol), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

func (c *APIClient) GetDailyReport(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/reports/daily", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}
