package Strategy

import (
	"log"
	"time"

	rt "../Retriever"
	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	aura "github.com/logrusorgru/aurora"
)

type CoinChannel struct {
	Symbol    string
	TimeFrame bfx.CandleResolution
	Freq      int
	Span      int
	Fail      error
	candles   []*bfx.Candle
	status    uint
	target    Target
	funds     float64
}

type Target struct {
	Start float64
	Out   float64
	Type  uint
}

type OrderData struct {
	Type   uint
	Price  float64
	Target float64
	MTS    float64
}

func (ch *CoinChannel) SetFunds(f float64) {
	ch.funds = f
}

func (ch *CoinChannel) Start() {
	log.Println("Comienza el scan")
	savedfreq := ch.Freq
	for {
		err := ch.scan()
		if err != nil {
			ch.Fail = err
			return
		}
		switch ch.status {
		case 0:

			if ch.isSpike() != 0 {
				ch.status = 1
				ch.Freq = 30
			}
		case 1:
			spike := ch.isSpike()
			target := ch.searchTarget(spike)
			ch.sendOrder(target)
			ch.status = 2
		case 2:
			ch.Freq = savedfreq
			if ch.waitClose() {
				ch.status = 0
			}

		}
		time.Sleep(time.Duration(60/ch.Freq) * time.Second)
	}

}

func (ch *CoinChannel) scan() error {
	client := rest.NewClient()
	//candles, err := client.Candles.History(bfx.TradingPrefix+bfx.BTCUSD, bfx.FifteenMinutes)
	now := time.Now()
	millis := now.UnixNano() / 1000000

	prior := now.Add(time.Duration(-24) * 1 * time.Hour)
	millisStart := prior.UnixNano() / 1000000
	start := bfx.Mts(millisStart)
	end := bfx.Mts(millis)

	candles, err := client.Candles.HistoryWithQuery(
		ch.Symbol,
		ch.TimeFrame,
		start,
		end,
		1000,
		bfx.OldestFirst,
	)
	if err != nil {
		log.Fatal(err.Error())

		return err
	}

	ch.candles = candles.Snapshot
	return nil

}

func (ch *CoinChannel) isSpike() uint {
	bbands := ch.getBBand()
	bband := bbands[len(bbands)-1]
	if bband.Price/bband.Upper > 1 || bband.Down/bband.Price > 1 {
		log.Printf(aura.Sprintf("Off band! p: %f, d: %f, u:%f, diff :%f , p/u : %f, d/p: %f \n", bband.Price, bband.Down, bband.Upper, aura.Blue(bband.Upper-bband.Down), aura.Colorize(bband.Price/bband.Upper, aura.RedBg), aura.Colorize(bband.Down/bband.Price, aura.GreenBg)))
	}

	if bband.Price/bband.Upper > TRIGGER {
		log.Printf("Spike! price: %f , bband: %f , short it! \n", bband.Price, bband.Upper)
		ch.status = 1
		return 1

	} else if bband.Down/bband.Price > TRIGGER {
		log.Printf("Spike! price: %f , bband: %f , long it! \n", bband.Price, bband.Down)
		ch.status = 1
		return 2
	}
	//log.Printf("No spike: p: %f, d: %f, u: %f \n", bband.Price, bband.Down, bband.Upper)
	return 0

}

func (ch *CoinChannel) searchTarget(spike uint) Target {

	return Target{
		Start: ch.candles[len(ch.candles)-1].Close,
		Out:   ch.candles[len(ch.candles)-4].Close,
		Type:  spike,
	}

}
func (ch *CoinChannel) sendOrder(t Target) {
	var tp string
	if t.Type == 1 {
		tp = "Short"
	} else {
		tp = "Long"
	}
	ch.target = t
	//log.Printf("Orden simulada, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)
	log.Printf(aura.Sprintf(aura.Green("Orden simulada, Tipo: %s, Entrada: %f, Target: %f "), tp, t.Start, t.Out))

}

func (ch *CoinChannel) waitClose() bool {
	last := ch.candles[len(ch.candles)-1].Close
	var profit float64
	if ch.target.Type == 1 {
		if ch.target.Out >= last {
			profit = ch.target.Start * 0.998 / last
			ch.funds = ch.funds * profit
			log.Printf("Orden Short cerrada, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			return true
		} else if last/ch.target.Start >= STOP {
			profit = ch.target.Start * 0.998 / last
			ch.funds = ch.funds * profit
			log.Printf("Orden Short liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			return true

		}
	} else if ch.target.Type == 2 {
		if last >= ch.target.Out {
			profit = last / ch.target.Start * 0.998
			ch.funds = ch.funds * profit
			log.Printf("Orden Long cerrada, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			return true

		} else if ch.target.Start/last >= STOP {
			profit = last / ch.target.Start * 0.998
			ch.funds = ch.funds * profit
			log.Printf("Orden Long liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			return true

		}
	}
	return false
}

/*
func (ch *CoinChannel) forecastBband(bdiff,bbsm) float64{
	bbdelta := ch.bbandDelta()
}
*/
func (ch *CoinChannel) bbandDelta() float64 {
	bbands := ch.getBBand()
	dt := len(bbands) - 1
	diff1 := bbands[dt].Upper - bbands[dt].Down
	diff2 := bbands[dt-1].Upper - bbands[dt-1].Down
	return (diff1 - diff2) / 2
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

func ema(x, em, k float64) float64 {
	k = 2 / (k + 1)
	return x*k + (1-k)*em

}
