package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
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

