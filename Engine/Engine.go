package Engine

import (
	"log"

	bitfinexCommon "github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"

	comms "github.com/madiazp/MakeMePoor/Comms"
	utils "github.com/madiazp/MakeMePoor/Utils"
)

// Engine methods
// scan the candles
func (ch *CoinChannel) Scan(TimeFrame bitfinexCommon.CandleResolution) error {
	var err error
	ch.candles, err = comms.Scan(TimeFrame, ch.Symbol)
	if err != nil {
		return err
	}
	return nil

}

// start an order
func (ch *CoinChannel) SendOrder(t Target) {
	if !ch.Params.IsInit {
		ch.Params.Init()
	}
	var tp string
	if t.Type == 1 {
		tp = "Short"
		if ch.Params.Fees >= t.Start/t.Out {
			ch.status = 0
			ch.target = Target{}
			log.Printf("ROI muy bajo!, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)
			return
		}
	} else {
		tp = "Long"
		if ch.Params.Fees >= t.Out/t.Start {
			ch.status = 0
			ch.target = Target{}
			log.Printf("ROI muy bajo!, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)
			return
		}
	}
	ch.target = t
	ch.status = 2

	log.Printf("Orden simulada, Tipo: %s, Entrada: %f, Target: %f ", tp, t.Start, t.Out)

}

func (ch *CoinChannel) ScanTargetOrStop(trend int) bool {
	candles := ch.Candles()
	if trend == utils.DOWN {
		return candles[0].Close < ch.target.Out || candles[0].Close > ch.target.Stop
	} else {
		return candles[0].Close > ch.target.Out || candles[0].Close < ch.target.Stop
	}
}
