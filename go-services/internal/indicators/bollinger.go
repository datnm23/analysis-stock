package indicators

import "math"

// BollingerBands represents Bollinger Bands values
type BollingerBands struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
	Width  float64 `json:"width"`
}

// BollingerBandsSeries represents Bollinger Bands for entire series
type BollingerBandsSeries struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
	Width  []float64
}

// CalculateBollingerBands calculates Bollinger Bands and returns the latest values
func CalculateBollingerBands(closes []float64, period int, stdDevMultiplier float64) *BollingerBands {
	if len(closes) < period {
		return nil
	}

	// Calculate SMA (middle band)
	sma := SMA(closes, period)
	if len(sma) == 0 {
		return nil
	}
	middle := sma[len(sma)-1]

	// Calculate standard deviation for recent period
	recentPrices := closes[len(closes)-period:]
	var sum float64
	for _, price := range recentPrices {
		diff := price - middle
		sum += diff * diff
	}
	stdDev := math.Sqrt(sum / float64(period))

	upper := middle + (stdDevMultiplier * stdDev)
	lower := middle - (stdDevMultiplier * stdDev)

	var width float64
	if middle != 0 {
		width = (upper - lower) / middle
	}

	return &BollingerBands{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
		Width:  width,
	}
}

// CalculateBollingerBandsSeries calculates Bollinger Bands for entire series
func CalculateBollingerBandsSeries(closes []float64, period int, stdDevMultiplier float64) *BollingerBandsSeries {
	if len(closes) < period {
		return nil
	}

	sma := SMA(closes, period)
	if len(sma) == 0 {
		return nil
	}

	upper := make([]float64, len(closes))
	lower := make([]float64, len(closes))
	width := make([]float64, len(closes))

	for i := 0; i < period-1; i++ {
		upper[i] = 0
		lower[i] = 0
		width[i] = 0
	}

	for i := period - 1; i < len(closes); i++ {
		middle := sma[i]

		// Calculate standard deviation
		startIdx := i - period + 1
		var sum float64
		for j := startIdx; j <= i; j++ {
			diff := closes[j] - middle
			sum += diff * diff
		}
		stdDev := math.Sqrt(sum / float64(period))

		upper[i] = middle + (stdDevMultiplier * stdDev)
		lower[i] = middle - (stdDevMultiplier * stdDev)

		if middle != 0 {
			width[i] = (upper[i] - lower[i]) / middle
		}
	}

	return &BollingerBandsSeries{
		Upper:  upper,
		Middle: sma,
		Lower:  lower,
		Width:  width,
	}
}
