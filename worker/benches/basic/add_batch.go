package basic

import (
	"context"
	"testing"

	"github.com/ipfs/go-ds-bench/options"
	"github.com/ipfs/go-ds-bench/worker/helpers"

	ds "github.com/ipfs/go-datastore"
)

func BenchAddBatch(b *testing.B, store ds.Batching, opt options.BenchOptions) {
	var keys []ds.Key
	var bufs [][]byte
	ctx := context.Background()

	for len(keys) < b.N {
		bufs = append(bufs, helpers.RandomBuf(opt.RecordSize))
		keys = append(keys, ds.RandomKey())
	}

	b.SetBytes(int64(opt.RecordSize))
	b.ResetTimer() // reset timer, this is start of real test

	batch, err := store.Batch(ctx)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		err := batch.Put(ctx, keys[i], bufs[i])
		if err != nil {
			b.Fatal(err)
		}

		if i%opt.BatchSize == opt.BatchSize-1 {
			err = batch.Commit(ctx)
			if err != nil {
				b.Fatal(err)
			}
			batch, err = store.Batch(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	err = batch.Commit(ctx)
	if err != nil {
		b.Fatal(err)
	}
}
