package Engine

import (
	bitfinexCandle "github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
)

// CoinChannel is the main struct, it hold the inner state of a coin
type CoinChannel struct {
	Symbol  string                   // the symbol of the coin
	Freq    int                      // the recuency of the scan
	Span    int                      // the rate of the bollinger bands, typically 20
	Fail    error                    // inner state of error, logged when simulated
	candles []*bitfinexCandle.Candle // candles
	status  uint                     // state of the control state machine
	trend   uint                     // trend indicator, 0 if tunneling, 1 for down-trend, 2 to up-trend
	target  Target                   // price action
	funds   float64                  // total fund
	Params  Params                   // triggers and options
}

// order state
type Target struct {
	Start  float64 // the start price of the position
	Out    float64 // the target price
	Type   uint    // 1 for short, 2 for long
	MTS    float64 // MTS of the start
	Active bool    //if the order start and is waiting for the end.
}

// order struct to API, unused in simulation
type OrderData struct {
	Type   uint
	Price  float64
	Target float64
	MTS    float64
}

// Atributte manipulation methods
// Status
func (ch *CoinChannel) SetStatus( newStatus uint ) {
	ch.status = newStatus
}

func (ch *CoinChannel) Status() uint {
    return ch.status
}
// Trend
func (ch *CoinChannel) IsTrend() bool {
	return ch.trend != 0
}
// Target
func (ch *CoinChannel) TargetReached() bool {
    return ch.target.Active
}

func (ch *CoinChannel) IsTarget(t uint) bool {
    return ch.target.Type == t
}
func (ch *CoinChannel) TargetMTSIsMet(currentMTS float64) bool {
    return currentMTS == ch.target.MTS
}

func (ch *CoinChannel) TargetStart() float64 {
    return ch.target.Start
}

func (ch *CoinChannel) TargetOut() float64 {
    return ch.target.Out
}
func (ch *CoinChannel) ResetTarget() {
    ch.target = Target{}
}
// Funds
func (ch *CoinChannel) SetFunds(f float64) {
	ch.funds = f
}
func (ch *CoinChannel) AddFunds( profit float64) {
    ch.funds = ch.funds + profit
}
func (ch *CoinChannel) UpdateFunds( profit float64) {
    ch.funds = ch.funds * profit
}
func (ch *CoinChannel) Funds() float64 {
    return ch.funds
}

// Candles
func (ch *CoinChannel) Candles() []*bitfinexCandle.Candle {
    return ch.candles
}

func Init(funds float64, freq, span int, ticket string) *CoinChannel {

	// create coin channel
	ccticket := &CoinChannel{
		Symbol: ticket,
		Freq:   freq,
		Span:   span,
	}
	ccticket.SetFunds(funds)
    return ccticket
}
func (ch *CoinChannel) InitParams() {
	if !ch.Params.IsInit {
		ch.Params.Init()
	}

}

