package indicators

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	result := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// Initialize with zeros for invalid periods
	for i := 0; i < period-1; i++ {
		result[i] = 0
	}

	// First EMA is SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate EMA for remaining values
	for i := period; i < len(prices); i++ {
		result[i] = (prices[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}

// EMALatest returns the most recent EMA value
func EMALatest(prices []float64, period int) float64 {
	ema := EMA(prices, period)
	if ema == nil || len(ema) == 0 {
		return 0
	}
	return ema[len(ema)-1]
}
