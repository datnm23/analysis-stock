package indicators

import "math"

// ADX represents Average Directional Index values
type ADX struct {
	ADX    float64 `json:"adx"`
	PlusDI float64 `json:"plus_di"`
	MinusDI float64 `json:"minus_di"`
}

// CalculateADX calculates ADX and returns latest values
func CalculateADX(highs, lows, closes []float64, period int) *ADX {
	n := len(closes)
	if n < period*2 {
		return nil
	}

	// Calculate True Range, +DM, -DM
	tr := make([]float64, n)
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)

	for i := 1; i < n; i++ {
		// True Range
		highLow := highs[i] - lows[i]
		highPrevClose := math.Abs(highs[i] - closes[i-1])
		lowPrevClose := math.Abs(lows[i] - closes[i-1])
		tr[i] = math.Max(highLow, math.Max(highPrevClose, lowPrevClose))

		// Directional Movement
		upMove := highs[i] - highs[i-1]
		downMove := lows[i-1] - lows[i]

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}
	}

	// Smooth TR, +DM, -DM using Wilder's smoothing
	smoothTR := wilderSmooth(tr, period)
	smoothPlusDM := wilderSmooth(plusDM, period)
	smoothMinusDM := wilderSmooth(minusDM, period)

	// Calculate +DI and -DI
	plusDI := make([]float64, n)
	minusDI := make([]float64, n)
	dx := make([]float64, n)

	for i := period; i < n; i++ {
		if smoothTR[i] != 0 {
			plusDI[i] = (smoothPlusDM[i] / smoothTR[i]) * 100
			minusDI[i] = (smoothMinusDM[i] / smoothTR[i]) * 100
		}

		diSum := plusDI[i] + minusDI[i]
		if diSum != 0 {
			dx[i] = (math.Abs(plusDI[i]-minusDI[i]) / diSum) * 100
		}
	}

	// Calculate ADX (smoothed DX)
	adxValues := wilderSmooth(dx, period)

	return &ADX{
		ADX:     adxValues[len(adxValues)-1],
		PlusDI:  plusDI[len(plusDI)-1],
		MinusDI: minusDI[len(minusDI)-1],
	}
}

// wilderSmooth applies Wilder's smoothing method
func wilderSmooth(values []float64, period int) []float64 {
	n := len(values)
	if n < period {
		return nil
	}

	result := make([]float64, n)

	// First value is sum of first period values
	var sum float64
	for i := 1; i <= period; i++ {
		sum += values[i]
	}
	result[period] = sum

	// Apply smoothing
	for i := period + 1; i < n; i++ {
		result[i] = result[i-1] - (result[i-1] / float64(period)) + values[i]
	}

	return result
}
