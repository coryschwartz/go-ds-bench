package master

import (
	"math"
	"strconv"
	"time"

	"github.com/ipfs/go-ds-bench/options"

	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
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

var yselAlocKB = &ysel{
	name: "allocKBs/op",
	sel: func(b *parse.Benchmark) float64 {
		return float64(b.AllocedBytesPerOp) / 1024.0
	},
}

var xselPrimeRecs = &xsel{
	name: "prime-count",
	sel: func(opt options.BenchOptions) float64 {
		return float64(opt.PrimeRecordCount)
	},
}

type Log2Ticks struct{}

var _ plot.Ticker = Log2Ticks{}

// Ticks returns Ticks in a specified range
func (t Log2Ticks) Ticks(min, max float64) []plot.Tick {
	if min < 0 {
		min = 1
	}

	val := math.Pow(2, math.Log2(min))
	if val == 0 {
		val = 1
	}
	max = math.Pow(2, math.Ceil(math.Log2(max)))
	var ticks []plot.Tick
	for val < max {
		for i := 1; i < 4; i++ {
			if i == 1 {
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
