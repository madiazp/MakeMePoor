package Strategy

import (
	"log"
	"time"

	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	aura "github.com/logrusorgru/aurora"
)

func (ch *CoinChannel) Start15M() {
	log.Println("Comienza el scan")
	savedfreq := ch.Freq
	for {
		err := ch.scan(bfx.FifteenMinutes)
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

func (ch *CoinChannel) isSpike() uint {
	bbands := ch.getBBand()
	bband := bbands[len(bbands)-1]
	pastband := bbands[len(bbands)-2]

	over := bband.High / bband.Upper
	below := bband.Down / bband.Low

	pover := pastband.High / pastband.Upper
	pbelow := pastband.Down / pastband.Low

	if over > 1 || below > 1 {
		log.Printf(aura.Sprintf("Off band! p: %f, d: %f, u:%f, diff :%f , p/u : %f, d/p: %f \n", bband.Price, bband.Down, bband.Upper, aura.Blue(ch.bbandDelta()), aura.Colorize(over, aura.RedBg), aura.Colorize(below, aura.GreenBg)))
	}

	if over > TRIGGER || over+pover > SECONDTRIGGER {
		log.Printf("Spike! price: %f , bband: %f , short it! \n", bband.Price, bband.Upper)
		ch.status = 1
		return 1

	} else if below > TRIGGER || below+pbelow > SECONDTRIGGER {
		log.Printf("Spike! price: %f , bband: %f , long it! \n", bband.Price, bband.Down)
		ch.status = 1
		return 2
	}
	//log.Printf("No spike: p: %f, d: %f, u: %f \n", bband.Price, bband.Down, bband.Upper)
	return 0

}

func (ch *CoinChannel) searchTarget(spike uint) Target {
	var out float64
	if spike == 1 {
		out = ch.candles[len(ch.candles)-4].Low
	} else if spike == 2 {
		out = ch.candles[len(ch.candles)-4].High
	}

	return Target{
		Start: ch.candles[len(ch.candles)-1].Close,
		Out:   out,
		Type:  spike,
		MTS:   float64(ch.candles[len(ch.candles)-1].MTS),
	}

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
		} else if last/ch.target.Start >= STOP || ch.EMAStop() {
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

		} else if ch.target.Start/last >= STOP || ch.EMAStop() {
			profit = last / ch.target.Start * 0.998
			ch.funds = ch.funds * profit
			log.Printf("Orden Long liquidada!!, start: %f, out: %f, profit: %f , acum: %f ", ch.target.Start, last, profit, ch.funds)
			return true

		}
	}
	return false
}

// First stop, if the price cross the ema 50 three times and gets far from the target price the position is closed.
func (ch *CoinChannel) EMAStop() bool {
	// get the ema 50
	ema50 := ch.getEMA50()

	var crossing, inv int
	var far bool

	for i, _ := range ch.candles {
		inv = len(ch.candles) - i // count from the last candle and below
		// if the candle reach the enter position  the loop stop
		if float64(ch.candles[inv].MTS) == ch.target.MTS {
			break
		}
		// count the ema50 crossings
		if emaCross(ch.candles[inv], ch.candles[inv-1], ema50[inv].Ema) {
			crossing++
		}
	}
	// see the last price
	last := ch.candles[len(ch.candles)-1]
	far = ema50[len(ema50)-1].Ema >= last.Close // long position condition for far price
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
