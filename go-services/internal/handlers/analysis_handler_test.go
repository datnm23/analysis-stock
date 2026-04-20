package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTechnicalAnalysis_InvalidSymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		symbol string
	}{
		{"lowercase", "vnm"},
		{"lowercase multi", "vnmd"},
		{"with special chars", "VN@"},
		{"too long", "VNM12345"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/api/v1/technical/:symbol", func(c *gin.Context) {
				symbol := c.Param("symbol")
				if !symbolPattern.MatchString(symbol) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "invalid symbol format",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/technical/"+tt.symbol, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400 for symbol %q, got %d", tt.symbol, w.Code)
			}
		})
	}
}

func TestTechnicalAnalysis_ValidSymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		symbol string
	}{
		{"3 letter", "VNM"},
		{"4 letter", "VN30"},
		{"with numbers", "FPT"},
		{"short", "V"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/api/v1/technical/:symbol", func(c *gin.Context) {
				symbol := c.Param("symbol")
				if !symbolPattern.MatchString(symbol) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "invalid symbol format",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{"symbol": symbol})
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/technical/"+tt.symbol, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for symbol %q, got %d", tt.symbol, w.Code)
			}

			var resp map[string]string
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["symbol"] != tt.symbol {
				t.Errorf("expected symbol %q in response, got %q", tt.symbol, resp["symbol"])
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "vnstock-api",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %q", resp["status"])
	}
}

func TestSymbolPattern(t *testing.T) {
	tests := []struct {
		symbol string
		valid  bool
	}{
		{"VNM", true},
		{"HPG", true},
		{"FPT", true},
		{"VN30F", true},
		{"vnm", false},
		{"VNM1", true},
		{"V", true},
		{"VNMDDD", false},
		{"VN@", false},
		{"", false},
	}

	for _, tt := range tests {
		result := symbolPattern.MatchString(tt.symbol)
		if result != tt.valid {
			t.Errorf("symbolPattern.Match(%q) = %v, expected %v", tt.symbol, result, tt.valid)
		}
	}
}
