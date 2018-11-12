package worker

import (
	"fmt"
	"github.com/ipfs/go-ds-bench/options"
	"gx/ipfs/QmdcULN1WCzgoQmcCaUAmEhwcxHYsDrbZ2LvRJKCL8dMrK/go-homedir"
	"io/ioutil"
	"os"

	ds "github.com/ipfs/go-datastore"
	flatfs "github.com/ipfs/go-ds-flatfs"
)

func nopCloser(_ ds.Batching) {}

type CandidateDatastore struct {
	Create  func() (ds.Batching, error)
	Destroy func(ds.Batching)
}

var datastores = map[string]func(options.WorkerDatastore)CandidateDatastore{
	"memory-map": CandidateMemoryMap,
	"flatfs": CandidateFlatfs,
}

var CandidateMemoryMap = func(options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (ds.Batching, error) {
			return ds.NewMapDatastore(), nil
		},
		Destroy: nopCloser,
	}
}

var CandidateFlatfs = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create:  func() (ds.Batching, error) {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return nil, err
			}

			os.Mkdir(d, 0775)

			dir, err := ioutil.TempDir(d, "bench")
			if err != nil {
				return nil, err
			}

			err = os.MkdirAll(dir, 0775)
			if err != nil {
				return nil, err
			}

			return flatfs.CreateOrOpen(dir, flatfs.NextToLast(2), spec.Params["Sync"].(bool))
		},
		Destroy: func(ds ds.Batching) {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return
			}

			err = os.RemoveAll(d)
			if err != nil {
				panic(err)
			}
		},
	}
}

var CandidateDs = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (ds.Batching, error) {
			ds, ok := datastores[spec.Type]
			if !ok {
				return nil, fmt.Errorf("unknows ds: '%s'", spec.Type)
			}

			return ds(spec).Create()
		},
		Destroy: func(d ds.Batching) {
			datastores[spec.Type](spec).Destroy(d)
		},
	}
}