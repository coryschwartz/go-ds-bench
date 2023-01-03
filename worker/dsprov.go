package worker

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	ds "github.com/ipfs/go-datastore"
	// badgerds "github.com/ipfs/go-ds-badger"
	"github.com/ipfs/go-ds-bench/options"

	// boltds "github.com/ipfs/go-ds-bolt"
	flatfs "github.com/ipfs/go-ds-flatfs"
	leveldb "github.com/ipfs/go-ds-leveldb"
	levelopt "github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/mitchellh/go-homedir"
	// "gx/ipfs/QmdcULN1WCzgoQmcCaUAmEhwcxHYsDrbZ2LvRJKCL8dMrK/go-homedir"
)

func nopCloser() {}

type CandidateDatastore struct {
	Create  func() (func(fast bool) (ds.Batching, io.Closer, error), error)
	Destroy func()
}

var datastores = map[string]func(options.WorkerDatastore) CandidateDatastore{
	"memory-map": CandidateMemoryMap,
	"flatfs":     CandidateFlatfs,
	// "badger":     CandidateBadger,
	"leveldb": CandidateLeveldb,
	// "bolt":       CandidateBolt,
}

var CandidateMemoryMap = func(options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (func(bool) (ds.Batching, io.Closer, error), error) {
			mds := ds.NewMapDatastore()

			return func(fast bool) (ds.Batching, io.Closer, error) {
				return mds, mds, nil
			}, nil
		},
		Destroy: nopCloser,
	}
}

var CandidateFlatfs = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (func(bool) (ds.Batching, io.Closer, error), error) {
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

			return func(fast bool) (ds.Batching, io.Closer, error) {
				fs, err := flatfs.CreateOrOpen(dir, flatfs.NextToLast(2), !fast && spec.Params["Sync"].(bool))
				return fs, fs, err
			}, nil
		},
		Destroy: func() {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return
			}

			os.RemoveAll(d)
		},
	}
}

// var CandidateBadger = func(spec options.WorkerDatastore) CandidateDatastore {
// 	return CandidateDatastore{
// 		Create: func() (func(bool) (ds.Batching, io.Closer, error), error) {
// 			d, err := homedir.Expand(spec.Params["DataDir"].(string))
// 			if err != nil {
// 				return nil, err
// 			}

// 			err = os.MkdirAll(d, 0775)
// 			if err != nil {
// 				return nil, err
// 			}

// 			dir, err := ioutil.TempDir(d, "bench")
// 			if err != nil {
// 				return nil, err
// 			}

// 			err = os.MkdirAll(dir, 0775)
// 			if err != nil {
// 				return nil, err
// 			}

// 			return func(fast bool) (ds.Batching, io.Closer, error) {
// 				opts := badgerds.DefaultOptions
// 				opts.SyncWrites = !fast && spec.Params["Sync"].(bool)
// 				d, err := badgerds.NewDatastore(dir, &opts)

// 				return d, d, err
// 			}, nil
// 		},
// 		Destroy: func() {
// 			d, err := homedir.Expand(spec.Params["DataDir"].(string))
// 			if err != nil {
// 				return
// 			}

// 			os.RemoveAll(d)
// 		},
// 	}
// }

var CandidateLeveldb = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (func(bool) (ds.Batching, io.Closer, error), error) {
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

			return func(fast bool) (ds.Batching, io.Closer, error) {
				opts := leveldb.Options{
					Compression: levelopt.DefaultCompression,
					NoSync:      !fast && !spec.Params["Sync"].(bool),
				}

				ldb, err := leveldb.NewDatastore(dir, &opts)
				return ldb, ldb, err
			}, nil
		},
		Destroy: func() {
			d, err := homedir.Expand(spec.Params["DataDir"].(string))
			if err != nil {
				return
			}

			os.RemoveAll(d)
		},
	}
}

// var CandidateBolt = func(spec options.WorkerDatastore) CandidateDatastore {
// 	return CandidateDatastore{
// 		Create: func() (func(bool) (ds.Batching, io.Closer, error), error) {
// 			d, err := homedir.Expand(spec.Params["DataDir"].(string))
// 			if err != nil {
// 				return nil, err
// 			}

// 			err = os.MkdirAll(d, 0775)
// 			if err != nil {
// 				return nil, err
// 			}

// 			dir, err := ioutil.TempDir(d, "bench")
// 			if err != nil {
// 				return nil, err
// 			}

// 			err = os.MkdirAll(dir, 0775)
// 			if err != nil {
// 				return nil, err
// 			}

// 			return func(fast bool) (ds.Batching, io.Closer, error) {
// 				d, err := boltds.NewBoltDatastore(dir, "test", fast || !spec.Params["Sync"].(bool))
// 				return d, d, err
// 			}, nil
// 		},
// 		Destroy: func() {
// 			d, err := homedir.Expand(spec.Params["DataDir"].(string))
// 			if err != nil {
// 				return
// 			}

// 			os.RemoveAll(d)
// 		},
// 	}
// }

var CandidateDs = func(spec options.WorkerDatastore) CandidateDatastore {
	return CandidateDatastore{
		Create: func() (func(bool) (ds.Batching, io.Closer, error), error) {
			d, ok := datastores[spec.Type]
			if !ok {
				return nil, fmt.Errorf("unknows ds: '%s'", spec.Type)
			}

			construct, err := d(spec).Create()
			if err != nil {
				return nil, err
			}

			return func(fast bool) (ds.Batching, io.Closer, error) {
				return construct(fast)
			}, nil
		},
		Destroy: func() {
			datastores[spec.Type](spec).Destroy()
		},
	}
}
