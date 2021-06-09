package Strategy

import (
	"time"

	bitfinexCandle "github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
	bitfinexCommon "github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	comms "github.com/madiazp/MakeMePoor/Comms"
	engine "github.com/madiazp/MakeMePoor/Engine"
	indicator "github.com/madiazp/MakeMePoor/Indicators"
	utils "github.com/madiazp/MakeMePoor/Utils"
)

// timeFrames for trend scan

type timeFrames struct {
	Main bitfinexCommon.CandleResolution
	Slow bitfinexCommon.CandleResolution
	Fast bitfinexCommon.CandleResolution
}

func (t *timeFrames) setTimesFrame(main uint) {
	globalTimeFrames := utils.TimesFrames()
	if int(main) > len(globalTimeFrames)-4 {
		utils.Error("time frame not supported")
		panic("time drame not supported")
	}
	t.Main = globalTimeFrames[main]
	t.Slow = globalTimeFrames[main+3]
	t.Fast = globalTimeFrames[main+1]
}

// InnerStatus for divergence and trend detector

type InnerStatus struct {
	MACD  float64
	Trend int
	Close float64
}

func (i *InnerStatus) setInerStatus(candles []*bitfinexCandle.Candle, trend int) {
	i.Trend = trend
	i.Close = candles[0].Close
	i.MACD, _, _ = indicator.MovingAverageConvergenceDivergence(candles)
}

func (i *InnerStatus) isDivergence(candles []*bitfinexCandle.Candle, trend int) (ans bool) {
	macd, _, _ := indicator.MovingAverageConvergenceDivergence(candles)
	last := candles[0].Close
	if trend == utils.UP {
		ans = macd > i.MACD && last < i.Close
	} else {
		ans = macd < i.MACD && last > i.Close
	}
	i.Close = last
	i.MACD = macd
	return ans
}
func (i *InnerStatus) ChangeTrend(trend int) (ans bool) {
	ans = trend != i.Trend
	i.Trend = trend
	return ans
}

// MACD divergence detector

func trendScan(slowCandles, fastCandles []*bitfinexCandle.Candle) (ans int) {

	slowEma := indicator.GetEMA(slowCandles, 50)
	fastEma := indicator.GetEMA(fastCandles, 50)
	ans = utils.UP
	if slowEma[len(slowEma)-1] < fastEma[len(fastEma)-1] {
		ans = utils.DOWN
	}
	return ans

}

func targetSet(candles []*bitfinexCandle.Candle, trend, stopFactor, targetFactor int) (target engine.Target) {
	atr := indicator.AvarageTrueRange(candles)
	if trend == utils.DOWN {
		atr = -atr
	}
	lastCandle := candles[len(candles)-1]
	return engine.Target{
		Start:  lastCandle.Close,
		Out:    lastCandle.Close + atr*float64(targetFactor),
		Type:   trend,
		MTS:    float64(lastCandle.MTS),
		Stop:   lastCandle.Close - atr*float64(stopFactor),
		Active: true,
	}
}

func Start(ch *engine.CoinChannel) {
	utils.Bootstrap(ch.Symbol, ch.Funds(), ch.Simulation())
	innerStatus := &InnerStatus{}
	tFrame := timeFrames{}
	tFrame.setTimesFrame(ch.TimeFrame())
	slowCandles, errSlow := comms.Scan(tFrame.Slow, ch.Symbol)
	if errSlow != nil {
		utils.Error(errSlow.Error())
	}
	fastCandles, errFast := comms.Scan(tFrame.Fast, ch.Symbol)
	if errFast != nil {
		utils.Error(errFast.Error())
	}
	trend := trendScan(slowCandles, fastCandles)
	ch.Scan(tFrame.Main)
	innerStatus.setInerStatus(ch.Candles(), trend)
	for {
		ch.Scan(tFrame.Main)
		machine(ch, trend, innerStatus)
		slowCandles, errSlow = comms.Scan(tFrame.Slow, ch.Symbol)
		if errSlow != nil {
			utils.Error(errSlow.Error())
		}
		fastCandles, errFast = comms.Scan(tFrame.Fast, ch.Symbol)
		if errFast != nil {
			utils.Error(errFast.Error())
		}
		trend = trendScan(slowCandles, fastCandles)
		time.Sleep(time.Duration(ch.Freq) * time.Second)
	}
}

func machine(ch *engine.CoinChannel, trend int, innerStatus *InnerStatus) {
	candles := ch.Candles()
	bband := indicator.BollingerBand(candles, 10, 1.5)
	switch ch.Status() {
	case 0:
		if trend == utils.DOWN {
			if bband.Price > bband.High {
				utils.SpikeAlert(ch.Symbol, bband.Price, innerStatus.MACD, innerStatus.Close, trend)
				ch.SetStatus(1)
			}
		} else {
			if bband.Price < bband.Low {
				utils.SpikeAlert(ch.Symbol, bband.Price, innerStatus.MACD, innerStatus.Close, trend)
				ch.SetStatus(1)
			}
		}
		innerStatus.Trend = trend
	case 1:
		if innerStatus.ChangeTrend(trend) {
			ch.SetStatus(0)
			break
		}
		if ((trend == utils.DOWN && bband.LastClose < bband.High) || (trend == utils.UP && bband.LastClose > bband.Low)) && // close spike condition
			((ch.Params.Divergence && innerStatus.isDivergence(candles, trend)) || !ch.Params.Divergence) { // divergence condition
			ch.SetTarget(targetSet(candles, trend, ch.Params.AtrStop, ch.Params.AtrTarget))
			ch.SetStatus(2)
			if ch.Simulation() {
				target := ch.Target()
				comms.EnterOrderSimulation(ch.Symbol, bband.Price, target.Out, target.Stop, trend)
			} else {
				comms.EnterOrder()
			}
		}
	case 2:
		if ch.ScanTargetOrStop(trend) {
			ch.SetStatus(0)
			if ch.Simulation() {
				target := ch.Target()
				ch.UpdateFunds(comms.CloseOrderSimulation(ch.Symbol, bband.Price, target.Start, ch.Funds(), trend))
			} else {
				comms.CloseOrder()
			}
			ch.ResetTarget()
		}
	}
}
