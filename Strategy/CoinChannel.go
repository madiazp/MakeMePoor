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
	Symbol  string
	Freq    int
	Span    int
	Fail    error
	candles []*bfx.Candle
	status  uint
	target  Target
	funds   float64
}

type Target struct {
	Start float64
	Out   float64
	Type  uint
	MTS   float64
}
type EmaData struct {
	Ema float64
	MTS float64
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

func (ch *CoinChannel) scan(TimeFrame bfx.CandleResolution) error {
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
		TimeFrame,
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

func (ch *CoinChannel) bbandDelta() float64 {
	bbands := ch.getBBand()
	i := len(bbands) - 1
	diff1 := bbands[i].Upper - bbands[i].Down
	diff2 := bbands[i-1].Upper - bbands[i-1].Down
	delta := (diff1 - diff2) / 2

	return delta
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

func (ch *CoinChannel) getEMA50() (ema50 []EmaData) {
	for i, cdls := range ch.candles {
		if i != 0 {
			ema50 = append(ema50, EmaData{
				Ema: ema(cdls.Close, ema50[i-1].Ema, 20),
				MTS: float64(cdls.MTS),
			})
		}
	}
	return ema50
}

func ema(x, em, k float64) float64 {
	k = 2 / (k + 1)
	return x*k + (1-k)*em

}
