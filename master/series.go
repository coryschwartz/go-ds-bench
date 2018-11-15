package master

import (
	"encoding/json"
	"errors"
	"gonum.org/v1/plot"
	"io/ioutil"
	"log"
	"os"

	"github.com/ipfs/go-ds-bench/options"

	"golang.org/x/tools/benchmark/parse"
)

var Cont *bool

var ErrExists = errors.New("results for this bench already exist")
var errDone = errors.New("done")

type Series struct {
	Opts     []options.BenchOptions
	Workers  []*Worker
	Test     string // defined in Worker/worker_test.go
	PlotName string

	// ds -> Opts
	Results map[string][]*parse.Benchmark
}

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
}

func (s *Series) saveResults() error {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return ioutil.WriteFile("results-"+s.PlotName+".json", b, 0664)
}

func (s *Series) standardPlots() error {
	os.Mkdir("x_plots", 0755)

	if err := s.benchPlots("x_plots/", s.Results); err != nil {
		return err
	}

	// tag -> ds_name
	tagged := map[string]map[string][]*parse.Benchmark{}
	// type -> ds_name
	typed := map[string]map[string][]*parse.Benchmark{}

	for _, w := range s.Workers {
		for _, ds := range w.Spec.Datastores {
			if _, ok := typed[ds.Type]; !ok {
				typed[ds.Type] = map[string][]*parse.Benchmark{}
			}
			typed[ds.Type][ds.Name] = s.Results[ds.Name]

			for _, tag := range ds.Tags {
				if _, ok := tagged[tag]; !ok {
					tagged[tag] = map[string][]*parse.Benchmark{}
				}
				tagged[tag][ds.Name] = s.Results[ds.Name]
			}
		}
	}

	for t, res := range tagged {
		os.Mkdir("x_plots/tag-"+t, 0755)

		if err := s.benchPlots("x_plots/tag-"+t+"/", res); err != nil {
			return err
		}
	}

	for t, res := range typed {
		os.Mkdir("x_plots/ds-"+t, 0755)

		if err := s.benchPlots("x_plots/ds-"+t+"/", res); err != nil {
			return err
		}
	}

	os.Mkdir("x_plots/tag--avg/", 0755)
	if err := s.benchPlots("x_plots/tag--avg/", s.doAvg(tagged)); err != nil {
		return err
	}

	os.Mkdir("x_plots/ds--avg/", 0755)
	if err := s.benchPlots("x_plots/ds--avg/", s.doAvg(typed)); err != nil {
		return err
	}

	return nil
}

// doAvg averages items across category
func (s *Series) doAvg(in map[string]map[string][]*parse.Benchmark) map[string][]*parse.Benchmark {
	out := map[string][]*parse.Benchmark{}

	for cat, items := range in {
		avg := make([]*parse.Benchmark, len(s.Opts))
		for i := range s.Opts {
			avg[i] = &parse.Benchmark{}
		}

		for _, e := range items {
			for i, bench := range e {
				avg[i].AllocedBytesPerOp += bench.AllocedBytesPerOp / uint64(len(items))
				avg[i].AllocsPerOp += bench.AllocsPerOp / uint64(len(items))
				//bench.N
				//bench.Name
				avg[i].NsPerOp += bench.NsPerOp / float64(len(items))
				avg[i].MBPerS += bench.MBPerS / float64(len(items))
				//bench.Measured
				//bench.Ord
			}
		}
		out[cat] = avg
	}

	return out
}

func (s *Series) benchPlots(path string, results map[string][]*parse.Benchmark) error {
	if err := s.plot(path, results, xselPrimeRecs, yselNsPerOp, plot.LinearScale{}, TimeTicks{plot.DefaultTicks{}}, ""); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselNsPerOp, ZeroLogScale{}, TimeTicks{Log2Ticks{}}, "-log"); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselAllocs, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselAllocs, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselAlocKB, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselAlocKB, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselMBps, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
		return err
	}

	if err := s.plot(path, results, xselPrimeRecs, yselMBps, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
		return err
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
