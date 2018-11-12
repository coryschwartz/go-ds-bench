package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ipfs/go-ds-bench/options"
	"github.com/pkg/errors"
	"golang.org/x/tools/benchmark/parse"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var workerBin = "./worker.test"

func init() {
	abs, err := filepath.Abs(workerBin)
	if err != nil {
		panic(err)
	}

	workerBin = abs
}

type workerSpec struct {
	Datastores []options.WorkerDatastore
}

type worker struct {
	Spec workerSpec
}

func loadWorkers() ([]*worker, error) {
	f, err := os.Open("test-spec.json")
	if err != nil {
		return nil, err
	}

	var spec workerSpec
	err = json.NewDecoder(f).Decode(&spec)
	if err != nil {
		return nil, err
	}

	return []*worker{{Spec: spec}}, nil
}

func (w *worker) log(f string, v ...interface{}) {
	log.Printf("[worker] "+f, v...)
}

func (w *worker) runSingle(spec options.TestSpec) (*parse.Benchmark, error) {
	wd, err := ioutil.TempDir("/tmp", "dsbench-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(wd)

	specJson, err := json.Marshal(&spec)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(wd, "spec.json"), specJson, 0644); err != nil {
		return nil, err
	}

	cmd := exec.Command(workerBin, "-test.benchtime=100ms", "-test.benchmem", "-test.bench", "BenchmarkSpec")
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	var sout io.Reader
	sout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	sout = io.TeeReader(sout, os.Stderr)

	w.log("start")

	var wg sync.WaitGroup
	wg.Add(1)

	var rerr error
	go func() {
		defer wg.Done()
		rerr = cmd.Run()
	}()

	w.log("parse")

	bset, err := parse.ParseSet(sout)
	if err != nil {
		return nil, err
	}

	w.log("wait")
	wg.Wait()

	if len(bset) != 1 {
		return nil, errors.New("unexpected bench count")
	}

	for _, b := range bset {
		if len(b) != 1 {
			return nil, errors.New("unexpected bench len")
		}

		return b[0], nil
	}

	panic("shouldn't be here")
}

type series struct {
	Opts    []options.BenchOptions
	Workers []*worker
	Test    string // defined in worker/worker_test.go

	// ds -> Opts
	Results map[string][]*parse.Benchmark
}

func (s *series) benchSeries() error {
	for _, w := range s.Workers {
		for _, ds := range w.Spec.Datastores {
			for _, opt := range s.Opts {
				r, err := w.runSingle(options.TestSpec{
					Datastore: ds,
					Options: opt,
					Test: s.Test,
				})
				if err != nil {
					return err
				}
				s.Results[ds.Name] = append(s.Results[ds.Name], r)
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: master [series]")
		fmt.Println("               series [Test] [option] [from] [to]")
		return
	}

	workers, err := loadWorkers()
	if err != nil {
		panic(err)
	}

	testSeries := &series{
		Test: "get",
		Opts: options.DefaultBenchOpts,

		Workers: workers,
		Results: map[string][]*parse.Benchmark{},
	}

	if err := testSeries.benchSeries(); err != nil {
		panic(err)
	}

	b, err := json.Marshal(testSeries)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("results.json", b, 0664)
}
