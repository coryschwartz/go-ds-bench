package main

import (
	"flag"
	"fmt"

	"github.com/ipfs/go-ds-bench/master"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: master [specfile]")
		return
	}

	workers, err := master.LoadWorkers()
	if err != nil {
		panic(err)
	}

	run(workers)
}

func run(w []*master.Worker) {
	assert(master.BenchBasicGet(w))
	assert(master.BenchBasicHas(w))
	assert(master.BenchBasicAdd(w))
	assert(master.BenchBasicAddBatch(w))
}
