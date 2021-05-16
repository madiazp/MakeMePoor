package Engine

type Params struct {
    Trigger       float64
    SecondTrigger float64
    TrendTrigger  float64
    Stop          float64
    Fees          float64
    IsInit        bool
}

func (p *Params) Init() {
    p.Trigger = TRIGGER
    p.SecondTrigger = SECONDTRIGGER
    p.TrendTrigger = TRENDTRIGGER
    p.Stop = STOP
    p.Fees = FEES
    p.IsInit = true
}
