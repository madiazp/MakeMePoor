package Strategy

import (
	"log"
	"time"

	rt "../Retriever"
	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	aura "github.com/logrusorgru/aurora"
)

// CoinChannel is the main struct, it hold the inner state of a coin
type CoinChannel struct {
	Symbol  string        // the symbol of the coin
	Freq    int           // the recuency of the scan
	Span    int           // the rate of the bollinger bands, typically 20
	Fail    error         // inner state of error, logged when simulated
	candles []*bfx.Candle // candles
	status  uint          // state of the control state machine
	trend   uint          // trend indicator, 0 if tunneling, 1 for down-trend, 2 to up-trend
	target  Target        // price action
	funds   float64       // total fund
    params  Params        // triggers and options
}

// order state
type Target struct {
	Start  float64 // the start price of the position
	Out    float64 // the target price
	Type   uint    // 1 for short, 2 for long
	MTS    float64 // MTS of the start
	Active bool    //if the order start and is waiting for the end.
}

// order struct to API, unused in simulation
type OrderData struct {
	Type   uint
	Price  float64
	Target float64
	MTS    float64
}

func (ch *CoinChannel) SetFunds(f float64) {
	ch.funds = f
}

// scan the candles
func (ch *CoinChannel) scan(TimeFrame bfx.CandleResolution) error {

	client := rest.NewClient()
	now := time.Now()
	millis := now.UnixNano() / 1000000

	prior := now.Add(time.Duration(-24) * 1 * time.Hour)
	millisStart := prior.UnixNano() / 1000000
	start := bfx.Mts(millisStart)
	end := bfx.Mts(millis)

	candles, err := client.Candles.HistoryWithQuery(
		ch.Symbol,
		TimeFrame,
		start,
		end,
		1000,
		bfx.OldestFirst,
	)
	if err != nil {

		return err
	}
	ch.candles = candles.Snapshot
	ch.scanTrend()
	return nil

}

// start an order
func (ch *CoinChannel) sendOrder(t Target) {
    if !ch.params.IsInit {
        ch.params.Init()
    }
	var tp string
	if t.Type == 1 {
		tp = "Short"
		if ch.params.Fees >= t.Start/t.Out {
			ch.status = 0
			ch.target = Target{}
			log.Printf(aura.Sprintf(aura.Green("ROI muy bajo!, Tipo: %s, Entrada: %f, Target: %f "), tp, t.Start, t.Out))
			return
		}
	} else {
		tp = "Long"
		if ch.params.Fees >= t.Out/t.Start {
			ch.status = 0
			ch.target = Target{}
			log.Printf(aura.Sprintf(aura.Green("ROI muy bajo!, Tipo: %s, Entrada: %f, Target: %f "), tp, t.Start, t.Out))
			return
		}
	}
	ch.target = t
	ch.status = 2

	log.Printf(aura.Sprintf(aura.BgBrightRed("Orden simulada, Tipo: %s, Entrada: %f, Target: %f "), tp, t.Start, t.Out))

}

func (ch *CoinChannel) bbandDelta() float64 {
	bbands := ch.getBBand()
	i := len(bbands) - 1
	diff1 := bbands[i].Upper - bbands[i].Down
	diff2 := bbands[i-1].Upper - bbands[i-1].Down
	delta := (diff1 - diff2) / 2

	return delta
}

// watch out for trending
func (ch *CoinChannel) scanTrend() {
    if !ch.params.IsInit {
        ch.params.Init()
    }
	ema50 := ch.getEMA(50)
	ema20 := ch.getEMA(20)
	lastEma50 := ema50[len(ema50)-1]
	lastEma20 := ema20[len(ema20)-1]
	lastPrice := ch.candles[len(ch.candles)-1].Close
	lastHigh := ch.candles[len(ch.candles)-1].High
	lastLow := ch.candles[len(ch.candles)-1].Low
	// if the price is trending look for the price to touch the ema5 again to end it.
	if ch.trend != 0 {
		if ch.trend == 1 && lastPrice >= lastEma50 {
			log.Printf(aura.Sprintf(aura.BrightCyan(" down trend is over, price: %f, ema: %f, Symbol: %s ! \n"), lastPrice, lastEma50, ch.Symbol))
			ch.trend = 0
			ch.status = 0
		} else if ch.trend == 2 && lastEma50 >= lastPrice {
			log.Printf(aura.Sprintf(aura.BrightCyan(" up trend is over, price: %f, ema: %f, Symbol: %s ! \n"), lastPrice, lastEma50, ch.Symbol))
			ch.trend = 0
			ch.status = 0

		}
		// use the ema50/ema20 ratio , if the ratio is more or less than the trigger ratio and the price is over/below the ema50 then the trend is delcared
	} else if lastEma50/lastEma20 >= ch.params.TrendTrigger && ch.trend != 1 && lastEma50 >= lastHigh {
		log.Printf(aura.Sprintf(aura.BrightCyan(" down trend start, rate: %f, ema20:%f, ema20:%f, symbol: %s ! \n"), lastEma50/lastEma20, lastEma20, lastEma50, ch.Symbol))
		ch.trend = 1

	} else if lastEma20/lastEma50 >= ch.params.TrendTrigger && ch.trend != 2 && lastLow >= lastEma50 {
		log.Printf(aura.Sprintf(aura.BrightCyan(" up trend start, rate: %f , ema20: %f, ema50:%f, symbol: %s \n"), lastEma20/lastEma50, lastEma20, lastEma50, ch.Symbol))
		ch.trend = 2
	}
}

func (ch *CoinChannel) getMACD() (macd, t []float64) {
	return rt.MACD(ch.candles, ch.Span)

}

func (ch *CoinChannel) getRSI() (rsi, t []float64) {
	return rt.RSI(ch.candles, ch.Span)

}

func (ch *CoinChannel) getBBand() (bband []rt.BBData) {
	return rt.BBand(ch.candles, ch.Span)

}

// get the ema of k periods of the candles
func (ch *CoinChannel) getEMA(k float64) (ema50 []float64) {
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
