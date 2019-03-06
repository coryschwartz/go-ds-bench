package master

import (
	"encoding/json"
	"errors"
	"gonum.org/v1/plot"
	"io/ioutil"
	"sync"

	"github.com/ipfs/go-ds-bench/options"

	"golang.org/x/tools/benchmark/parse"
)

var ErrExists = errors.New("results for this bench already exist")
var errDone = errors.New("done")

type Series struct {
	Opts     []options.BenchOptions
	Test     string // defined in Worker/worker_test.go
	PlotName string

	// ds -> Opts
	Results map[string]map[int]*parse.Benchmark

	lk sync.Mutex
}

func (s *Series) todo(ds string) []int {
	out := make([]int, 0, len(s.Opts))

	if s.Results[ds] == nil {
		s.Results[ds] = map[int]*parse.Benchmark{}
	}

	for n := range s.Opts {
		if _, ok := s.Results[ds][n]; !ok {
			out = append(out, n)
		}
	}

	return out
}

/*
func (s *Series) benchSeries(f ...DsFilter) error {
	log.Printf("BEGIN %s", s.PlotName)
	if _, err := os.Stat("results-" + s.PlotName + ".json"); !*Cont && !os.IsNotExist(err) {
		return ErrExists
	} else if !os.IsNotExist(err) && *Cont {
		log.Println("Continuing from existing results")
		s.loadExistingResults()
	}

	for {
		err := s.doSingle(f...)
		if err != nil {
			if err != errDone {
				return err
			}
			break
		}
	}

	return s.saveResults()
}

func (s *Series) doSingle(f ...DsFilter) error {
	for _, w := range s.Workers {
		for _, ds := range applyFilters(f, w.Spec.Datastores) {
			for n, opt := range s.Opts {
				if len(s.Results[ds.Name]) > n {
					continue
				}
				log.Printf("START %s-%s-%d\n", s.PlotName, ds.Name, n)

				r, err := w.runSingle(options.TestSpec{
					Datastore: ds,
					Options:   opt,
					Test:      s.Test,
				})
				if err != nil {
					return err
				}
				s.Results[ds.Name] = append(s.Results[ds.Name], r)

				log.Println("saving progress")
				s.saveResults()
				return nil
			}
		}
	}

	return errDone
}*/

func (s *Series) saveResults() error {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return ioutil.WriteFile("results-"+s.PlotName+".json", b, 0664)
}

// doAvg averages items across category
//TODO: verify it works properly after map results refactor
func (s *Series) doAvg(in map[string]map[string]map[int]*parse.Benchmark) map[string]map[int]*parse.Benchmark {
	out := map[string]map[int]*parse.Benchmark{}

	for cat, items := range in {
		avg := make(map[int]*parse.Benchmark, len(s.Opts))
		for i := range s.Opts {
			avg[i] = &parse.Benchmark{}
		}

		for _, e := range items {
			for i, bench := range e {
				if bench != nil { //TODO: might skew results, warn or something
					avg[i].AllocedBytesPerOp += bench.AllocedBytesPerOp / uint64(len(items))
					avg[i].AllocsPerOp += bench.AllocsPerOp / uint64(len(items))
					// bench.N
					// bench.Name
					avg[i].NsPerOp += bench.NsPerOp / float64(len(items))
					avg[i].MBPerS += bench.MBPerS / float64(len(items))
					// bench.Measured
					// bench.Ord
				}
			}
		}
		out[cat] = avg
	}

	return out
}

func benchPlots(plotName string, path string, bopts []options.BenchOptions, results map[string]map[int]*parse.Benchmark) error {
	var sels []*xsel

	if bopts[0].BatchSize != bopts[1].BatchSize {
		sels =append(sels, xselBatchSize)
	}

	if bopts[0].RecordSize != bopts[1].RecordSize {
		sels =append(sels, xselRecordSize)
	}

	if bopts[0].PrimeRecordCount != bopts[1].PrimeRecordCount {
		sels =append(sels, xselPrimeRecs)
	}

	for _, ixsel := range sels {
		if err := genplots(plotName, path, bopts, results, ixsel, yselNsPerOp, plot.LinearScale{}, TimeTicks{plot.DefaultTicks{}}, ""); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselNsPerOp, ZeroLogScale{}, TimeTicks{Log2Ticks{}}, "-log"); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselAllocs, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselAllocs, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselAlocKB, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselAlocKB, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselMBps, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
			return err
		}

		if err := genplots(plotName, path, bopts, results, ixsel, yselMBps, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Series) loadExistingResults() error {
	b, err := ioutil.ReadFile("results-" + s.PlotName + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, s)
}
