package Indicators

import (
	bitfinexCandle "github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
	"gonum.org/v1/gonum/stat"
)

type BollingerBandData struct {
	High      float64
	Low       float64
	Mid       float64
	Price     float64
	LastClose float64
	Central   float64
	Upper     float64
	Lower     float64
	Timestamp int64
}

// BollingerBand computes bollinger band data struct from a set of candles
// (see more on the wiki https://github.com/madiazp/MakeMePoor/wiki).
// Apparently a set of candles is called a Snapshot, but bitfinex api docs are ass.
// TODO: evaluate whether or not to use Snapshot instead of Slice
func BollingerBand(candles []*bitfinexCandle.Candle, span int, sensitivity float64) (bollingerBand BollingerBandData) {
	var mean, std float64
	var data []float64

	for _, candle := range candles {
		data = append(data, candle.Close)
		// wait for enough data
	}
	dataLen := len(data) - 1
	candleLen := len(candles) - 1
	init := dataLen - span
	if init < 0 {
		init = 0
	}
	mean, std = stat.MeanStdDev(data[init:dataLen], nil)
	return BollingerBandData{
		High:      candles[candleLen].High,
		Low:       candles[candleLen].Low,
		Mid:       (candles[candleLen].High + candles[candleLen].Low) / 2,
		Price:     data[dataLen],
		LastClose: data[dataLen-1],
		Central:   mean,
		Upper:     mean + sensitivity*std,
		Lower:     mean - sensitivity*std,
		Timestamp: candles[candleLen].MTS,
	}
}

// RelativeStrengthIndex computes this indicator from a set of candles.
// Entry not yet on the wiki.
func RelativeStrengthIndex(candles []*bitfinexCandle.Candle, span int) (rsiValues []float64, timestamps []int64) {
	// initialization
	previousClose := candles[0].Close
	// first smma evaluated with the first candle (to avoid div 0)
	downwardSMMA := candles[0].Close
	upwardSMMA := candles[0].Close
	var downwardChange, upwardChange float64

	for _, candle := range candles[1:] {
		timestamps = append(timestamps, candle.MTS)

		if previousClose > candle.Close {
			// Loss
			downwardChange = previousClose - candle.Close
			upwardChange = 0
		} else {
			// Gain
			downwardChange = 0
			upwardChange = candle.Close - previousClose

		}

		downwardSMMA = smma(downwardSMMA, downwardChange, span)
		upwardSMMA = smma(upwardSMMA, upwardChange, span)

		rsiValues = append(rsiValues, 100-100/(1+upwardSMMA/downwardSMMA))
		previousClose = candle.Close
	}
	return rsiValues, timestamps
}

// MovingAverageConvergenceDivergence computes this indicator from a set of candles.
// Entry not yet on the wiki.
func MovingAverageConvergenceDivergence(candles []*bitfinexCandle.Candle) (macd float64, macdHistogram []float64,
	timestamps []int64) {

	longTermEMA := candles[0].Close
	shortTermEMA := candles[0].Close
	var signal float64

	for _, candle := range candles[1:] {
		timestamps = append(timestamps, candle.MTS)

		longTermEMA = Ema(candle.Close, longTermEMA, 26)
		shortTermEMA = Ema(candle.Close, shortTermEMA, 12)
		macd = shortTermEMA - longTermEMA

		signal = Ema(macd, signal, 9)
		macdHistogram = append(macdHistogram, macd-signal)
	}
	return macd, macdHistogram, timestamps
}

// watch out for trending

func GetEMA(candles []*bitfinexCandle.Candle, k int) (ema50 []float64) {
	ema50 = append(ema50, candles[0].Close)
	for i, cdls := range candles {
		if i > 0 {
			ema50 = append(ema50, Ema(cdls.Close, ema50[i-1], k))
		}
	}
	return ema50
}

// smma computes the smoothed moving average from the previous value and the upward/downward variation.
func smma(previousSMMA, variation float64, periods int) float64 {
	return (float64(periods-1)*previousSMMA + variation) / float64(periods)
}

// ema computes the exponential moving average from the candle close and the previous value.
func Ema(closeValue, previousEMA float64, periods int) float64 {
	alpha := 2 / float64(periods+1)
	return alpha*closeValue + (1-alpha)*previousEMA
}
