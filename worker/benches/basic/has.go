package basic

import (
	"context"
	"testing"

	"github.com/ipfs/go-ds-bench/options"
	"github.com/ipfs/go-ds-bench/worker/helpers"
	"github.com/remeh/sizedwaitgroup"

	ds "github.com/ipfs/go-datastore"
)

func BenchHas(b *testing.B, store ds.Batching, opt options.BenchOptions) {
	ctx := context.Background()

	n := b.N
	if n > opt.PrimeRecordCount/5 {
		n = opt.PrimeRecordCount / 5
	}

	buf := make([]byte, opt.RecordSize)
	keys := make([]ds.Key, n)
	swg := sizedwaitgroup.New(256)

	for i := 0; i < n; i++ {
		buf = helpers.RandomBuf(opt.RecordSize)
		keys[i] = ds.RandomKey()

		if i%2 == 0 {
			swg.Add()
			go func(i int) {
				defer swg.Done()
				store.Put(ctx, keys[i], buf)
			}(i)
		}
	}
	swg.Wait()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.Has(ctx, keys[i%n])
		if err != nil {
			b.Fatal(err)
		}
	}
}
