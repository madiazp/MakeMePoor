package Comms

import utils "github.com/madiazp/MakeMePoor/Utils"

func EnterOrder() {
}

func EnterOrderSimulation(symbol string, price, target, stoploss float64, trend int) {
	utils.EnterAlert(symbol, price, target, stoploss, trend)
}

func CloseOrder() {
}

func CloseOrderSimulation(symbol string, price, open, funds float64, trend int) float64 {
	factor := price / open
	if trend == utils.DOWN {
		factor = 1 / factor
	}
	newFunds := funds * factor
	utils.CloseAlert(symbol, newFunds-funds, price, open, newFunds, trend)
	return factor
}
