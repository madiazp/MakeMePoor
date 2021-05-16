package Comms

import(
	bitfinexCommon "github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	bitfinexCandle "github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
    "time"
)

func Scan(TimeFrame bitfinexCommon.CandleResolution, ticket string) ([]*bitfinexCandle.Candle, error) {

	client := rest.NewClient()
	now := time.Now()
	millis := now.UnixNano() / 1000000

	prior := now.Add(time.Duration(-24) * 1 * time.Hour)
	millisStart := prior.UnixNano() / 1000000
	start := bitfinexCommon.Mts(millisStart)
	end := bitfinexCommon.Mts(millis)

	candles, err := client.Candles.HistoryWithQuery(
		ticket,
		TimeFrame,
		start,
		end,
		1000,
		bitfinexCommon.OldestFirst,
	)
	if err != nil {

		return nil, err
	}
	return candles.Snapshot, nil

}
