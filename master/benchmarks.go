package master

import (
	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
)

var LargeBlockOpts = options.OptionsRange2pow( // up to 16G of 256k records
	options.BenchOptions{1, 1 << 18, 64},
	options.BenchOptions{1 << 16, 1 << 18, 64}, 9)

var BlockSizeOpts = options.OptionsRange2pow( // up to 16G of scanning record sizes
	options.BenchOptions{1 << 16, 1, 64},
	options.BenchOptions{1 << 16, 1 << 18, 64}, 9)

var BatchSizeBlockSizeOpts = options.OptionsRange2pow(
	options.BenchOptions{1 << 16, 1, 1},
	options.BenchOptions{1 << 16, 1 << 18, 512}, 9)

func BenchBasicGet() *Series {
	return &Series{
		Test:     "get",
		PlotName: "get",
		Opts:     LargeBlockOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBasicHas() *Series {
	return &Series{
		Test:     "has",
		PlotName: "has",
		Opts:     LargeBlockOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBasicAdd() *Series {
	return &Series{
		Test:     "add",
		PlotName: "add",
		Opts:     LargeBlockOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBasicAddBatch() *Series {
	return &Series{
		Test:     "add-batch",
		PlotName: "add-batch",
		Opts:     LargeBlockOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

// Size scanning

func BenchBSizingGet() *Series {
	return &Series{
		Test:     "get",
		PlotName: "get-bsize",
		Opts:     BlockSizeOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBSizingHas() *Series {
	return &Series{
		Test:     "has",
		PlotName: "has-bsize",
		Opts:     BlockSizeOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBSizingAdd() *Series {
	return &Series{
		Test:     "add",
		PlotName: "add-bsize",
		Opts:     BlockSizeOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

func BenchBSizingAddBatch() *Series {
	return &Series{
		Test:     "add-batch",
		PlotName: "add-batch-bsize",
		Opts:     BlockSizeOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}

// Batch sizing

func BenchAddBatch() *Series {
	return &Series{
		Test:     "add-batch",
		PlotName: "add-batch-record-batch",
		Opts:     BatchSizeBlockSizeOpts,

		Results: map[string]map[int]*parse.Benchmark{},
	}
}
