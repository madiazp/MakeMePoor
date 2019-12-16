package Retriever

import (
	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type BBData struct {
	Price   float64
	Central float64
	Upper   float64
	Down    float64
	T       float64
}

// Bollinger Band.
func BBand(cdls []*bfx.Candle, span int) (bband []BBData) {

	var avg, std float64
	var data []float64

	for i, cdl := range cdls {
		data = append(data, cdl.Close)
		// wait for enought data
		if i > span {

			avg, std = stat.MeanStdDev(data[i-span:i], nil)
			bband = append(bband, BBData{
				Price:   data[i],
				Central: avg,
				Upper:   avg + 2*std,
				Down:    avg - 2*std,
				T:       float64(cdl.MTS),
			})

		}
	}
	return bband

}

// RSI
func RSI(cdls []*bfx.Candle, span int) (rsi, t []float64) {
	// primer cierre
	close := cdls[0].Close
	// smmas evaluated with the first candle (to avoid div 0)
	smd := close
	smu := smd
	var actuald, actualu float64

	for _, cdl := range cdls[1:] {
		t = append(t, float64(cdl.MTS))
		// Gain & Loss
		if close > cdl.Close {
			//Loss
			actuald = close - cdl.Close
			actualu = 0
		} else {
			//Gain
			actuald = 0
			actualu = cdl.Close - close

		}
		// upper and down smma
		smd = smma(smd, actuald, span)
		smu = smma(smu, actualu, span)
		// rsi
		rsi = append(rsi, 100-100/(1+smu/smd))
		close = cdl.Close

	}
	return rsi, t
}

//MACD
func MACD(cdls []*bfx.Candle, span int) (macdHist, t []float64) {
	fema := cdls[0].Close
	lema := fema
	var signal, macd float64

	for _, cdl := range cdls[1:] {
		t = append(t, float64(cdl.MTS))
		fema = ema(cdl.Close, fema, 26) //fast ema
		lema = ema(cdl.Close, lema, 10) //longer ema
		macd = lema - fema
		signal = ema(macd, signal, 9)

		macdHist = append(macdHist, macd-signal) //MACD Histogram

	}
	return macdHist, t
}

func MACDPlotter(macd, t []float64) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "RSI"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err = plotutil.AddLines(p,
		"MACD", floattopoint(macd, t),
	)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(15*vg.Inch, 15*vg.Inch, "macdpoints.png"); err != nil {
		panic(err)
	}
}
func RSIPlotter(rsi, t []float64) {

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "RSI"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err = plotutil.AddLines(p,
		"rsi", floattopoint(rsi, t),
		"up", consttopoint(70, t),
		"down", consttopoint(30, t),
	)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(15*vg.Inch, 15*vg.Inch, "rsipoints.png"); err != nil {
		panic(err)
	}

}

func BandPlotter(bband []BBData) {

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	price, central, down, upper := bbandtopoint(bband)

	err = plotutil.AddLines(p,
		"down", down,
		"central", central,
		"upper", upper,
		"price", price,
	)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(15*vg.Inch, 15*vg.Inch, "points.png"); err != nil {
		panic(err)
	}

}

// smooth movil avarage
func smma(sm, x float64, n int) float64 {
	fn := float64(n)
	return ((fn-1)*sm + x) / fn

}

// exponential moving avarage
func ema(x, em, k float64) float64 {
	k = 2 / (k + 1)
	return x*k + (1-k)*em

}

func floattopoint(point, t []float64) plotter.XYs {
	pts := make(plotter.XYs, len(point))
	for i, _ := range pts {

		pts[i].X = t[i] / 60000
		pts[i].Y = point[i]

	}
	return pts
}

func bbandtopoint(bband []BBData) (price, central, down, upper plotter.XYs) {
	price = make(plotter.XYs, len(bband))
	central = make(plotter.XYs, len(bband))
	down = make(plotter.XYs, len(bband))
	upper = make(plotter.XYs, len(bband))
	for i, bbd := range bband {

		price[i].X = bbd.T
		price[i].Y = bbd.Price

		central[i].X = bbd.T
		central[i].Y = bbd.Central

		down[i].X = bbd.T
		down[i].Y = bbd.Down

		upper[i].X = bbd.T
		upper[i].Y = bbd.Upper
	}
	return price, central, down, upper
}

func consttopoint(con float64, t []float64) plotter.XYs {
	pts := make(plotter.XYs, len(t))
	for i, _ := range pts {

		pts[i].X = t[i] / 60000
		pts[i].Y = con

	}
	return pts

}
