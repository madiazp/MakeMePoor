package Utils

import (
	"fmt"
	"log"
)

func red(s interface{}) string {
	return fmt.Sprintf("\033[1;91m%v\033[0m", s)
}
func redbg(s interface{}) string {
	return fmt.Sprintf("\033[1;41m%v\033[0m", s)
}
func yellow(s interface{}) string {
	return fmt.Sprintf("\033[1;93m%v\033[0m", s)
}
func cyan(s interface{}) string {
	return fmt.Sprintf("\033[1;96m%v\033[0m", s)
}
func green(s interface{}) string {
	return fmt.Sprintf("\033[1;92m%v\033[0m", s)
}

func logger(alert, s string) {
	log.Printf("%s: %s\n", alert, s)
}

func EnterAlert(symbol string, price, target, stoploss float64, trend int) {
	trendStr := "SHORT"
	if trend == UP {
		trendStr = "LONG"
	}

	logger(yellow("[Activity]"), fmt.Sprintf("Open %s position %s: price: %s, target: %f, stop loss: $f", trendStr, symbol, green(price), target, stoploss))
}

func CloseAlert(symbol string, price, open, funds float64, trend int) {
	trendStr := "SHORT"
	if trend == UP {
		trendStr = "LONG"
	}
	gain := price - open
	var gainStr string
	if gain > 0 {
		gainStr = green(gain)
	} else {
		gainStr = red(gain)
	}
	logger(yellow("[Activity]"), fmt.Sprintf("Close %s position %s: price %s, open: %f, gain: %s, funds: %f ", trendStr, symbol, green(price), open, gainStr, funds))
}

func SpikeAlert(symbol string, price, lastPrice, lastMacd float64, trend int) {
	trendStr := "Up"
	if trend == DOWN {
		trendStr = "Down"
	}
	logger(cyan("[Info]"), fmt.Sprintf("Spike detected in %s trend on %s: price: %s, last price: %f, last macd: %f", trendStr, symbol, green(price), lastPrice, lastMacd))
}

func Error(msg string) {
	logger(redbg("[Error]"), red(msg))
}

func Activity(s string) {
	logger(yellow("[Activity]"), s)
}

func Bootstrap(symbol string, funds float64, simulator bool) {
	isSimulator := redbg("THIS IS NOT A SIMULATION")
	if simulator {
		isSimulator = "This is a Simulation"
	}
	logger(cyan("[Init]"), fmt.Sprintf("Starting MakeMePoor, trading %s with %f USD of funds, %s", symbol, funds, isSimulator))
}
