package master

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

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

func (s *Series) plot(x *xsel, y *ysel, yscale plot.Normalizer, ymarker plot.Ticker, suffix string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = s.PlotName
	p.Y.Label.Text = y.name
	p.X.Label.Text = x.name
	p.X.Scale = plot.LogScale{}
	p.Y.Scale = yscale
	p.Legend.Top = true
	p.X.Tick.Marker = Log2Ticks{}
	p.Y.Tick.Marker = ymarker

	p.Add(plotter.NewGrid())

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

	plotName := fmt.Sprintf("plot-%s-%s-%s%s.png", s.PlotName, x.name, y.name, suffix)
	plotName = strings.Replace(plotName, "/", "", -1)
	return p.Save(8*vg.Inch, 6*vg.Inch, plotName)
}

type Log2Ticks struct{}

var _ plot.Ticker = Log2Ticks{}

// Ticks returns Ticks in a specified range
func (t Log2Ticks) Ticks(min, max float64) []plot.Tick {
	if min < 0 {
		min = 1
	}

	val := math.Pow(2, math.Log2(min))
	max = math.Pow(2, math.Ceil(math.Log2(max)))
	var ticks []plot.Tick
	for val < max {
		for i := 1; i < 4; i++ {
			if i == 1{
				ticks = append(ticks, plot.Tick{Value: val, Label: formatFloatTick(val)})
			}
			ticks = append(ticks, plot.Tick{Value: val * float64(i)})
		}
		val *= 4
	}
	ticks = append(ticks, plot.Tick{Value: val, Label: formatFloatTick(val)})

	return ticks
}

func formatFloatTick(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// TimeTicks is suitable for axes representing time values.
type TimeTicks struct {
	Ticker plot.Ticker
}

// Ticks implements plot.Ticker.
func (t TimeTicks) Ticks(min, max float64) []plot.Tick {
	if t.Ticker == nil {
		t.Ticker = Log2Ticks{}
	}

	ticks := t.Ticker.Ticks(min, max)
	for i := range ticks {
		tick := &ticks[i]
		if tick.Label == "" {
			continue
		}
		tick.Label = time.Duration(tick.Value).String()
	}
	return ticks
}

// hacky version of logscale which doesn't panic
type ZeroLogScale struct{}

// Normalize returns the fractional logarithmic distance of
// x between min and max.
func (ZeroLogScale) Normalize(min, max, x float64) float64 {
	logMin := math.Log(min)
	return (math.Log(x) - logMin) / (math.Log(max) - logMin)
}