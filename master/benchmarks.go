package master

import (
	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
)

func BenchBasicGet() *Series {
	return &Series{
		Test:     "get",
		PlotName: "get",
		Opts:     options.DefaultBenchOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBasicHas() *Series {
	return &Series{
		Test:     "has",
		PlotName: "has",
		Opts:     options.DefaultBenchOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBasicAdd() *Series {
	return &Series{
		Test:     "add",
		PlotName: "add",
		Opts:     options.DefaultBenchOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBasicAddBatch() *Series {
	return &Series{
		Test:     "add-batch",
		PlotName: "add-batch",
		Opts:     options.DefaultBenchOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}
