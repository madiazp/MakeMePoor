package Strategy

import (
	"log"
	"time"

	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
)

func Simulate(funds float64, freq, span int) {

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
	go ccbtc.Start15M()
	go cceth.Start15M()
	go ccxrp.Start15M()
	for {
		if ccbtc.Fail != nil {
			log.Fatal(ccbtc.Fail)
		}
		if cceth.Fail != nil {
			log.Fatal(ccbtc.Fail)
		}
		if ccxrp.Fail != nil {
			log.Fatal(ccbtc.Fail)
		}
		time.Sleep(2 * time.Second)
	}

}
