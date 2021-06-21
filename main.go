package main

import (
	engine "github.com/madiazp/MakeMePoor/Engine"
	strategy "github.com/madiazp/MakeMePoor/Strategy"
)

func main() {
	coinChannel := engine.Init(200, 20, 20, 0, "tBTCUSD", true, false)
	strategy.Start(coinChannel)
}
