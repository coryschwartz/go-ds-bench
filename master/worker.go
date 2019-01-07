package master

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ipfs/go-ds-bench/master/env"
	"io"
	"io/ioutil"
	"log"
	"os"
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

type Worker struct {
	Type string
	Spec map[string]interface{}
}

type BatchSpec struct {
	Datastores []options.WorkerDatastore
	// instance type -> worker id
	Workers map[string][]Worker

	// instance type -> series id
	Jobs map[string][]*Series
}

type workUnit struct {
	series int // test type, like add /  get
	ds     int // datastore
	point  int // which part is to be done
}

func BuildBatch(new func() (*BatchSpec, error), cont bool) (*BatchSpec, error) {
	if _, err := os.Stat("results.json"); !cont && !os.IsNotExist(err) {
		return nil, ErrExists
	} else if !os.IsNotExist(err) && cont {
		log.Println("Continuing from existing results")
		b, err := ioutil.ReadFile("results.json")
		if err != nil {
			return nil, err
		}

		var s BatchSpec
		err = json.Unmarshal(b, &s)
		return &s, err
	}

	return new()
}

// wuQueue implements a chan based FIFO queue
func wuQueue() struct {
	in  chan<- workUnit
	out <-chan workUnit
} {
	in := make(chan workUnit)
	out := make(chan workUnit)

	go func() {
		defer close(out)
		var todo []workUnit

		for {
			if len(todo) == 0 {
				if in == nil {
					break
				}
				wu, ok := <-in
				if !ok {
					break
				}
				todo = append(todo, wu)
			}
			select {
			case wu, ok := <-in:
				if !ok {
					in = nil
					break
				}
				todo = append(todo, wu)
			case out <- todo[len(todo)-1]:
				todo = todo[:len(todo)-1]
			}
		}
	}()

	return struct {
		in  chan<- workUnit
		out <-chan workUnit
	}{in: in, out: out}
}

func (b *BatchSpec) Start() error {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	// workUnit queues per instance type
	queues := map[string]struct {
		in  chan<- workUnit
		out <-chan workUnit
	}{}
	var wg sync.WaitGroup

	for itype, workers := range b.Workers {
		queues[itype] = wuQueue()

		for id, worker := range workers {
			log.Printf("Starting worker %s-%d", itype, id)
			wg.Add(1)
			go func(itype string, worker Worker) {
				defer wg.Done()
				for {
					select {
					case wu, ok := <-queues[itype].out:
						if !ok {
							log.Printf("Stopping worker %s-%d", itype, id)
							return
						}
						// TODO: don't ignore results (_)
						_, err := worker.run(b.Datastores[wu.ds], b.Jobs[itype][wu.series], wu.point)
						if err != nil {
							//TODO: figure this out
							panic(fmt.Errorf("UNHANDLED ERROR [%s-%d]: %s", itype, id, err))
						}
					case <-ctx.Done():
						log.Printf("Stopping worker %s-%d (ctx expired)", itype, id)
						return
					}
				}
			}(itype, worker)
		}
	}

	// create jobs
	for itype, srss := range b.Jobs {
		for dsid, ds := range b.Datastores {
			for sid, series := range srss {
				for _, point := range series.todo(ds.Name) {
					queues[itype].in <- workUnit{
						series: sid,
						ds:     dsid,
						point:  point,
					}
				}
			}
		}
		close(queues[itype].in)
	}

	wg.Wait()

	// TODO: grab errors
	// TODO: port/call func (s *Series) standardPlots()
	return nil
}

func (w *Worker) log(f string, v ...interface{}) {
	log.Printf("[Worker] "+f, v...)
}

func (w *Worker) replaceVars(s []string) []string {
	s1 := s[1]
	for m, r := range w.Spec["Vars"].(map[string]interface{}) {
		s1 = strings.Replace(s1, m, r.(string), -1)
	}
	return []string{s[0], s1}
}

type Piper interface {
	TeeReader(r io.Reader, w io.Writer) io.Reader
}

func (w *Worker) run(ids options.WorkerDatastore, series *Series, point int) (*parse.Benchmark, error) {
	init, ok := env.Handlers[w.Type]
	if !ok {
		return nil, fmt.Errorf("unknown remote type: %s", w.Type)
	}

	env, err := init(w.Spec)
	if err != nil {
		return nil, err
	}

	var ds options.WorkerDatastore
	if err := clone(&ids, &ds); err != nil {
		return nil, err
	}

	for n, t := range ds.Params {
		if sparam, ok := t.(string); ok {
			ds.Params[n] = w.replaceVars([]string{"", sparam})[1]
		}
	}

	spec := options.TestSpec{
		Datastore: ds,
		Options:   series.Opts[point],
		Test:      series.Test,
	}
	specJson, err := json.Marshal(&spec)
	if err != nil {
		return nil, err
	}
	if err := env.WriteFile("spec.json", specJson, 0644); err != nil {
		return nil, err
	}

	if err := env.CopyFile(ds.Scripts.Pre[0], "prerun.sh", 0755); err != nil {
		return nil, err
	}
	if err := env.CopyFile(ds.Scripts.Post[0], "postrun.sh", 0755); err != nil {
		return nil, err
	}
	if err := env.CopyFile(workerBin, "worker.test", 0755); err != nil {
		return nil, err
	}

	if len(ds.Scripts.Pre) != 0 {
		log.Printf("running pre-run script for datastore %s: %s", ds.Name, w.replaceVars(ds.Scripts.Pre))

		run := env.Cmd("/usr/bin/env", []string{"bash", "-c", "./prerun.sh " + w.replaceVars(ds.Scripts.Pre)[1]}, os.Stderr, os.Stderr)
		if err := run(); err != nil {
			return nil, err
		}
	}

	args := []string{"-test.benchmem", "-test.bench", "BenchmarkSpec"}

	pr, pw := io.Pipe()
	run := env.Cmd("./worker.test", args, pw, os.Stderr)
	sout := io.TeeReader(pr, os.Stderr)

	w.log("start %s [%s]", ds.Name, strings.Join(args, " "))

	var wg sync.WaitGroup
	wg.Add(1)

	var rerr error
	go func() {
		defer pw.Close()
		defer wg.Done()
		rerr = run()
	}()

	w.log("parse")

	bset, err := parse.ParseSet(sout)
	if err != nil {
		return nil, err
	}

	w.log("wait")
	wg.Wait()

	if len(ds.Scripts.Post) != 0 {
		log.Printf("running post-run script for datastore %s: %s", ds.Name, w.replaceVars(ds.Scripts.Post))
		run := env.Cmd("/usr/bin/env", []string{"bash", "-c", "./postrun.sh " + w.replaceVars(ds.Scripts.Post)[1]}, os.Stderr, os.Stderr)
		if err := run(); err != nil {
			return nil, err
		}
	}

	if len(bset) != 1 {
		return nil, fmt.Errorf("unexpected bench count: %d", len(bset))
	}

	for _, b := range bset {
		if len(b) != 1 {
			return nil, errors.New("unexpected bench len")
		}

		return b[0], nil
	}

	panic("shouldn't be here")
}

// HACK!
func clone(in interface{}, out interface{}) error {
	b, _ := json.Marshal(in)
	return json.Unmarshal(b, out)
}
