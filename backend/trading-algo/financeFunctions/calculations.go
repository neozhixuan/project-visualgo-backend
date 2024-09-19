package financeFunctions

type Candlestick struct {
	Open, High, Low, Close float64
	Volume                 float64
}

// CalculateEMA calculates the Exponential Moving Average (EMA)
func CalculateEMA(candlesticks []Candlestick, period int) []float64 {
	k := 2 / float64(period+1)
	prices := extractClosingPrices(candlesticks)
	ema := make([]float64, len(prices))
	ema[0] = prices[0]

	for i := 1; i < len(prices); i++ {
		ema[i] = prices[i]*k + ema[i-1]*(1-k)
	}
	return ema
}

// CalculateVWAP calculates the Volume Weighted Average Price (VWAP)
func CalculateVWAP(candlesticks []Candlestick) []float64 {
	vwap := make([]float64, len(candlesticks))
	cumulativePriceVolume := 0.0
	cumulativeVolume := 0.0

	for i, candle := range candlesticks {
		typicalPrice := (candle.High + candle.Low + candle.Close) / 3
		priceVolume := typicalPrice * candle.Volume
		cumulativePriceVolume += priceVolume
		cumulativeVolume += candle.Volume
		vwap[i] = cumulativePriceVolume / cumulativeVolume
	}
	return vwap
}

// Helper function to extract closing prices from candlesticks
func extractClosingPrices(candlesticks []Candlestick) []float64 {
	prices := make([]float64, len(candlesticks))
	for i, candle := range candlesticks {
		prices[i] = candle.Close
	}
	return prices
}
