package Strategy

import (
	"log"
	"time"

	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	aura "github.com/logrusorgru/aurora"
)

// main state machine of the strategy
func (ch *CoinChannel) Start15M() {
	log.Println("Comienza el scan")
	savedfreq := ch.Freq // hold the freq
	for {
		err := ch.scan(bfx.FifteenMinutes) // use 15 min candles
		if err != nil {
			ch.Fail = err
		}
		// if trending stop everything
		if ch.trend != 0 {
			ch.status = 3
		}

		switch ch.status {
		// first state, look up for spike out of the bollinger bands
		case 0:

			if ch.isSpike() != 0 {
				ch.status = 1
				ch.Freq = 30
			}
		// if a spike is found then make an order and hold a position
		case 1:
			spike := ch.isSpike()
			target := ch.searchTarget(spike)
			ch.sendOrder(target)
		// wait for the price to reach the target and then reset
		case 2:
			ch.Freq = savedfreq
			if ch.waitClose() {
				ch.status = 0
			}
		// if the price is trending, liquidate every position and wait to stop trending
		case 3:
			ch.Freq = savedfreq
			if ch.target.Active {
				ch.liquidateOrders()
			}

		}
		// the max rate of the bitfinex api is 60 request per minute
		time.Sleep(time.Duration(60/ch.Freq) * time.Second)
	}

}

// search if there is a spike in the price above(or below) a factor (the trigger) of the bollinger band
func (ch *CoinChannel) isSpike() (spike uint) {
	bbands := ch.getBBand()
	bband := bbands[len(bbands)-1]
	pastband := bbands[len(bbands)-2]
	// see if the last price is over the trigger
	over := bband.High / bband.Upper
	below := bband.Down / bband.Low
	// see if the close price of the last closed candle and the actual price are above a second trigger
	pover := pastband.High / pastband.Upper
	pbelow := pastband.Down / pastband.Low
	// log when teh price is off band
	if over > 1 || below > 1 {
		log.Printf(aura.Sprintf("Off band! p: %f, d: %f, u:%f, diff :%f , p/u : %f, d/p: %f \n", bband.Price, bband.Down, bband.Upper, aura.Blue(ch.bbandDelta()), aura.Colorize(over, aura.RedBg), aura.Colorize(below, aura.GreenBg)))
	}

	if over > TRIGGER || over+pover > SECONDTRIGGER {
		spike = 1
		if over+pover > SECONDTRIGGER {
			spike = 2
		}
		log.Printf("Spike! price: %f , bband: %f , short it! \n", bband.Price, bband.Upper)

		return spike

	} else if below > TRIGGER || below+pbelow > SECONDTRIGGER {
		spike = 3
		if below+pbelow > SECONDTRIGGER {
			spike = 4
		}
		log.Printf("Spike! price: %f , bband: %f , long it! \n", bband.Price, bband.Down)
		return spike
	}
	//log.Printf("No spike: p: %f, d: %f, u: %f \n", bband.Price, bband.Down, bband.Upper)
	return 0

}

func (ch *CoinChannel) searchTarget(spike uint) Target {
	var out float64
	var otype uint
	switch spike {
	// if the last candle spike down alone set the target to the lower price of the 3rd last candle
	case 1:
		out = ch.candles[len(ch.candles)-3].Low
		otype = 1
		// if 2 consecutives candles spike down  set the target to the lower price of the 4th last candle
	case 2:
		out = ch.candles[len(ch.candles)-4].Low
		otype = 1
		// if the last candle spike up alone set the target to the higher price of the 3rd last candle
	case 3:
		out = ch.candles[len(ch.candles)-3].High
		otype = 2
		// if 2 consecutives candles spike up  set the target to the higher price of the 4th last candle
	case 4:
		out = ch.candles[len(ch.candles)-4].High
		otype = 2

	}

	return Target{
		Start: ch.candles[len(ch.candles)-1].Close,
		Out:   out,
		Type:  otype,
		MTS:   float64(ch.candles[len(ch.candles)-1].MTS),
	}

}

// wait for the price to reach the target or enter in a terminate condition
func (ch *CoinChannel) waitClose() bool {
	last := ch.candles[len(ch.candles)-1].Close
	var profit float64
	// if short
	if ch.target.Type == 1 {
		// reach the target
		if ch.target.Out >= last {
			profit = ch.target.Start * 0.998 / last
			ch.funds = ch.funds * profit
			log.Printf("Orden Short cerrada, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			ch.target = Target{}
			return true
			// reach the stop loss or cross 3 times the ema50
		} else if last/ch.target.Start >= STOP || ch.eMAStop() {
			profit = ch.target.Start * 0.998 / last
			ch.funds = ch.funds * profit
			log.Printf("Orden Short liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			ch.target = Target{}
			return true

		}
		//if long
	} else if ch.target.Type == 2 {
		// reach the target
		if last >= ch.target.Out {
			profit = last / ch.target.Start * 0.998
			ch.funds = ch.funds * profit
			log.Printf("Orden Long cerrada, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			ch.target = Target{}
			return true
			// reach the stop loss or cross 3 times the ema50
		} else if ch.target.Start/last >= STOP || ch.eMAStop() {
			profit = last / ch.target.Start * 0.998
			ch.funds = ch.funds * profit

			log.Printf(aura.Sprintf(aura.Colorize("Orden Long liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", aura.BrightBg), ch.target.Start, last, profit, ch.funds))
			ch.target = Target{}
			return true

		}
	}
	return false
}

// liquidate the order
func (ch *CoinChannel) liquidateOrders() {
	last := ch.candles[len(ch.candles)-1].Close
	profit := ch.target.Start * 0.998 / last

	if ch.target.Type == 1 {
		ch.funds = ch.funds * profit
	} else if profit != 0 {
		ch.funds = ch.funds / profit
	}
	log.Printf(aura.Sprintf(aura.Colorize("Orden liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", aura.BrightBg), ch.target.Start, last, profit, ch.funds))
	ch.target = Target{}

}

// First stop, if the price cross the ema 50 three times and gets far from the target price the position is closed.
func (ch *CoinChannel) eMAStop() bool {
	// get the ema 50
	ema50 := ch.getEMA(50)

	var crossing, inv int
	var far bool

	for i, _ := range ch.candles {
		inv = len(ch.candles) - i - 1 // count from the last candle and below
		// if the candle reach the enter position  the loop stop
		if float64(ch.candles[inv].MTS) == ch.target.MTS {
			break
		}
		// count the ema50 crossings
		if emaCross(ch.candles[inv], ch.candles[inv-1], ema50[inv]) {
			crossing++
		}
	}
	// see the last price
	last := ch.candles[len(ch.candles)-1]
	far = ema50[len(ema50)-1] >= last.Close // long position condition for far price
	if ch.target.Type == 2 {
		far = !far // short position condition for far price
	}
	return far && crossing >= 3
}

func emaCross(past, last *bfx.Candle, ema float64) bool {
	down := past.Close >= ema && ema >= last.Close
	up := last.Close >= ema && ema >= past.Close
	return down || up
}
