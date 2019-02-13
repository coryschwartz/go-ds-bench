## go-datastore benchmarks

[WIP, WIP, WIP, WIP]

#### Running

First, build the worker:
```
make worker.test
upx --brute worker.test # optionally pack the worker binary
```

Create `systems.json` in run dir (here), Example:
```json
{
  "c5d.large": [
    {"Type": "ssh", "Spec": {"KeyFile": "/home/user/.ssh/bench-ec2.pem", "User": "ubuntu", "Addr": "10.1.2.3:22", "Vars": {"BDEV": "/dev/nvme0n1p1", "MDIR": "/mnt0"}}},
    {"Type": "ssh", "Spec": {"KeyFile": "/home/user/.ssh/bench-ec2.pem", "User": "ubuntu", "Addr": "10.1.2.4:22", "Vars": {"BDEV": "/dev/nvme0n1p1", "MDIR": "/mnt1"}}},
    {"Type": "ssh", "Spec": {"KeyFile": "/home/user/.ssh/bench-ec2.pem", "User": "ubuntu", "Addr": "10.1.2.5:22", "Vars": {"BDEV": "/dev/nvme0n1p1", "MDIR": "/mnt2"}}},
  ]
}
```

Open `master.go`, have a look at `newSpec` optionally adjusting it to run
specific benchmarks. If it looks good, run `go run master.go -continue` and
hope that it does it's thing.

Make sure to install filesystem tools on workers:
```bash
sudo apt install e2fsprogs btrfs-progs jfsutils xfsprogs ntfs-3g f2fs-tools
```

#### "Architecture"

Top level directory contains a few important parts:
* `worker/` - test executor - entrypoint is in `worker_test.go:BenchmarkSpec`
* `master/` - the thing that runs things (responsible for setting up workers (creating filesystems/crashing kernels), assigning jobs, collecting results, creating plots, not crashing irrecoverably after running benchmarks for days/weeks/months, etc.)
* `options/` - some shared code
* `scripts/` - shell scripts that are ran on workers before/after tests
* `master.go` - benchmark matrix 'config'
* `systems.json` - worker definition
  * map[systemType][]system
  * each systemType runs whole benchmark matrix, jobs are distributed across `system`s which are assumed to have the same hardware
* `results.json` - usually contains partial results before master crashes
* `plots/` - contains plots generated after running all benchmarks

##### Master

The most important struct in master is the `master.BenchSpec`:

```go
type master.BatchSpec struct { // results.json
	Datastores []options.WorkerDatastore // datastores x filesystems
	// instance type -> worker id
	Workers map[string][]Worker // systems.json

	// instance type -> series id
	Jobs map[string][]*Series
}
```

`master.Series` is another important struct which defines which test and with what params should be ran, and stores the results
```go
type master.Series struct {
	Opts     []options.BenchOptions
	Test     string // defined in Worker/worker_test.go
	PlotName string

	// ds -> Opts
	Results map[string]map[int]*parse.Benchmark

	lk sync.Mutex
}
```
