package indicators

// VWAP calculates Volume Weighted Average Price
func VWAP(highs, lows, closes []float64, volumes []int64) []float64 {
	n := len(closes)
	if n == 0 || len(highs) != n || len(lows) != n || len(volumes) != n {
		return nil
	}

	vwap := make([]float64, n)
	var cumulativeTPV float64 // Cumulative Typical Price * Volume
	var cumulativeVolume float64

	for i := 0; i < n; i++ {
		// Typical Price = (High + Low + Close) / 3
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3

		cumulativeTPV += typicalPrice * float64(volumes[i])
		cumulativeVolume += float64(volumes[i])

		if cumulativeVolume > 0 {
			vwap[i] = cumulativeTPV / cumulativeVolume
		}
	}

	return vwap
}

// VWAPLatest returns the most recent VWAP value
func VWAPLatest(highs, lows, closes []float64, volumes []int64) float64 {
	vwap := VWAP(highs, lows, closes, volumes)
	if vwap == nil || len(vwap) == 0 {
		return 0
	}
	return vwap[len(vwap)-1]
}
