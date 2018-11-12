package master

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/ipfs/go-ds-bench/options"

	"golang.org/x/tools/benchmark/parse"
)

var ErrExists = errors.New("results for this bench already exist")

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
	if _, err := os.Stat("results-" + s.PlotName + ".json"); !os.IsNotExist(err) {
		return ErrExists
	}

	for _, w := range s.Workers {
		for _, ds := range applyFilters(f, w.Spec.Datastores) {
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

func (s *Series) loadExistingResults() error {
	b, err := ioutil.ReadFile("results-" + s.PlotName + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, s)
}
