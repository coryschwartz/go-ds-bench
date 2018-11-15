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
		log.Println("single")
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
					log.Printf("SKIPPING %s-%s-%d\n", s.PlotName, ds.Name, n)
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
				s.standardPlots()
				log.Println("saved")
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
	if err := s.plot(xselPrimeRecs, yselNsPerOp, plot.LinearScale{}, TimeTicks{plot.DefaultTicks{}}, ""); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselNsPerOp, ZeroLogScale{}, TimeTicks{Log2Ticks{}}, "-log"); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselAllocs, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselAllocs, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselMBps, plot.LinearScale{}, plot.DefaultTicks{}, ""); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselMBps, ZeroLogScale{}, Log2Ticks{}, "-log"); err != nil {
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
