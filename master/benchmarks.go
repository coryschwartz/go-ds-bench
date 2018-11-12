package master

import (
	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
)

func defaultBench(s *Series) error {
	if err := s.benchSeries(); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselNsPerOp, plot.LogScale{}, plot.LinearScale{}); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselAllocs, plot.LogScale{}, plot.LinearScale{}); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselMBps, plot.LogScale{}, plot.LinearScale{}); err != nil {
		return err
	}

	return nil
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
		Test:     "get",
		PlotName: "get-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}

func BenchBasicAdd(w []*Worker) error {
	testSeries := &Series{
		Test:     "get",
		PlotName: "get-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}

func BenchBasicAddBatch(w []*Worker) error {
	testSeries := &Series{
		Test:     "get",
		PlotName: "get-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	return defaultBench(testSeries)
}
