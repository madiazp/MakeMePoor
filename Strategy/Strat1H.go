package Strategy

import (
	"log"
	"time"

	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
)

func (ch *CoinChannel) Start1H() {
	log.Println("Comienza el scan")
	savedfreq := ch.Freq
	var detect bool
	var spike uint
	for {
		err := ch.scan(bfx.FifteenMinutes)
		if err != nil {
			ch.Fail = err
			return
		}
		switch ch.status {
		case 0:

			if spike, detect = ch.isSpike(detect); spike != 0 {
				ch.status = 1
				ch.Freq = 30
			}
		case 1:
			spike, detect = ch.isSpike(detect)
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

func (ch *CoinChannel) BBDetect() uint {
	bands := ch.getBBand()
	bband := bands[len(bands)-1]
	if bband.Mid*0.998 > bband.Upper {
		return 1
	} else if bband.Down > bband.Mid*0.998 {
		return 2

	}
	return 0

}
