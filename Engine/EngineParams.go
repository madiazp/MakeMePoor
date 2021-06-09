package Engine

type Params struct {
	Trigger       float64
	SecondTrigger float64
	TrendTrigger  float64
	Stop          float64
	Fees          float64
	AtrStop       int
	AtrTarget     int
	IsInit        bool
	Divergence    bool
}

func (p *Params) Init() {
	p.Trigger = TRIGGER
	p.SecondTrigger = SECONDTRIGGER
	p.TrendTrigger = TRENDTRIGGER
	p.Stop = STOP
	p.Fees = FEES
	p.AtrStop = ATRSTOP
	p.AtrTarget = ATRTARGET
	p.IsInit = true
}
