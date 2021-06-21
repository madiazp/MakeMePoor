package Engine

import (
	bitfinexCandle "github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
)

// CoinChannel is the main struct, it hold the inner state of a coin
type CoinChannel struct {
	Symbol       string                   // the symbol of the coin
	Freq         int64                    // the recuency of the scan
	Span         int                      // the rate of the bollinger bands, typically 20
	Fail         error                    // inner state of error, logged when simulated
	candles      []*bitfinexCandle.Candle // candles
	status       uint                     // state of the control state machine
	trend        int                      // trend indicator, 0 if tunneling, 1 for down-trend, 2 to up-trend
	target       Target                   // price action
	funds        float64                  // total fund
	timeFrame    uint                     // main candle resolution
	isSimulation bool                     // true if the instance is a simulation
	Params       *Params                  // triggers and options
}

// order state
type Target struct {
	Start  float64 // the start price of the position
	Out    float64 // the target price
	Type   int     // 1 for short, 2 for long
	MTS    float64 // MTS of the start
	Stop   float64 // Stop loss
	Active bool    //if the order start and is waiting for the end.
}

// order struct to API, unused in simulation
type OrderData struct {
	Type   int
	Price  float64
	Target float64
	MTS    float64
}

// Atributte manipulation methods
// Status
func (ch *CoinChannel) SetStatus(newStatus uint) {
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
func (ch *CoinChannel) Target() Target {
	return ch.target
}

func (ch *CoinChannel) SetTarget(t Target) {
	ch.target = t
}
func (ch *CoinChannel) ResetTarget() {
	ch.target = Target{}
}

// Funds
func (ch *CoinChannel) SetFunds(f float64) {
	ch.funds = f
}
func (ch *CoinChannel) AddFunds(profit float64) {
	ch.funds = ch.funds + profit
}
func (ch *CoinChannel) UpdateFunds(profit float64) {
	ch.funds = ch.funds * profit
}
func (ch *CoinChannel) Funds() float64 {
	return ch.funds
}

// Candles
func (ch *CoinChannel) Candles() []*bitfinexCandle.Candle {
	return ch.candles
}

func (ch *CoinChannel) TimeFrame() uint {
	return ch.timeFrame
}
func (ch *CoinChannel) Simulation() bool {
	return ch.isSimulation
}
func (ch *CoinChannel) Simulate() {
	ch.isSimulation = true
}

func Init(funds float64, freq int64, span int, mainTimeFrame uint, ticket string, simulate, divergence bool) *CoinChannel {

	// create coin channel
	ccticket := &CoinChannel{
		Symbol:    ticket,
		Freq:      freq,
		Span:      span,
		timeFrame: mainTimeFrame,
		Params:    &Params{},
	}
	ccticket.SetFunds(funds)
	if simulate {
		ccticket.Simulate()
	}
	ccticket.InitParams(divergence)
	return ccticket
}
func (ch *CoinChannel) InitParams(divergence bool) {
	ch.Params.Init()
	ch.Params.Divergence = divergence

}
