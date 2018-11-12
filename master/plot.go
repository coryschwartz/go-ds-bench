package master

import (
	"fmt"
	"strings"

	"github.com/ipfs/go-ds-bench/options"

	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type ysel struct {
	name string
	sel  func(*parse.Benchmark) float64
}

type xsel struct {
	name string
	sel  func(options.BenchOptions) float64
}

var yselNsPerOp = &ysel{
	name: "ns/op",
	sel: func(b *parse.Benchmark) float64 {
		return b.NsPerOp
	},
}

var yselMBps = &ysel{
	name: "MB/s",
	sel: func(b *parse.Benchmark) float64 {
		return b.MBPerS
	},
}

var yselAllocs = &ysel{
	name: "alloc/op",
	sel: func(b *parse.Benchmark) float64 {
		return float64(b.AllocsPerOp)
	},
}

var xselPrimeRecs = &xsel{
	name: "prime-count",
	sel: func(opt options.BenchOptions) float64 {
		return float64(opt.PrimeRecordCount)
	},
}

func (s *Series) plot(x *xsel, y *ysel, xscale, yscale plot.Normalizer) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = s.PlotName
	p.Y.Label.Text = y.name
	p.X.Label.Text = x.name
	p.X.Scale = xscale
	p.Y.Scale = yscale

	var lp []interface{}
	for dsname, p := range s.Results {
		pts := make(plotter.XYs, len(s.Opts))

		for n, bench := range p {
			pts[n].X = x.sel(s.Opts[n])
			pts[n].Y = y.sel(bench)
		}
		lp = append(lp, dsname, pts)
	}

	if err := plotutil.AddLinePoints(p, lp...); err != nil {
		return err
	}

	plotName := fmt.Sprintf("plot-%s-%s-%s.png", s.PlotName, x.name, y.name)
	plotName = strings.Replace(plotName, "/", "", -1)
	return p.Save(4*vg.Inch, 4*vg.Inch, plotName)
}

