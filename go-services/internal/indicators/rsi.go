package indicators

// RSI calculates the Relative Strength Index for the entire series
func RSI(closes []float64, period int) []float64 {
	if len(closes) < period+1 {
		return nil
	}

	result := make([]float64, len(closes))

	// Initialize with zeros for invalid periods
	for i := 0; i < period; i++ {
		result[i] = 0
	}

	var gains, losses float64

	// Calculate initial average gain/loss
	for i := 1; i <= period; i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate first RSI
	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100.0 - (100.0 / (1.0 + rs))
	}

	// Apply Wilder's smoothing for remaining periods
	for i := period + 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		var currentGain, currentLoss float64
		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		avgGain = (avgGain*float64(period-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + currentLoss) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100.0 - (100.0 / (1.0 + rs))
		}
	}

	return result
}

// RSILatest returns the most recent RSI value
func RSILatest(closes []float64, period int) float64 {
	rsi := RSI(closes, period)
	if rsi == nil || len(rsi) == 0 {
		return 0
	}
	return rsi[len(rsi)-1]
}
