package master

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/ipfs/go-ds-bench/options"

	"golang.org/x/tools/benchmark/parse"
)

type Series struct {
	Opts     []options.BenchOptions
	Workers  []*Worker
	Test     string // defined in Worker/worker_test.go
	PlotName string

	// ds -> Opts
	Results map[string][]*parse.Benchmark
}

func (s *Series) benchSeries() error {
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
