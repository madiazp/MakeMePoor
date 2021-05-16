package main

import engine "github.com/madiazp/MakeMePoor/Engine"
import strategy "github.com/madiazp/MakeMePoor/Strategy"

func main() {
    coinChannel := engine.Init(200, 20, 20, "tBTCUSD")
	strategy.Start15M(coinChannel)
}
