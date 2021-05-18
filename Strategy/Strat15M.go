package Strategy

import (
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"log"
	"time"
    engine "github.com/madiazp/MakeMePoor/Engine"

)

const (
    SHORT = 1
    LONG = 2
)

// Start15M : main state machine of the strategy
func Start15M(ch *engine.CoinChannel) {
    ch.InitParams()
	log.Println("Scan starting")
	savedfreq := ch.Freq // hold the freq
	detect := false
	var spike uint
	for {
		err := ch.Scan(common.FifteenMinutes) // use 15 min candles
		if err != nil {
			ch.Fail = err
		}
		// if trending stop everything
		if ch.IsTrend() {
			ch.SetStatus(3)
		}

		switch ch.Status() {
		// first state, look up for spike out of the bollinger bands
		case 0:
			spike, detect = isSpike(ch, detect)
			if spike != 0 {
				ch.SetStatus(1)
				ch.Freq = 30
			}
		// if a spike is found then make an order and hold a position
		case 1:
			spike, detect = isSpike(ch, detect)
			target := searchTarget(ch, spike)
			ch.SendOrder(target)
		// wait for the price to reach the target and then reset
		case 2:
			ch.Freq = savedfreq
			if waitClose(ch) {
				ch.SetStatus(0)
			}
		// if the price is trending, liquidate every position and wait to stop trending
		case 3:
			ch.Freq = savedfreq
			if ch.TargetReached() {
				liquidateOrders(ch)
			}

		}
		// the max rate of the bitfinex api is 60 request per minute
		time.Sleep(time.Duration(60/ch.Freq) * time.Second)
	}

}

// search if there is a spike in the price above(or below) a factor (the trigger) of the bollinger band
func isSpike(ch *engine.CoinChannel, detectin bool) (spike uint, detect bool) {
	bbands := ch.GetBBand()
	bband := bbands[len(bbands)-1]
	pastband := bbands[len(bbands)-2]
	// see if the last price is over the trigger
	over := bband.High / bband.Upper
	below := bband.Lower / bband.Low
	// see if the close price of the last closed candle and the actual price are above a second trigger
	pover := pastband.High / pastband.Upper
	pbelow := pastband.Lower / pastband.Low
	// log when teh price is off band
	detect = false
	if over > 1 || below > 1 {
		if !detectin {
			log.Printf("Off band! p: %f, d: %f, u:%f, diff :%f , p/u : %f, d/p: %f \n", bband.Price, bband.Lower, bband.Upper, ch.BbandDelta(), over, below)
		}
		detect = true
	}

	if over > ch.Params.Trigger || over+pover > ch.Params.SecondTrigger {
		spike = 1
		if over+pover > ch.Params.SecondTrigger {
			spike = 2
		}
		log.Printf("Spike! price: %f , bband: %f , short it! \n", bband.Price, bband.Upper)

		return spike, detect

	} else if below > ch.Params.Trigger || below+pbelow > ch.Params.SecondTrigger {
		spike = 3
		if below+pbelow > ch.Params.SecondTrigger {
			spike = 4
		}
		log.Printf("Spike! price: %f , bband: %f , long it! \n", bband.Price, bband.Lower)
		return spike, detect
	}
	//log.Printf("No spike: p: %f, d: %f, u: %f \n", bband.Price, bband.Lower, bband.Upper)
	return 0, detect

}

func searchTarget(ch *engine.CoinChannel, spike uint) engine.Target {
	var out float64
	var otype uint
    channelCandles := ch.Candles()
	switch spike {
	// if the last candle spike down alone set the target to the lower price of the 3rd last candle
	case 1:
		out = channelCandles[len(channelCandles)-3].Low
		otype = 1
		// if 2 consecutives candles spike down  set the target to the lower price of the 4th last candle
	case 2:
		out = channelCandles[len(channelCandles)-4].Low
		otype = 1
		// if the last candle spike up alone set the target to the higher price of the 3rd last candle
	case 3:
		out = channelCandles[len(channelCandles)-3].High
		otype = 2
		// if 2 consecutives candles spike up  set the target to the higher price of the 4th last candle
	case 4:
		out = channelCandles[len(channelCandles)-4].High
		otype = 2

	}

	return engine.Target{
		Start:  channelCandles[len(channelCandles)-1].Close,
		Out:    out,
		Type:   otype,
		MTS:    float64(channelCandles[len(channelCandles)-1].MTS),
		Active: true,
	}

}

// wait for the price to reach the target or enter in a terminate condition
func waitClose(ch *engine.CoinChannel) bool {
    channelCandles := ch.Candles()
	last := channelCandles[len(channelCandles)-1].Close
    targetOut := ch.TargetOut()
    targetStart := ch.TargetStart()
	var profit float64
	// if short
	if ch.IsTarget(SHORT) {
		// reach the target
		if targetOut >= last {
            profit = targetStart * 0.998 / last
            ch.UpdateFunds(profit)
			log.Printf("Short order closed, start: %f, out: %f, profit: %f , acum: %f ", targetStart, last, profit, ch.Funds())
			ch.ResetTarget()
			return true
			// reach the stop loss or cross 3 times the ema50
		} else if last/targetStart >= ch.Params.Stop || eMAStop(ch) {
			profit = targetStart * 0.998 / last
            ch.UpdateFunds(profit)
			log.Printf("Orden Short liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", targetStart, last, profit, ch.Funds())
			ch.ResetTarget()
			return true

		}
		//if long
	} else if ch.IsTarget(LONG) {
		// reach the target
		if last >= targetOut {
			profit = last / targetStart * 0.998
            ch.UpdateFunds(profit)
			log.Printf("Orden Long cerrada, start: %f, out: %f, profit: %f , acum: %f ", targetStart, last, profit, ch.Funds())
			ch.ResetTarget()
			return true
			// reach the stop loss or cross 3 times the ema50
		} else if targetStart/last >= ch.Params.Stop || eMAStop(ch) {
			profit = last / targetStart * 0.998
            ch.UpdateFunds(profit)
			log.Printf("Orden Long liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", targetStart, last, profit, ch.Funds())
			ch.ResetTarget()
			return true

		}
	}
	return false
}

// liquidate the order
func liquidateOrders(ch *engine.CoinChannel) {
    channelCandles := ch.Candles()
	last := channelCandles[len(channelCandles)-1].Close
    targetStart := ch.TargetStart()
	profit := targetStart * 0.998 / last

	if ch.IsTarget(LONG) {
		profit = 1 / profit
	}
	ch.AddFunds(profit)
	log.Printf("Orden liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", targetStart, last, profit, ch.Funds())
	ch.ResetTarget()

}

// First stop, if the price cross the ema 50 three times and gets far from the target price the position is closed.
func eMAStop(ch *engine.CoinChannel) bool {
	// get the ema 50
	ema50 := ch.GetEMA(50)
    channelCandles := ch.Candles()
	var crossing, inv int
	var far bool

	for i := range channelCandles {
		inv = len(channelCandles) - i - 1 // count from the last candle and below
		// if the candle reach the enter position  the loop stop
		if ch.TargetMTSIsMet(float64(channelCandles[inv].MTS)) {
			break
		}
		// count the ema50 crossings
		if emaCross(channelCandles[inv], channelCandles[inv-1], ema50[inv]) {
			crossing++
		}
	}
	// see the last price
	last := channelCandles[len(channelCandles)-1]
	far = ema50[len(ema50)-1] >= last.Close // long position condition for far price
	if ch.IsTarget(LONG) {
		far = !far // short position condition for far price
	}
	return far && crossing >= 3
}

func emaCross(past, last *candle.Candle, ema float64) bool {
	down := past.Close >= ema && ema >= last.Close
	up := last.Close >= ema && ema >= past.Close
	return down || up
}
