package master

import (
	"fmt"
	"github.com/gonum/stat"
	"github.com/ipfs/go-ds-bench/options"
	"os"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var plotWg sync.WaitGroup

type pt struct{
	plotter.XYs
	plotter.YErrors
}
func (p *pt) Len() int { return len(p.XYs) }
func (p *pt) Less(i, j int) bool { return p.XYs[i].X < p.XYs[j].X }
func (p *pt) Swap(i, j int) {
	p.XYs[i], p.XYs[j] = p.XYs[j], p.XYs[i]
	p.YErrors[i], p.YErrors[j] = p.YErrors[j], p.YErrors[i]
}

func genplots(plotName string, pathPrefix string, bopts []options.BenchOptions, results map[string]map[int][]*parse.Benchmark, x *xsel, y *ysel, yscale plot.Normalizer, ymarker plot.Ticker, suffix string) error {
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
		var lpe []interface{}
		for dsname, p := range results {
			byX := map[float64][]float64{}

			var pts pt
			//pts := make(plotter.XYs, 0, len(p))

			for n, benches := range p {
				for _, bench := range benches {
					if bench != nil {
						byX[x.sel(bopts[n])] = append(byX[x.sel(bopts[n])], y.sel(bench))
					}
				}
			}

			for x, ys := range byX {
				y, stddev := stat.MeanStdDev(ys, nil)

				pts.XYs = append(pts.XYs, plotter.XY{
					X: x,
					Y: y,
				})
				pts.YErrors = append(pts.YErrors, struct{ Low, High float64 }{Low: stddev/-2, High: stddev/2})
			}

			sort.Sort(&pts)

			lp = append(lp, dsname, &pts)
			lpe = append(lpe, &pts)
		}

		if err := plotutil.AddLinePoints(p, lp...); err != nil {
			panic(err)
		}

		if err := plotutil.AddErrorBars(p, lpe...); err != nil {
			//panic(err)
		}

		if err := os.Mkdir(pathPrefix+plotName, 0755); err != nil && !os.IsExist(err) {
			panic(err)
		}
		fName := fmt.Sprintf("%s-%s%s.png", x.name, y.name, suffix)
		fName = strings.Replace(fName, "/", "", -1)
		if err := p.Save(8*vg.Inch, 6*vg.Inch, pathPrefix+plotName+"/"+fName); err != nil {
			panic(err)
		}
	}()
	return nil
}
