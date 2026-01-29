package indicators

import (
	"math"
	"testing"
)

const tolerance = 0.0001

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestSMA(t *testing.T) {
	prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// Test SMA with period 5
	sma := SMA(prices, 5)
	if sma == nil {
		t.Fatal("SMA returned nil")
	}

	// First valid SMA at index 4 should be (10+11+12+13+14)/5 = 12
	expected := 12.0
	if !almostEqual(sma[4], expected) {
		t.Errorf("SMA[4] = %v, expected %v", sma[4], expected)
	}

	// Last SMA should be (16+17+18+19+20)/5 = 18
	expected = 18.0
	if !almostEqual(sma[len(sma)-1], expected) {
		t.Errorf("SMA[last] = %v, expected %v", sma[len(sma)-1], expected)
	}
}

func TestEMA(t *testing.T) {
	prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	ema := EMA(prices, 5)
	if ema == nil {
		t.Fatal("EMA returned nil")
	}

	// First EMA should equal first SMA = 12
	if !almostEqual(ema[4], 12.0) {
		t.Errorf("EMA[4] = %v, expected 12.0", ema[4])
	}

	// EMA should be higher than first value for uptrending prices
	if ema[len(ema)-1] <= ema[4] {
		t.Error("EMA should increase for uptrending prices")
	}
}

func TestRSI(t *testing.T) {
	// Test with uptrending prices - RSI should be high
	uptrend := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
	rsi := RSI(uptrend, 14)
	if rsi == nil {
		t.Fatal("RSI returned nil")
	}

	lastRSI := rsi[len(rsi)-1]
	if lastRSI < 70 {
		t.Errorf("RSI for uptrend = %v, expected > 70", lastRSI)
	}

	// Test with downtrending prices - RSI should be low
	downtrend := []float64{25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10}
	rsi = RSI(downtrend, 14)
	lastRSI = rsi[len(rsi)-1]
	if lastRSI > 30 {
		t.Errorf("RSI for downtrend = %v, expected < 30", lastRSI)
	}
}

func TestMACD(t *testing.T) {
	// Generate enough data points
	prices := make([]float64, 50)
	for i := range prices {
		prices[i] = 100 + float64(i)*0.5 // Uptrending
	}

	macd := CalculateMACD(prices, 12, 26, 9)
	if macd == nil {
		t.Fatal("MACD returned nil")
	}

	// For uptrend, MACD line should be positive
	if macd.MACDLine <= 0 {
		t.Errorf("MACD line = %v, expected > 0 for uptrend", macd.MACDLine)
	}
}

func TestBollingerBands(t *testing.T) {
	prices := []float64{20, 21, 22, 21, 20, 19, 20, 21, 22, 23, 22, 21, 20, 21, 22, 21, 20, 19, 20, 21}

	bb := CalculateBollingerBands(prices, 20, 2.0)
	if bb == nil {
		t.Fatal("BollingerBands returned nil")
	}

	// Upper should be > Middle > Lower
	if bb.Upper <= bb.Middle || bb.Middle <= bb.Lower {
		t.Errorf("Invalid Bollinger Bands: Upper=%v, Middle=%v, Lower=%v", bb.Upper, bb.Middle, bb.Lower)
	}

	// Width should be positive
	if bb.Width <= 0 {
		t.Errorf("Bollinger Width = %v, expected > 0", bb.Width)
	}
}

func TestStochastic(t *testing.T) {
	highs := []float64{25, 26, 27, 28, 29, 30, 29, 28, 27, 26, 25, 24, 23, 24, 25}
	lows := []float64{23, 24, 25, 26, 27, 28, 27, 26, 25, 24, 23, 22, 21, 22, 23}
	closes := []float64{24, 25, 26, 27, 28, 29, 28, 27, 26, 25, 24, 23, 22, 23, 24}

	stoch := CalculateStochastic(highs, lows, closes, 14, 3)
	if stoch == nil {
		t.Fatal("Stochastic returned nil")
	}

	// K and D should be between 0 and 100
	if stoch.K < 0 || stoch.K > 100 {
		t.Errorf("Stochastic K = %v, expected 0-100", stoch.K)
	}
	if stoch.D < 0 || stoch.D > 100 {
		t.Errorf("Stochastic D = %v, expected 0-100", stoch.D)
	}
}

func TestATR(t *testing.T) {
	highs := []float64{22, 23, 24, 25, 26, 27, 28, 27, 26, 25, 24, 23, 22, 23, 24}
	lows := []float64{20, 21, 22, 23, 24, 25, 26, 25, 24, 23, 22, 21, 20, 21, 22}
	closes := []float64{21, 22, 23, 24, 25, 26, 27, 26, 25, 24, 23, 22, 21, 22, 23}

	atr := ATR(highs, lows, closes, 14)
	if atr == nil {
		t.Fatal("ATR returned nil")
	}

	// ATR should be positive
	lastATR := atr[len(atr)-1]
	if lastATR <= 0 {
		t.Errorf("ATR = %v, expected > 0", lastATR)
	}
}
