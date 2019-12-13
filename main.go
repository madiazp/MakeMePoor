package main

import (
	"log"
	"time"

	rt "./Retriever"

	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
)

func main() {
	client := rest.NewClient()
	//candles, err := client.Candles.History(bfx.TradingPrefix+bfx.BTCUSD, bfx.FifteenMinutes)
	now := time.Now()
	millis := now.UnixNano() / 1000000

	prior := now.Add(time.Duration(-24) * 1 * time.Hour)
	millisStart := prior.UnixNano() / 1000000
	start := bfx.Mts(millisStart)
	end := bfx.Mts(millis)

	candles, err := client.Candles.HistoryWithQuery(
		bfx.TradingPrefix+bfx.BTCUSD,
		"15m",
		start,
		end,
		1000,
		bfx.OldestFirst,
	)
	if err != nil {
		log.Println(err.Error())
	}
	p := 14
	macd, t := rt.MACD(candles.Snapshot, p)
	rt.MACDPlotter(macd, t)

}
