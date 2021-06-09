package Utils

import (
	bitfinexCommon "github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
)

const (
	UP   = 1
	DOWN = 2
)

func TimesFrames() []bitfinexCommon.CandleResolution {
	return []bitfinexCommon.CandleResolution{
		bitfinexCommon.OneMinute,      //0
		bitfinexCommon.FiveMinutes,    //1
		bitfinexCommon.FifteenMinutes, //2
		bitfinexCommon.ThirtyMinutes,  //3
		bitfinexCommon.OneHour,        //4
		bitfinexCommon.ThreeHours,     //5
		bitfinexCommon.SixHours,       //6
		bitfinexCommon.TwelveHours,    //7
		bitfinexCommon.OneDay,         //8
		bitfinexCommon.OneWeek,        //9
		bitfinexCommon.TwoWeeks,       //10
		bitfinexCommon.OneMonth,       //11
	}
}
