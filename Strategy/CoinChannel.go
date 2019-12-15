package Strategy

import (
	"log"
	"time"

	rt "../Retriever"
	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
)

type CoinChannel struct {
	Symbol    string
	TimeFrame bfx.CandleResolution
	Freq      int
	Span      int
	candles   []*bfx.Candle
}

func (ch *CoinChannel) Start() error {
	client := rest.NewClient()
	//candles, err := client.Candles.History(bfx.TradingPrefix+bfx.BTCUSD, bfx.FifteenMinutes)
	now := time.Now()
	millis := now.UnixNano() / 1000000

	prior := now.Add(time.Duration(-24) * 1 * time.Hour)
	millisStart := prior.UnixNano() / 1000000
	start := bfx.Mts(millisStart)
	end := bfx.Mts(millis)
	for {

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
		time.Sleep(time.Duration(60/ch.Freq) * time.Second)

	}

}

func (ch *CoinChannel) GetMACD() (macd, t []float64) {
	return rt.MACD(ch.candles, ch.Span)

}

func (ch *CoinChannel) GetRSI() (rsi, t []float64) {
	return rt.RSI(ch.candles, ch.Span)

}

func (ch *CoinChannel) GetBBand() (central, upper, down, data, t []float64) {
	return rt.BBand(ch.candles, ch.Span)

}
