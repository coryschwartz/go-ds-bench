package basic

import (
	"context"
	"testing"

	"github.com/ipfs/go-ds-bench/options"
	"github.com/ipfs/go-ds-bench/worker/helpers"

	ds "github.com/ipfs/go-datastore"
)

func BenchAdd(b *testing.B, store ds.Batching, opt options.BenchOptions) {
	var keys []ds.Key
	var bufs [][]byte
	ctx := context.Background()
	for len(keys) < b.N {
		bufs = append(bufs, helpers.RandomBuf(opt.RecordSize))
		keys = append(keys, ds.RandomKey())
	}

	b.SetBytes(int64(opt.RecordSize))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := store.Put(ctx, keys[i], bufs[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}
