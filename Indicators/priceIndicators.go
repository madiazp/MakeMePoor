package Indicators

import (
	"math"

	bitfinexCandle "github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
)

// AvarageTrueRange computes this indicator from a set of candles.
// Entry not yet on the wiki.
func AvarageTrueRange(candles []*bitfinexCandle.Candle) (atr float64) {

	lastValue := candles[0].Close
	accumulator := 0.0

	for i, candle := range candles[1:] {
		dayRange := candle.High - candle.Low
		highRange := math.Abs(candle.High - lastValue)
		lowRange := math.Abs(candle.Low - lastValue)
		trueRange := math.Max(dayRange, highRange)
		trueRange = math.Max(trueRange, lowRange)
		accumulator += trueRange
		atr = accumulator / float64(i+1)
		lastValue = candle.Close
	}
	return atr
}
