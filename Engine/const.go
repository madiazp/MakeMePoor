package Engine

const (
	TRIGGER       = 1.004 // trigger rate for one candle
	SECONDTRIGGER = 2.004 // trigger rate for 2 consecutives candles
	TRENDTRIGGER  = 1.005 // trigger ratio between the ema50 and ema10 for trending
	STOP          = 1.03  // Stop loss ratio
	FEES          = 1.002
)
