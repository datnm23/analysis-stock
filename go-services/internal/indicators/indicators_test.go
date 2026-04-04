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
	// kPeriod=14, dPeriod=3 → requires at least kPeriod+dPeriod-1 = 16 points
	highs := []float64{25, 26, 27, 28, 29, 30, 29, 28, 27, 26, 25, 24, 23, 24, 25, 26}
	lows := []float64{23, 24, 25, 26, 27, 28, 27, 26, 25, 24, 23, 22, 21, 22, 23, 24}
	closes := []float64{24, 25, 26, 27, 28, 29, 28, 27, 26, 25, 24, 23, 22, 23, 24, 25}

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

func TestADX(t *testing.T) {
	// Use enough data points: period*2 minimum
	highs := []float64{25, 26, 27, 28, 29, 30, 31, 30, 29, 28, 27, 26, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}
	lows := []float64{23, 24, 25, 26, 27, 28, 29, 28, 27, 26, 25, 24, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38}
	closes := []float64{24, 25, 26, 27, 28, 29, 30, 29, 28, 27, 26, 25, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39}

	adx := CalculateADX(highs, lows, closes, 14)
	if adx == nil {
		t.Fatal("ADX returned nil")
	}

	// ADX should be between 0 and 100
	if adx.ADX < 0 || adx.ADX > 100 {
		t.Errorf("ADX = %v, expected 0-100", adx.ADX)
	}

	// +DI and -DI should be non-negative
	if adx.PlusDI < 0 {
		t.Errorf("+DI = %v, expected >= 0", adx.PlusDI)
	}
	if adx.MinusDI < 0 {
		t.Errorf("-DI = %v, expected >= 0", adx.MinusDI)
	}

	// For strong uptrend, +DI should be > -DI
	if adx.PlusDI <= adx.MinusDI {
		t.Errorf("+DI (%v) should be > -DI (%v) for uptrend", adx.PlusDI, adx.MinusDI)
	}
}

func TestADX_InsufficientData(t *testing.T) {
	highs := []float64{25, 26, 27}
	lows := []float64{23, 24, 25}
	closes := []float64{24, 25, 26}

	adx := CalculateADX(highs, lows, closes, 14)
	if adx != nil {
		t.Error("ADX should return nil for insufficient data")
	}
}

func TestVWAP(t *testing.T) {
	highs := []float64{22, 23, 24, 25}
	lows := []float64{20, 21, 22, 23}
	closes := []float64{21, 22, 23, 24}
	volumes := []int64{1000, 2000, 1500, 1000}

	vwap := VWAP(highs, lows, closes, volumes)
	if vwap == nil {
		t.Fatal("VWAP returned nil")
	}
	if len(vwap) != 4 {
		t.Fatalf("VWAP length = %v, expected 4", len(vwap))
	}

	// First VWAP: typical price = (22+20+21)/3 = 21.0, volume = 1000
	expectedFirst := 21.0
	if !almostEqual(vwap[0], expectedFirst) {
		t.Errorf("VWAP[0] = %v, expected %v", vwap[0], expectedFirst)
	}

	// VWAP should be positive for all bars
	for i, v := range vwap {
		if v <= 0 {
			t.Errorf("VWAP[%d] = %v, expected > 0", i, v)
		}
	}
}

func TestVWAP_MismatchedLengths(t *testing.T) {
	highs := []float64{22, 23}
	lows := []float64{20, 21, 22} // different length
	closes := []float64{21, 22}
	volumes := []int64{1000, 2000}

	vwap := VWAP(highs, lows, closes, volumes)
	if vwap != nil {
		t.Error("VWAP should return nil for mismatched lengths")
	}
}

func TestVWAPLatest(t *testing.T) {
	highs := []float64{22, 23, 24}
	lows := []float64{20, 21, 22}
	closes := []float64{21, 22, 23}
	volumes := []int64{1000, 2000, 1500}

	latest := VWAPLatest(highs, lows, closes, volumes)
	if latest <= 0 {
		t.Errorf("VWAPLatest = %v, expected > 0", latest)
	}
}

func TestRSI_AllGains(t *testing.T) {
	// Purely uptrending: avgLoss == 0, RSI should be 100
	prices := make([]float64, 20)
	for i := range prices {
		prices[i] = float64(i + 1)
	}

	rsi := RSI(prices, 14)
	if rsi == nil {
		t.Fatal("RSI returned nil")
	}

	lastRSI := rsi[len(rsi)-1]
	if !almostEqual(lastRSI, 100.0) {
		t.Errorf("RSI for all-gains = %v, expected 100", lastRSI)
	}
}

func TestRSI_AllLosses(t *testing.T) {
	// Purely downtrending: avgGain == 0, RSI should be 0
	prices := make([]float64, 20)
	for i := range prices {
		prices[i] = float64(20 - i)
	}

	rsi := RSI(prices, 14)
	if rsi == nil {
		t.Fatal("RSI returned nil")
	}

	lastRSI := rsi[len(rsi)-1]
	if !almostEqual(lastRSI, 0.0) {
		t.Errorf("RSI for all-losses = %v, expected 0", lastRSI)
	}
}

func TestBollingerBands_SampleStdDev(t *testing.T) {
	// Verify sample stddev is used: for 2 values [0, 2], sample stddev = sqrt(2) ≈ 1.414
	// Use period=2, prices=[0, 2]: middle=1, stddev=sqrt((0-1)^2+(2-1)^2 / (2-1)) = sqrt(2)
	prices := []float64{0, 2}
	bb := CalculateBollingerBands(prices, 2, 1.0)
	if bb == nil {
		t.Fatal("BollingerBands returned nil")
	}

	expectedStdDev := math.Sqrt(2.0) // sample stddev
	expectedUpper := 1.0 + expectedStdDev
	if !almostEqual(bb.Upper, expectedUpper) {
		t.Errorf("Upper band = %v, expected %v (verifying sample stddev)", bb.Upper, expectedUpper)
	}
}

func TestSMA_InsufficientData(t *testing.T) {
	prices := []float64{10, 11}
	sma := SMA(prices, 5)
	if sma != nil {
		t.Error("SMA should return nil for insufficient data")
	}
}

func TestEMA_InsufficientData(t *testing.T) {
	prices := []float64{10, 11}
	ema := EMA(prices, 5)
	if ema != nil {
		t.Error("EMA should return nil for insufficient data")
	}
}
