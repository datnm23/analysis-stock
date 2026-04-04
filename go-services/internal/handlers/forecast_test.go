package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestForecastAnalysis_InvalidSymbol(t *testing.T) {
	router := gin.New()
	// We can't easily mock a ForecastService, but we can test validation
	router.GET("/forecast/:symbol", func(c *gin.Context) {
		symbol := c.Param("symbol")
		if !symbolPattern.MatchString(symbol) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol format"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"symbol": symbol})
	})

	tests := []struct {
		name   string
		symbol string
		status int
	}{
		{"valid VNM", "VNM", http.StatusOK},
		{"valid FPT", "FPT", http.StatusOK},
		{"valid VNINDEX", "VNINDEX", http.StatusOK},
		{"valid VN30", "VN30", http.StatusOK},
		{"invalid lowercase", "vnm", http.StatusBadRequest},
		{"invalid too long", "ABCDEFGHIJK", http.StatusBadRequest},
		{"invalid empty", "", http.StatusNotFound}, // Gin returns 404 for missing path param
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/forecast/"+tt.symbol, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Errorf("GET /forecast/%s = %d, want %d", tt.symbol, w.Code, tt.status)
			}
		})
	}
}

func TestMarketData_MissingSymbol(t *testing.T) {
	router := gin.New()
	router.POST("/market/data", func(c *gin.Context) {
		var req struct {
			Symbol string `json:"symbol" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !symbolPattern.MatchString(req.Symbol) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbol"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"symbol": req.Symbol})
	})

	// Missing symbol
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/market/data", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /market/data without symbol = %d, want 400", w.Code)
	}

	// Valid symbol
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/market/data", strings.NewReader(`{"symbol":"VNM"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /market/data with VNM = %d, want 200", w.Code)
	}
}

func TestSymbolPattern(t *testing.T) {
	valid := []string{"VNM", "FPT", "VIC", "HPG", "ABC", "VNINDEX", "VN30", "C4G", "123", "V1M"}
	invalid := []string{"vnm", "abc", "ABCDEFGHIJK", ""} // lowercase, >10 chars, empty

	for _, s := range valid {
		if !symbolPattern.MatchString(s) {
			t.Errorf("symbolPattern should match %q", s)
		}
	}
	for _, s := range invalid {
		if symbolPattern.MatchString(s) {
			t.Errorf("symbolPattern should NOT match %q", s)
		}
	}
}

func TestSynthesizeHandler_EmptyBody(t *testing.T) {
	router := gin.New()
	router.POST("/synthesize", func(c *gin.Context) {
		var req struct {
			Symbol string `json:"symbol"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"symbol": req.Symbol})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/synthesize", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /synthesize with empty body = %d, want 200", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["symbol"] != "" {
		t.Errorf("expected empty symbol, got %q", resp["symbol"])
	}
}
