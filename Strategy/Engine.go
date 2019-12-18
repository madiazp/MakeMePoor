package Strategy

import (
	"log"
	"time"

	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
)

// start a simulation without real money
func Simulate(funds float64, freq, span int) {

	// use 3 coin channels: btc(bitcoin), eth(ethereum), xrp(ripple)
	ccbtc := CoinChannel{
		Symbol: bfx.TradingPrefix + bfx.BTCUSD,
		Freq:   freq,
		Span:   span,
	}
	cceth := CoinChannel{
		Symbol: bfx.TradingPrefix + bfx.ETHUSD,
		Freq:   freq,
		Span:   span,
	}
	ccxrp := CoinChannel{
		Symbol: bfx.TradingPrefix + bfx.XRPUSD,
		Freq:   freq,
		Span:   span,
	}
	ccbtc.SetFunds(funds)
	cceth.SetFunds(funds)
	ccxrp.SetFunds(funds)
	// use the 15 minutes candle strategy
	go ccbtc.Start15M()
	go cceth.Start15M()
	go ccxrp.Start15M()
	// for simulation only logs the errors
	for {
		if ccbtc.Fail != nil {
			log.Println("Error BTC: " + ccbtc.Fail.Error())
			ccbtc.Fail = nil
		}
		if cceth.Fail != nil {
			log.Fatal("Error ETH: " + cceth.Fail.Error())
			cceth.Fail = nil
		}
		if ccxrp.Fail != nil {
			log.Fatal("Error XRP: " + ccxrp.Fail.Error())
			ccxrp.Fail = nil
		}
		time.Sleep(2 * time.Second)
	}

}
