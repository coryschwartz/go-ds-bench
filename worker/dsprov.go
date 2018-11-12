package worker

import (
	"fmt"
	ds "github.com/ipfs/go-datastore"
	badgerds "github.com/ipfs/go-ds-badger"
	"github.com/ipfs/go-ds-bench/options"
	"github.com/ipfs/go-ds-flatfs"
	"github.com/ipfs/go-ds-leveldb"
	levelopt "github.com/syndtr/goleveldb/leveldb/opt"
	"gx/ipfs/QmdcULN1WCzgoQmcCaUAmEhwcxHYsDrbZ2LvRJKCL8dMrK/go-homedir"
	"io/ioutil"
	"os"
)

func nopCloser(_ ds.Batching) {}

type CandidateDatastore struct {
	Create  func() (ds.Batching, error)
	Destroy func(ds.Batching)
}

var datastores = map[string]func(options.WorkerDatastore) CandidateDatastore{
	"memory-map": CandidateMemoryMap,
	"flatfs":     CandidateFlatfs,
	"badger":     CandidateBadger,
	"leveldb":    CandidateLeveldb,
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
		Create: func() (ds.Batching, error) {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return nil, err
			}

			err = os.MkdirAll(d, 0775)
			if err != nil {
				return nil, err
			}

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

			os.RemoveAll(d)
		},
	}
}

var CandidateBadger = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (ds.Batching, error) {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return nil, err
			}

			err = os.MkdirAll(d, 0775)
			if err != nil {
				return nil, err
			}

			dir, err := ioutil.TempDir(d, "bench")
			if err != nil {
				return nil, err
			}

			err = os.MkdirAll(dir, 0775)
			if err != nil {
				return nil, err
			}

			opts := badgerds.DefaultOptions
			opts.SyncWrites = spec.Params["Sync"].(bool)

			return badgerds.NewDatastore(dir, &opts)
		},
		Destroy: func(ds ds.Batching) {
			ds.(*badgerds.Datastore).Close()

			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return
			}

			os.RemoveAll(d)
		},
	}
}

var CandidateLeveldb = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (ds.Batching, error) {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return nil, err
			}

			err = os.MkdirAll(d, 0775)
			if err != nil {
				return nil, err
			}

			dir, err := ioutil.TempDir(d, "bench")
			if err != nil {
				return nil, err
			}

			err = os.MkdirAll(dir, 0775)
			if err != nil {
				return nil, err
			}

			opts := leveldb.Options{
				Compression: levelopt.DefaultCompression,
				NoSync:      !spec.Params["Sync"].(bool),
			}

			return leveldb.NewDatastore(dir, &opts)
		},
		Destroy: func(ds ds.Batching) {
			ds.(*leveldb.Datastore).Close()

			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return
			}

			os.RemoveAll(d)
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
