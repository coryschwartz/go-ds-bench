package master

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ipfs/go-ds-bench/options"
	"golang.org/x/tools/benchmark/parse"
)

type DsFilter func([]options.WorkerDatastore) []options.WorkerDatastore

func NoTag(tag string) func([]options.WorkerDatastore) []options.WorkerDatastore {
	return func(datastores []options.WorkerDatastore) []options.WorkerDatastore {
		out := make([]options.WorkerDatastore, 0, len(datastores))
	next:
		for _, d := range datastores {
			for _, t := range d.Tags {
				if t == tag {
					continue next
				}
			}
			out = append(out, d)
		}
		return out
	}
}

func applyFilters(filters []DsFilter, opts []options.WorkerDatastore) []options.WorkerDatastore {
	for _, f := range filters {
		opts = f(opts)
	}
	return opts
}

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

type Worker struct {
	Spec workerSpec
}

func LoadWorkers() ([]*Worker, error) {
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return nil, err
	}

	var spec workerSpec
	err = json.NewDecoder(f).Decode(&spec)
	if err != nil {
		return nil, err
	}

	return []*Worker{{Spec: spec}}, nil
}

func (w *Worker) log(f string, v ...interface{}) {
	log.Printf("[Worker] "+f, v...)
}

func (w *Worker) runSingle(spec options.TestSpec) (*parse.Benchmark, error) {
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

	if spec.Datastore.Scripts.Pre != "" {
		log.Printf("running pre-run script for datastore %s: %s", spec.Datastore.Name, spec.Datastore.Scripts.Pre)
		c := exec.Command("/usr/bin/env", "bash", "-c", spec.Datastore.Scripts.Pre)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			return nil, err
		}
	}

	args := []string{"-test.benchtime=100ms", "-test.benchmem", "-test.bench", "BenchmarkSpec"}
	cmd := exec.Command(workerBin, args...)
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	var sout io.Reader
	sout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	sout = io.TeeReader(sout, os.Stderr)

	w.log("wd=%s", wd)
	w.log("start %s [cd %s; %s %s]", spec.Datastore.Name, wd, workerBin, strings.Join(args, " "))

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

	if spec.Datastore.Scripts.Post != "" {
		log.Printf("running post-run script for datastore %s: %s", spec.Datastore.Name, spec.Datastore.Scripts.Post)
		c := exec.Command("/usr/bin/env", "bash", "-c", spec.Datastore.Scripts.Post)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			panic(err)
		}
	}

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
