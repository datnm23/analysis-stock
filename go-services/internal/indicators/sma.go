package indicators

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	result := make([]float64, len(prices))

	// Initialize with NaN-like behavior (0 for invalid periods)
	for i := 0; i < period-1; i++ {
		result[i] = 0
	}

	// Calculate first SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate remaining SMAs using sliding window
	for i := period; i < len(prices); i++ {
		sum = sum - prices[i-period] + prices[i]
		result[i] = sum / float64(period)
	}

	return result
}

// SMALatest returns the most recent SMA value
func SMALatest(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	var sum float64
	startIdx := len(prices) - period
	for i := startIdx; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}
