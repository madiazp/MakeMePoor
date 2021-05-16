package Engine

import (
	bitfinexCommon "github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"log"

    indicator "github.com/madiazp/MakeMePoor/Indicators"
    comms "github.com/madiazp/MakeMePoor/Comms"

)

// Engine methods
// scan the candles
func (ch *CoinChannel) Scan(TimeFrame bitfinexCommon.CandleResolution) error {
    var err error
	ch.candles, err = comms.Scan(TimeFrame, ch.Symbol)
    if err != nil {
      return err
    }
	ch.scanTrend()
	return nil

}

// start an order
func (ch *CoinChannel) SendOrder(t Target) {
	if !ch.Params.IsInit {
		ch.Params.Init()
	}
	var tp string
	if t.Type == 1 {
		tp = "Short"
		if ch.Params.Fees >= t.Start/t.Out {
			ch.status = 0
			ch.target = Target{}
			log.Printf("ROI muy bajo!, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)
			return
		}
	} else {
		tp = "Long"
		if ch.Params.Fees >= t.Out/t.Start {
			ch.status = 0
			ch.target = Target{}
			log.Printf("ROI muy bajo!, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)
			return
		}
	}
	ch.target = t
	ch.status = 2

	log.Printf("Orden simulada, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)

}

func (ch *CoinChannel) BbandDelta() float64 {
	bbands := ch.GetBBand()
	i := len(bbands) - 1
	diff1 := bbands[i].Upper - bbands[i].Lower
	diff2 := bbands[i-1].Upper - bbands[i-1].Lower
	delta := (diff1 - diff2) / 2

	return delta
}

// watch out for trending
func (ch *CoinChannel) scanTrend() {
	if !ch.Params.IsInit {
		ch.Params.Init()
	}
	ema50 := ch.GetEMA(50)
	ema20 := ch.GetEMA(20)
	lastEma50 := ema50[len(ema50)-1]
	lastEma20 := ema20[len(ema20)-1]
	lastPrice := ch.candles[len(ch.candles)-1].Close
	lastHigh := ch.candles[len(ch.candles)-1].High
	lastLow := ch.candles[len(ch.candles)-1].Low
	// if the price is trending look for the price to touch the ema5 again to end it.
	if ch.trend != 0 {
		if ch.trend == 1 && lastPrice >= lastEma50 {
			log.Printf(" down trend is over, price: %f, ema: %f, Symbol: %s ! \n", lastPrice, lastEma50, ch.Symbol)
			ch.trend = 0
			ch.status = 0
		} else if ch.trend == 2 && lastEma50 >= lastPrice {
			log.Printf(" up trend is over, price: %f, ema: %f, Symbol: %s ! \n", lastPrice, lastEma50, ch.Symbol)
			ch.trend = 0
			ch.status = 0

		}
		// use the ema50/ema20 ratio , if the ratio is more or less than the trigger ratio and the price is over/below the ema50 then the trend is delcared
	} else if lastEma50/lastEma20 >= ch.Params.TrendTrigger && ch.trend != 1 && lastEma50 >= lastHigh {
		log.Printf(" down trend start, rate: %f, ema20:%f, ema20:%f, symbol: %s ! \n", lastEma50/lastEma20, lastEma20, lastEma50, ch.Symbol)
		ch.trend = 1

	} else if lastEma20/lastEma50 >= ch.Params.TrendTrigger && ch.trend != 2 && lastLow >= lastEma50 {
		log.Printf(" up trend start, rate: %f , ema20: %f, ema50:%f, symbol: %s \n", lastEma20/lastEma50, lastEma20, lastEma50, ch.Symbol)
		ch.trend = 2
	}
}

func (ch *CoinChannel) GetMACD() (macd []float64, t []int64) {
	return indicator.MovingAverageConvergenceDivergence(ch.candles)

}

func (ch *CoinChannel) GetRSI() (rsi []float64, t []int64) {
	return indicator.RelativeStrengthIndex(ch.candles, ch.Span)

}

func (ch *CoinChannel) GetBBand() (bband []indicator.BollingerBandData) {
	return indicator.BollingerBand(ch.candles, ch.Span)

}

// get the ema of k periods of the candles
func (ch *CoinChannel) GetEMA(k float64) (ema50 []float64) {
	ema50 = append(ema50, ch.candles[0].Close)
	for i, cdls := range ch.candles {
		if i > 0 {
			ema50 = append(ema50, ema(cdls.Close, ema50[i-1], k))
		}
	}
	return ema50
}
// exponential mov avarage
func ema(x, em, k float64) float64 {
	k = 2 / (k + 1)
	return x*k + (1-k)*em
}
