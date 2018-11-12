package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ipfs/go-ds-bench/options"
	"github.com/pkg/errors"
	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var workerBin = "./worker.test"

func init() {
	abs, err := filepath.Abs(workerBin)
	if err != nil {
		panic(err)
	}

	workerBin = abs
}

type workerSpec struct {
	Datastores []options.WorkerDatastore
}

type worker struct {
	Spec workerSpec
}

func loadWorkers() ([]*worker, error) {
	f, err := os.Open("test-spec.json")
	if err != nil {
		return nil, err
	}

	var spec workerSpec
	err = json.NewDecoder(f).Decode(&spec)
	if err != nil {
		return nil, err
	}

	return []*worker{{Spec: spec}}, nil
}

func (w *worker) log(f string, v ...interface{}) {
	log.Printf("[worker] "+f, v...)
}

func (w *worker) runSingle(spec options.TestSpec) (*parse.Benchmark, error) {
	wd, err := ioutil.TempDir("/tmp", "dsbench-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(wd)

	specJson, err := json.Marshal(&spec)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(wd, "spec.json"), specJson, 0644); err != nil {
		return nil, err
	}

	cmd := exec.Command(workerBin, "-test.benchtime=100ms", "-test.benchmem", "-test.bench", "BenchmarkSpec")
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	var sout io.Reader
	sout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	sout = io.TeeReader(sout, os.Stderr)

	w.log("start")

	var wg sync.WaitGroup
	wg.Add(1)

	var rerr error
	go func() {
		defer wg.Done()
		rerr = cmd.Run()
	}()

	w.log("parse")

	bset, err := parse.ParseSet(sout)
	if err != nil {
		return nil, err
	}

	w.log("wait")
	wg.Wait()

	if len(bset) != 1 {
		return nil, errors.New("unexpected bench count")
	}

	for _, b := range bset {
		if len(b) != 1 {
			return nil, errors.New("unexpected bench len")
		}

		return b[0], nil
	}

	panic("shouldn't be here")
}

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

type series struct {
	Opts     []options.BenchOptions
	Workers  []*worker
	Test     string // defined in worker/worker_test.go
	PlotName string

	// ds -> Opts
	Results map[string][]*parse.Benchmark
}

func (s *series) benchSeries() error {
	if _, err := os.Stat("results-" + s.PlotName + ".json"); !os.IsNotExist(err) {
		return errors.New("results for this bench already exist")
	}

	for _, w := range s.Workers {
		for _, ds := range w.Spec.Datastores {
			for _, opt := range s.Opts {
				r, err := w.runSingle(options.TestSpec{
					Datastore: ds,
					Options:   opt,
					Test:      s.Test,
				})
				if err != nil {
					return err
				}
				s.Results[ds.Name] = append(s.Results[ds.Name], r)
			}
		}
	}

	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return ioutil.WriteFile("results-"+s.PlotName+".json", b, 0664)
}

func (s *series) plot(x *xsel, y *ysel, xscale, yscale plot.Normalizer) error {
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

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: master [series]")
		fmt.Println("               series [Test] [option] [from] [to]")
		return
	}

	workers, err := loadWorkers()
	if err != nil {
		panic(err)
	}

	testSeries := &series{
		Test:     "get",
		PlotName: "get-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: workers,
		Results: map[string][]*parse.Benchmark{},
	}

	if err := testSeries.benchSeries(); err != nil {
		panic(err)
	}

	if err := testSeries.plot(xselPrimeRecs, yselNsPerOp, plot.LogScale{}, plot.LinearScale{}); err != nil {
		panic(err)
	}

	if err := testSeries.plot(xselPrimeRecs, yselAllocs, plot.LogScale{}, plot.LinearScale{}); err != nil {
		panic(err)
	}

	if err := testSeries.plot(xselPrimeRecs, yselMBps, plot.LogScale{}, plot.LinearScale{}); err != nil {
		panic(err)
	}
}
