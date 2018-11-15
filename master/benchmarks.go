package master

import (
	"log"

	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
)

func defaultBench(s *Series) error {
	if err := s.benchSeries(NoTag("memory")); err != nil {
		if err == ErrExists {
			log.Printf("SKIPPING %s (using existing results)", s.PlotName)
			if err := s.loadExistingResults(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return s.standardPlots()
}

func BenchBasicGet(w []*Worker) error {
	testSeries := &Series{
		Test:     "get",
		PlotName: "get-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}

func BenchBasicHas(w []*Worker) error {
	testSeries := &Series{
		Test:     "has",
		PlotName: "has-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}

func BenchBasicAdd(w []*Worker) error {
	testSeries := &Series{
		Test:     "add",
		PlotName: "add-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}

func BenchBasicAddBatch(w []*Worker) error {
	testSeries := &Series{
		Test:     "add-batch",
		PlotName: "add-batch-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}
