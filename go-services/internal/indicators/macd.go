package indicators

// MACD represents MACD indicator values
type MACD struct {
	MACDLine   float64 `json:"macd_line"`
	SignalLine float64 `json:"signal_line"`
	Histogram  float64 `json:"histogram"`
}

// MACDSeries represents MACD values for entire series
type MACDSeries struct {
	MACDLine   []float64
	SignalLine []float64
	Histogram  []float64
}

// CalculateMACD calculates MACD indicator and returns the latest values
func CalculateMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) *MACD {
	if len(closes) < slowPeriod+signalPeriod {
		return nil
	}

	emaFast := EMA(closes, fastPeriod)
	emaSlow := EMA(closes, slowPeriod)

	if len(emaFast) == 0 || len(emaSlow) == 0 {
		return nil
	}

	// Calculate MACD line (Fast EMA - Slow EMA)
	macdLine := make([]float64, len(closes))
	startIdx := slowPeriod - 1 // Start from where slow EMA is valid

	for i := startIdx; i < len(closes); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}

	// Calculate Signal line (EMA of MACD line)
	validMACD := macdLine[startIdx:]
	signalEMA := EMA(validMACD, signalPeriod)

	if len(signalEMA) == 0 {
		return nil
	}

	lastMACD := macdLine[len(macdLine)-1]
	lastSignal := signalEMA[len(signalEMA)-1]

	return &MACD{
		MACDLine:   lastMACD,
		SignalLine: lastSignal,
		Histogram:  lastMACD - lastSignal,
	}
}

// CalculateMACDSeries calculates MACD for entire price series
func CalculateMACDSeries(closes []float64, fastPeriod, slowPeriod, signalPeriod int) *MACDSeries {
	if len(closes) < slowPeriod+signalPeriod {
		return nil
	}

	emaFast := EMA(closes, fastPeriod)
	emaSlow := EMA(closes, slowPeriod)

	if len(emaFast) == 0 || len(emaSlow) == 0 {
		return nil
	}

	macdLine := make([]float64, len(closes))
	startIdx := slowPeriod - 1

	for i := 0; i < startIdx; i++ {
		macdLine[i] = 0
	}

	for i := startIdx; i < len(closes); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}

	// Calculate Signal line
	validMACD := macdLine[startIdx:]
	signalEMA := EMA(validMACD, signalPeriod)

	signalLine := make([]float64, len(closes))
	histogram := make([]float64, len(closes))

	signalStartIdx := startIdx + signalPeriod - 1
	for i := 0; i < signalStartIdx; i++ {
		signalLine[i] = 0
		histogram[i] = 0
	}

	for i := 0; i < len(signalEMA); i++ {
		idx := startIdx + i
		if idx < len(signalLine) {
			signalLine[idx] = signalEMA[i]
			histogram[idx] = macdLine[idx] - signalEMA[i]
		}
	}

	return &MACDSeries{
		MACDLine:   macdLine,
		SignalLine: signalLine,
		Histogram:  histogram,
	}
}
