package master

import (
	"fmt"
	"github.com/ipfs/go-ds-bench/options"
	"strings"
	"sync"

	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var plotWg sync.WaitGroup

func genplots(plotName string, pathPrefix string, bopts []options.BenchOptions, results map[string]map[int]*parse.Benchmark, x *xsel, y *ysel, yscale plot.Normalizer, ymarker plot.Ticker, suffix string) error {
	plotWg.Add(1)
	go func() {
		defer plotWg.Done()
		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = plotName
		p.Y.Label.Text = y.name
		p.X.Label.Text = x.name
		p.X.Scale = ZeroLogScale{}
		p.Y.Scale = yscale
		p.Legend.Top = true
		p.X.Tick.Marker = Log2Ticks{}
		p.Y.Tick.Marker = ymarker

		p.Add(plotter.NewGrid())

		var lp []interface{}
		for dsname, p := range results {
			pts := make(plotter.XYs, 0, len(p))

			for n, bench := range p {
				if bench != nil {
					pts = append(pts, plotter.XY{
						X: x.sel(bopts[n]),
						Y: y.sel(bench),
					})
				}
			}
			lp = append(lp, dsname, pts)
		}

		if err := plotutil.AddLinePoints(p, lp...); err != nil {
			panic(err)
		}

		plotName = fmt.Sprintf("%s-%s-%s%s.png", plotName, x.name, y.name, suffix)
		plotName = strings.Replace(plotName, "/", "", -1)
		if err := p.Save(8*vg.Inch, 6*vg.Inch, pathPrefix+plotName); err != nil {
			panic(err)
		}
	}()
	return nil
}
