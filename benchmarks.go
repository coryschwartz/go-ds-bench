package main

import (
	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
)

func run(w []*worker) error {
 return benchBasicGet(w)
}

func benchBasicGet(w []*worker) error {
	testSeries := &series{
		Test:     "get",
		PlotName: "get-all-default",
		Opts:     options.DefaultBenchOpts,

		Workers: w,
		Results: map[string][]*parse.Benchmark{},
	}

	if err := testSeries.benchSeries(); err != nil {
		return err
	}

	if err := testSeries.plot(xselPrimeRecs, yselNsPerOp, plot.LogScale{}, plot.LinearScale{}); err != nil {
		return err
	}

	if err := testSeries.plot(xselPrimeRecs, yselAllocs, plot.LogScale{}, plot.LinearScale{}); err != nil {
		return err
	}

	if err := testSeries.plot(xselPrimeRecs, yselMBps, plot.LogScale{}, plot.LinearScale{}); err != nil {
		return err
	}
	return nil
}
