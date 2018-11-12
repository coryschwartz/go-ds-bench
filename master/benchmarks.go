package master

import (
	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
	"log"
)

func defaultBench(s *Series) error {
	if err := s.benchSeries(NoTag("memory")); err != nil {
		if err == ErrExists {
			log.Printf("SKIPPING %s", s.PlotName)
			return nil
		}
		return err
	}

	if err := s.plot(xselPrimeRecs, yselNsPerOp, plot.LogScale{}, plot.LinearScale{}, ""); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselNsPerOp, plot.LogScale{}, plot.LogScale{}, "-log"); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselAllocs, plot.LogScale{}, plot.LinearScale{}, ""); err != nil {
		return err
	}

	if err := s.plot(xselPrimeRecs, yselMBps, plot.LogScale{}, plot.LinearScale{}, ""); err != nil {
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
