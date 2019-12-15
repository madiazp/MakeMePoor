package Retriever

import (
	bfx "github.com/bitfinexcom/bitfinex-api-go/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func SellVsBuy(ords []*bfx.BookUpdate) float64 {
	//log.Printf("orders: %d", len(ords))
	var buy, sell float64
	for _, ord := range ords {
		if ord.Side == 1 {

			sell += ord.Amount

		} else {

			buy += ord.Amount

		}
	}
	return sell / buy
}

func SvBPlotter(ratio []float64) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Sell vs Buy"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err = plotutil.AddLines(p,
		"ratio", fivesecpoint(ratio),
	)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(15*vg.Inch, 15*vg.Inch, "sentpoints.png"); err != nil {
		panic(err)
	}
}

func fivesecpoint(point []float64) plotter.XYs {
	pts := make(plotter.XYs, len(point))
	for i, _ := range pts {

		pts[i].X = float64(i) * 5
		pts[i].Y = point[i]

	}
	return pts
}
