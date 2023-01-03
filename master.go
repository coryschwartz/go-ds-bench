package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ipfs/go-ds-bench/master"
	"github.com/ipfs/go-ds-bench/options"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cont := flag.Bool("continue", false, "Continue previous work")
	flag.Parse()

	if flag.NArg() != 0 {
		fmt.Println("Usage: master [-continue]")
		return
	}

	b, err := master.BuildBatch(newSpec, *cont)
	if err != nil {
		panic(err)
	}

	if err := b.Start(); err != nil {
		panic(err)
	}
}

func newSpec() (*master.BatchSpec, error) {
	// Job matrix - systems x filesystems x datastores x series
	// datastores := []string{"flatfs", "badger", "leveldb", "bolt"}
	datastores := []string{"flatfs", "leveldb", "rados"}
	// filesystems := []fs{fsBtrfs, fsExt4, fsExt3, fsNtfs, fsJfs, fsXfs}
	filesystems := []fs{nofs}
	jobs := func() []*master.Series {
		return []*master.Series{
			master.BenchBasicAddBatch(),
			master.BenchBasicAdd(),
			master.BenchBasicGet(),
			master.BenchBasicHas(),

			master.BenchBSizingAddBatch(),
			master.BenchBSizingAdd(),
			master.BenchBSizingGet(),
			master.BenchBSizingHas(),

			master.BenchAddBatch(),
		}
	}

	// Load systems
	systems := map[string][]master.Worker{}

	f, err := os.Open("systems.json")
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(f).Decode(&systems)
	if err != nil {
		return nil, err
	}

	s := &master.BatchSpec{
		Workers: systems,

		Jobs: map[string][]*master.Series{},
	}

	// Populate jobs / Datastores
	for system := range s.Workers {
		s.Jobs[system] = jobs()
	}

	for _, ds := range datastores {
		for _, fs := range filesystems {
			s.Datastores = append(s.Datastores, options.WorkerDatastore{
				Type:    ds,
				Name:    fmt.Sprintf("%s-%s", ds, fs.Name),
				Scripts: fs.Scripts,
				Tags:    []string{fs.Name},
				Params: map[string]interface{}{
					"Sync":    true,
					"DataDir": "MDIR",
				},
			})
		}
	}

	return s, nil
}

type fs struct {
	Name    string
	Scripts struct {
		Pre  []string
		Post []string
	}
}

var nofs = fs{
	Name: "none",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/nothing.sh", "wtf"},
		Post: []string{"scripts/nothing.sh", "wtf"},
	},
}

var fsExt4 = fs{
	Name: "ext4",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/fs_prerun.sh", "ext4 BDEV MDIR"},
		Post: []string{"scripts/fs_postrun.sh", "MDIR"},
	},
}

var fsExt3 = fs{
	Name: "ext3",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/fs_prerun.sh", "ext3 BDEV MDIR"},
		Post: []string{"scripts/fs_postrun.sh", "MDIR"},
	},
}

var fsXfs = fs{
	Name: "xfs ",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/fs_prerun.sh", "xfs BDEV MDIR"},
		Post: []string{"scripts/fs_postrun.sh", "MDIR"},
	},
}

var fsBtrfs = fs{
	Name: "btrfs",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/fs_prerun.sh", "btrfs BDEV MDIR"},
		Post: []string{"scripts/fs_postrun.sh", "MDIR"},
	},
}

var fsJfs = fs{
	Name: "jfs",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/fs_prerun.sh", "'jfs -q' BDEV MDIR"},
		Post: []string{"scripts/fs_postrun.sh", "MDIR"},
	},
}

var fsNtfs = fs{
	Name: "ntfs",
	Scripts: struct {
		Pre  []string
		Post []string
	}{
		Pre:  []string{"scripts/fs_prerun.sh", "'ntfs -f' BDEV MDIR"},
		Post: []string{"scripts/fs_postrun.sh", "MDIR"},
	},
}
