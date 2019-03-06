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

	nspec, err := new()
	if err != nil {
		return &BatchSpec{}, err
	}

	return nspec, nspec.save()
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
				fmt.Printf("\x1b[32m~~~~~~~~ %d JOBS LEFT ~~~~~~~~\x1b[39m\n", len(todo))
			}
		}
	}()

	return struct {
		in  chan<- workUnit
		out <-chan workUnit
	}{in: in, out: out}
}

type result struct {
	b   *parse.Benchmark
	err error

	instanceType string
	wu           workUnit
}

func (b *BatchSpec) save() error {
	// take all series locks
	for _, ij := range b.Jobs {
		for _, dj := range ij {
			dj.lk.Lock()
		}
	}

	defer func() {
		for _, ij := range b.Jobs {
			for _, dj := range ij {
				dj.lk.Unlock()
			}
		}
	}()

	m, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		panic(err)
	}

	//TODO: backups in case of borkage
	return ioutil.WriteFile("results.json", m, 0664)
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

	results := make(chan result)
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
						b, err := worker.run(b.Datastores[wu.ds], b.Jobs[itype][wu.series], wu.point)
						results <- result{
							b:   b,
							err: err,

							instanceType: itype,
							wu:           wu,
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

	go func() {
		wg.Wait()
		close(results)
	}()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				log.Printf("Stopping result collection (results chan closed)")
				return b.standardPlots()
			}

			if result.err != nil {
				// TODO: handle gracefully(-ier)
				log.Printf("WORKER ERROR: %s", result.err)
			}

			series := b.Jobs[result.instanceType][result.wu.series]
			series.lk.Lock()
			series.Results[b.Datastores[result.wu.ds].Name][result.wu.point] = result.b

			series.lk.Unlock()
			if err := b.save(); err != nil {
				return err
			}

		case <-ctx.Done():
			log.Printf("Stopping result collection (ctx expired)")
			return b.standardPlots()
		}
	}

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

		run := env.Cmd("/usr/bin/env", []string{"bash", "-c", "./prerun.sh " + w.replaceVars(ds.Scripts.Pre)[1]}, os.Stdout, os.Stdout)
		if err := run(); err != nil {
			return nil, err
		}
	}

	args := []string{"-test.benchmem", "-test.bench", "BenchmarkSpec"}

	pr, pw := io.Pipe()
	run := env.Cmd("./worker.test", args, pw, os.Stdout)
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
		run := env.Cmd("/usr/bin/env", []string{"bash", "-c", "./postrun.sh " + w.replaceVars(ds.Scripts.Post)[1]}, os.Stdout, os.Stdout)
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

func (b *BatchSpec) standardPlots() error {
	os.Mkdir("x_plots", 0755)

	for itype, srs := range b.Jobs {
		os.Mkdir("x_plots/"+itype, 0755)

		for _, s := range srs {
			if err := benchPlots(s.PlotName, "x_plots/"+itype+"/", s.Opts, s.Results); err != nil {
				return err
			}
		}
	}

	for itype, inst := range b.Jobs {

		for _, series := range inst {
			// tag -> ds_name
			tagged := map[string]map[string]map[int]*parse.Benchmark{}
			// type -> ds_name
			typed := map[string]map[string]map[int]*parse.Benchmark{}

			for _, ds := range b.Datastores {
				if _, ok := typed[ds.Type]; !ok {
					typed[ds.Type] = map[string]map[int]*parse.Benchmark{}
				}
				typed[ds.Type][ds.Name] = series.Results[ds.Name]

				for _, tag := range ds.Tags {
					if _, ok := tagged[tag]; !ok {
						tagged[tag] = map[string]map[int]*parse.Benchmark{}
					}
					tagged[tag][ds.Name] = series.Results[ds.Name]
				}
			}

			for t, res := range tagged {
				os.Mkdir("x_plots/"+itype+"/tag-"+t, 0755)

				if err := benchPlots(series.PlotName, "x_plots/"+itype+"/tag-"+t+"/", series.Opts, res); err != nil {
					return err
				}
			}

			for t, res := range typed {
				os.Mkdir("x_plots/"+itype+"/ds-"+t, 0755)

				if err := benchPlots(series.PlotName, "x_plots/"+itype+"/ds-"+t+"/", series.Opts, res); err != nil {
					return err
				}
			}

			os.Mkdir("x_plots/"+itype+"/tag--avg/", 0755)
			if err := benchPlots(series.PlotName, "x_plots/"+itype+"/tag--avg/", series.Opts, series.doAvg(tagged)); err != nil {
				return err
			}

			os.Mkdir("x_plots/"+itype+"/ds--avg/", 0755)
			if err := benchPlots(series.PlotName, "x_plots/"+itype+"/ds--avg/", series.Opts, series.doAvg(typed)); err != nil {
				return err
			}
		}
	}

	plotWg.Wait()

	return nil
}

// HACK!
func clone(in interface{}, out interface{}) error {
	b, _ := json.Marshal(in)
	return json.Unmarshal(b, out)
}
