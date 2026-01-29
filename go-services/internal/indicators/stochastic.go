package indicators

// Stochastic represents Stochastic Oscillator values
type Stochastic struct {
	K float64 `json:"k"`
	D float64 `json:"d"`
}

// StochasticSeries represents Stochastic values for entire series
type StochasticSeries struct {
	K []float64
	D []float64
}

// CalculateStochastic calculates Stochastic Oscillator and returns latest values
func CalculateStochastic(highs, lows, closes []float64, kPeriod, dPeriod int) *Stochastic {
	if len(closes) < kPeriod+dPeriod-1 {
		return nil
	}

	kValues := calculateRawK(highs, lows, closes, kPeriod)
	if kValues == nil {
		return nil
	}

	// Calculate %D (SMA of %K)
	dValues := SMA(kValues, dPeriod)
	if dValues == nil {
		return nil
	}

	return &Stochastic{
		K: kValues[len(kValues)-1],
		D: dValues[len(dValues)-1],
	}
}

// CalculateStochasticSeries calculates Stochastic for entire series
func CalculateStochasticSeries(highs, lows, closes []float64, kPeriod, dPeriod int) *StochasticSeries {
	if len(closes) < kPeriod+dPeriod-1 {
		return nil
	}

	kValues := calculateRawK(highs, lows, closes, kPeriod)
	if kValues == nil {
		return nil
	}

	dValues := SMA(kValues, dPeriod)
	if dValues == nil {
		return nil
	}

	return &StochasticSeries{
		K: kValues,
		D: dValues,
	}
}

// calculateRawK calculates raw %K values
func calculateRawK(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	if n < period {
		return nil
	}

	kValues := make([]float64, n)

	for i := 0; i < period-1; i++ {
		kValues[i] = 0
	}

	for i := period - 1; i < n; i++ {
		// Find highest high and lowest low in period
		highestHigh := highs[i-period+1]
		lowestLow := lows[i-period+1]

		for j := i - period + 2; j <= i; j++ {
			if highs[j] > highestHigh {
				highestHigh = highs[j]
			}
			if lows[j] < lowestLow {
				lowestLow = lows[j]
			}
		}

		// Calculate %K
		diff := highestHigh - lowestLow
		if diff == 0 {
			kValues[i] = 50 // Neutral when no range
		} else {
			kValues[i] = ((closes[i] - lowestLow) / diff) * 100
		}
	}

	return kValues
}
