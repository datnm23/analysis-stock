package indicators

import "math"

// ATR calculates Average True Range
func ATR(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	if n < period+1 {
		return nil
	}

	// Calculate True Range
	tr := make([]float64, n)
	tr[0] = highs[0] - lows[0]

	for i := 1; i < n; i++ {
		highLow := highs[i] - lows[i]
		highPrevClose := math.Abs(highs[i] - closes[i-1])
		lowPrevClose := math.Abs(lows[i] - closes[i-1])
		tr[i] = math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
	}

	// Calculate ATR using Wilder's smoothing (similar to EMA)
	atr := make([]float64, n)

	// First ATR is simple average
	var sum float64
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	atr[period-1] = sum / float64(period)

	// Apply Wilder's smoothing
	multiplier := 1.0 / float64(period)
	for i := period; i < n; i++ {
		atr[i] = (atr[i-1] * float64(period-1) + tr[i]) * multiplier
	}

	return atr
}

// ATRLatest returns the most recent ATR value
func ATRLatest(highs, lows, closes []float64, period int) float64 {
	atr := ATR(highs, lows, closes, period)
	if atr == nil || len(atr) == 0 {
		return 0
	}
	return atr[len(atr)-1]
}
